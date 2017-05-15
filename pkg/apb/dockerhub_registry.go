package apb

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/containers/image/transports"
	"github.com/containers/image/types"
	logging "github.com/op/go-logging"
)

var ListImagesScript = "get_images_for_org.sh"

type DockerHubRegistry struct {
	config     RegistryConfig
	log        *logging.Logger
	ScriptsDir string
}

func (r *DockerHubRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("DockerHubRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

func (r *DockerHubRegistry) LoadSpecs() ([]*Spec, error) {
	r.log.Debug("DockerHubRegistry::LoadSpecs")
	var err error
	var rawBundleData []*ImageData
	var specs []*Spec

	if rawBundleData, err = r.loadBundleImageData(r.config.Org); err != nil {
		return nil, err
	}

	r.log.Debug("Raw image bundle size: %d", len(rawBundleData))
	if specs, err = r.createSpecs(rawBundleData); err != nil {
		return nil, err
	}

	////////////////////////////////////////////////////////////
	// TODO: DEBUG Remove dump
	////////////////////////////////////////////////////////////
	specsLogDump(specs, r.log)
	////////////////////////////////////////////////////////////

	return specs, nil
}

func (r *DockerHubRegistry) createSpecs(rawBundleData []*ImageData) ([]*Spec, error) {
	var err error
	var spec *Spec

	datToSpec := func(dat *ImageData) (*Spec, error) {
		var _err error
		_spec := &Spec{}

		encodedSpec := dat.Labels[BundleSpecLabel]
		if encodedSpec == "" {
			msg := fmt.Sprintf("Spec label not found on image -> [ %s ]", dat.Name)
			return nil, errors.New(msg)
		}

		decodedSpecYaml, _err := b64.StdEncoding.DecodeString(encodedSpec)
		if _err != nil {
			r.log.Error("Something went wrong deciding spec from label")
			return nil, _err
		}

		if _err = LoadYAML(string(decodedSpecYaml), _spec); _err != nil {
			r.log.Error("Something went wrong loading decoded spec yaml - %v - %v", string(decodedSpecYaml), _err)
			return nil, _err
		}

		return _spec, nil
	}

	specs := []*Spec{}
	for _, dat := range rawBundleData {
		if spec, err = datToSpec(dat); err != nil {
			r.log.Errorf("Unable to create spec", err)
		} else {
			specs = append(specs, spec)
		}
	}

	return specs, nil
}

func (r *DockerHubRegistry) loadBundleImageData(org string) ([]*ImageData, error) {
	r.log.Debug("DockerHubRegistry::loadBundleImageData")
	r.log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	r.log.Debug("Loading image list for org: [ %s ]", org)

	orgScript := path.Join(r.ScriptsDir, ListImagesScript)
	output, err := runCommand(
		"bash", orgScript, org, r.config.User, r.config.Pass)
	if err != nil {
		return nil, err
	}

	imageNames := strings.Split(string(output), "\n")
	imageNames = imageNames[:len(imageNames)-1]

	channel := make(chan *ImageData)

	for _, imageName := range imageNames {
		r.log.Debug("Trying to load " + imageName)
		go r.loadImageData("docker://"+imageName, channel)
	}

	var apbData []*ImageData
	counter := 1
	for imageData := range channel {
		if imageData.Error != nil {
			r.log.Error("Something went wrong loading img data for [ %s ]", imageData.Name)
			r.log.Error(fmt.Sprintf("Error: %s", imageData.Error))
			counter++
			continue
		}

		if imageData.IsPlaybookBundle {
			r.log.Notice("We have a playbook bundle, adding its imagedata")
			apbData = append(apbData, imageData)
		} else {
			r.log.Notice("We did NOT add the imageData for some reason")
		}

		if counter != len(imageNames) {
			counter++
		} else {
			close(channel)
		}
	}

	r.log.Info("Found apbs:")
	for _, dat := range apbData {
		r.log.Info(fmt.Sprintf("%s", dat.Name))
	}

	return apbData, nil
}

func (r *DockerHubRegistry) loadImageData(imageName string, channel chan<- *ImageData) {
	// TODO: Error handling!
	img, err := parseImage(imageName)
	if err != nil {
		channel <- &ImageData{Name: imageName, Error: err}
		return
	}
	defer img.Close()

	imgInspect, err := img.Inspect()
	if err != nil {
		channel <- &ImageData{Error: err}
		return
	}

	outputData := ImageData{
		Name:             imageName,
		Tag:              imgInspect.Tag,
		Labels:           imgInspect.Labels,
		Layers:           imgInspect.Layers,
		IsPlaybookBundle: true,
		Error:            nil,
	}

	if outputData.Labels[BundleSpecLabel] != "" {
		outputData.IsPlaybookBundle = true
		channel <- &outputData
	} else {
		outputData.IsPlaybookBundle = false
		channel <- &outputData
	}
}

func parseImage(imgName string) (types.Image, error) {
	ref, err := transports.ParseImageName(imgName)
	if err != nil {
		return nil, err
	}
	return ref.NewImage(contextFromGlobalOptions(false))
}

func contextFromGlobalOptions(tlsVerify bool) *types.SystemContext {
	return &types.SystemContext{
		RegistriesDirPath:           "",
		DockerCertPath:              "",
		DockerInsecureSkipTLSVerify: !tlsVerify,
	}
}
