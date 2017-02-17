package ansibleapp

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/containers/image/transports"
	"github.com/containers/image/types"
	"github.com/op/go-logging"
	"path"
	"strings"
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
	var rawAnsibleAppData []*ImageData
	var specs []*Spec

	if rawAnsibleAppData, err = r.loadAnsibleAppImageData(r.config.Org); err != nil {
		return nil, err
	}

	if specs, err = r.createSpecs(rawAnsibleAppData); err != nil {
		return nil, err
	}

	////////////////////////////////////////////////////////////
	// TODO: DEBUG Remove dump
	////////////////////////////////////////////////////////////
	specsLogDump(specs, r.log)
	////////////////////////////////////////////////////////////

	return specs, nil
}

func (r *DockerHubRegistry) createSpecs(
	rawAnsibleAppData []*ImageData,
) ([]*Spec, error) {
	var err error
	var spec *Spec

	datToSpec := func(dat *ImageData) (*Spec, error) {
		var _err error
		_spec := &Spec{}

		encodedSpec := dat.Labels[AnsibleAppSpecLabel]
		if encodedSpec == "" {
			msg := fmt.Sprintf("Spec label not found on image -> [ %s ]", dat.Name)
			return nil, errors.New(msg)
		}

		decodedSpecYaml, _err := b64.StdEncoding.DecodeString(encodedSpec)
		if _err != nil {
			return nil, err
		}

		if _err = LoadYAML(string(decodedSpecYaml), _spec); _err != nil {
			return nil, _err
		}

		return _spec, nil
	}

	specs := []*Spec{}
	for _, dat := range rawAnsibleAppData {
		if spec, err = datToSpec(dat); err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func (r *DockerHubRegistry) loadAnsibleAppImageData(
	org string,
) ([]*ImageData, error) {
	r.log.Debug("DockerHubRegistry::loadAnsibleAppImageData")
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

	var ansibleAppData []*ImageData
	counter := 1
	for imageData := range channel {
		if imageData.Error != nil {
			r.log.Error("Something went wrong loading img data for [ %s ]", imageData.Name)
			r.log.Error(fmt.Sprintf("Error: %s", imageData.Error))
			counter++
			continue
		}

		if imageData.IsAnsibleApp {
			ansibleAppData = append(ansibleAppData, imageData)
		}

		if counter != len(imageNames) {
			counter++
		} else {
			close(channel)
		}
	}

	r.log.Info("Found ansibleapps:")
	for _, dat := range ansibleAppData {
		r.log.Info(fmt.Sprintf("%s", dat.Name))
	}

	return ansibleAppData, nil
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
		Name:         imageName,
		Tag:          imgInspect.Tag,
		Labels:       imgInspect.Labels,
		Layers:       imgInspect.Layers,
		IsAnsibleApp: true,
		Error:        nil,
	}

	if outputData.Labels[AnsibleAppSpecLabel] != "" {
		outputData.IsAnsibleApp = true
		channel <- &outputData
	} else {
		outputData.IsAnsibleApp = false
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
