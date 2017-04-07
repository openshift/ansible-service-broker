package apb

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/containers/image/transports"
	"github.com/containers/image/types"
	logging "github.com/op/go-logging"
)

// DockerHubRegistry - Docker Hub registry
type DockerHubRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

// Init - Initialize the docker hub registry
func (r *DockerHubRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("DockerHubRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// LoadSpecs - Will load the specs from the docker hub registry.
func (r *DockerHubRegistry) LoadSpecs() ([]*Spec, int, error) {
	r.log.Debug("DockerHubRegistry::LoadSpecs")
	var err error
	var rawBundleData []*ImageData
	var specs []*Spec

	if rawBundleData, err = r.loadBundleImageData(r.config.Org); err != nil {
		return nil, 0, err
	}

	r.log.Debug("Raw image bundle size: %d", len(rawBundleData))
	if specs, err = r.createSpecs(rawBundleData); err != nil {
		return nil, len(rawBundleData), err
	}

	////////////////////////////////////////////////////////////
	// TODO: DEBUG Remove dump
	////////////////////////////////////////////////////////////
	specsLogDump(specs, r.log)
	////////////////////////////////////////////////////////////

	return specs, len(rawBundleData), nil
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

		// Convert YAML to JSON.  We only want to traffic in JSON
		SpecJson, _err := ToJSON(decodedSpecYaml)
		if err != nil {
			return nil, _err
		}

		if _err = LoadJSON(string(SpecJson), _spec); _err != nil {
			r.log.Error("Something went wrong loading decoded spec yaml - %v - %v", string(decodedSpecYaml), _err)
			return nil, _err
		}

		return _spec, nil
	}

	var specs []*Spec
	for _, dat := range rawBundleData {
		if spec, err = datToSpec(dat); err != nil {
			r.log.Errorf("Unable to create spec - %v image: %v", err, dat.Name)
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

	type Payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type TokenResponse struct {
		Token string `json:"token"`
	}

	type Images struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}

	type ImageResponse struct {
		Count   int       `json:"count"`
		Results []*Images `json:"results"`
	}

	data := Payload{
		Username: r.config.User,
		Password: r.config.Pass,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonToken, err := ioutil.ReadAll(resp.Body)

	tokenResp := TokenResponse{}
	err = json.Unmarshal(jsonToken, &tokenResp)
	if err != nil {
		return nil, err
	}

	// TODO: Introduce an asyc search for APBs so that bootstrap will search
	// through the entire repo instead of being maxed at 100.
	// (Example) rhallisey has > 100 images in his repo and the limiting
	// page size prevents new APBs from appearing in the broker.
	req, err = http.NewRequest("GET", os.ExpandEnv(fmt.Sprintf("https://hub.docker.com/v2/repositories/%v/?page_size=100", org)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("JWT %v", string(tokenResp.Token))))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := ImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return nil, err
	}

	channel := make(chan *ImageData)

	for _, imageName := range imageResp.Results {
		r.log.Debug("Trying to load " + imageName.Namespace + "/" + imageName.Name)
		go r.loadImageData("docker://"+imageName.Namespace+"/"+imageName.Name, channel)
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

		if counter != len(imageResp.Results) {
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
