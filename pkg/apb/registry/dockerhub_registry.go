package registry

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containers/image/transports"
	"github.com/containers/image/types"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	yaml "gopkg.in/yaml.v2"
)

// DockerHubRegistry - Docker Hub registry
type DockerHubRegistry struct {
	config Config
	log    *logging.Logger
}

// Init - Initialize the docker hub registry
func (r *DockerHubRegistry) Init(config Config, log *logging.Logger) error {
	log.Debug("DockerHubRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// LoadSpecs - Will load the specs from the docker hub registry.
func (r *DockerHubRegistry) LoadSpecs() ([]*apb.Spec, int, error) {
	r.log.Debug("DockerHubRegistry::LoadSpecs")
	var err error
	var rawBundleData []*ImageData
	var specs []*apb.Spec

	if rawBundleData, err = r.loadBundleImageData(r.config.Org); err != nil {
		return nil, 0, err
	}

	r.log.Debug("Raw image bundle size: %d image bundle -%v", len(rawBundleData), rawBundleData)
	if specs, err = r.createSpecs(rawBundleData); err != nil {
		return nil, len(rawBundleData), err
	}

	////////////////////////////////////////////////////////////
	// TODO: DEBUG Remove dump
	////////////////////////////////////////////////////////////
	apb.SpecsLogDump(specs, r.log)
	////////////////////////////////////////////////////////////

	return specs, len(rawBundleData), nil
}

// Fail - will determine if this reqistry can cause a failure.
func (r DockerHubRegistry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}

func (r *DockerHubRegistry) createSpecs(rawBundleData []*ImageData) ([]*apb.Spec, error) {
	var err error
	var spec *apb.Spec

	datToSpec := func(dat *ImageData) (*apb.Spec, error) {
		var _err error
		_spec := &apb.Spec{}

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

		if _err = yaml.Unmarshal(decodedSpecYaml, _spec); _err != nil {
			r.log.Error("Something went wrong loading decoded spec yaml - %v - %v", string(decodedSpecYaml), _err)
			return nil, _err
		}

		return _spec, nil
	}

	var specs []*apb.Spec
	for _, dat := range rawBundleData {
		if spec, err = datToSpec(dat); err != nil {
			r.log.Errorf("Unable to create spec - %v image: %v", err, dat.Name)
		} else {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// getDockerHubToken - will retrieve the docker hub token.
func (r DockerHubRegistry) getDockerHubToken() (string, error) {
	type Payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type TokenResponse struct {
		Token string `json:"token"`
	}
	data := Payload{
		Username: r.config.User,
		Password: r.config.Pass,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	jsonToken, err := ioutil.ReadAll(resp.Body)

	tokenResp := TokenResponse{}
	err = json.Unmarshal(jsonToken, &tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

func (r DockerHubRegistry) loadBundleImageData(org string) ([]*ImageData, error) {
	r.log.Debug("DockerHubRegistry::loadBundleImageData")
	r.log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	r.log.Debug("Loading image list for org: [ %s ]", org)

	type Images struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}

	type ImageResponse struct {
		Count   int       `json:"count"`
		Results []*Images `json:"results"`
		Next    string    `json:"next"`
	}

	token, err := r.getDockerHubToken()

	channel := make(chan *ImageData)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Will kick of go routines for each result in the images.
	loadImagesFromResult := func(ctx context.Context, images []*Images) {
		for _, imageName := range images {
			r.log.Debugf("Trying to load %v/%v", imageName.Namespace, imageName.Name)
			go r.loadImageData(ctx, "docker://"+imageName.Namespace+"/"+imageName.Name, channel)
		}
	}

	// getNextImages - will follow the next URL using go routines.
	// The workflow is query the url, if Next is defined then kickoff another go routine to get those images.
	// then kick of the loading of image data.
	var getNextImages func(ctx context.Context, org, token, url string, ch chan<- *ImageData, cancelFunc context.CancelFunc) (*ImageResponse, error)

	getNextImages = func(ctx context.Context, org, token, url string, ch chan<- *ImageData, cancelFunc context.CancelFunc) (*ImageResponse, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			r.log.Errorf("unable to get next images for url: %v - %v", url, err)
			cancelFunc()
			close(ch)
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("JWT %v", token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			r.log.Errorf("unable to get next images for url: %v - %v", url, err)
			cancelFunc()
			close(ch)
			return nil, err
		}
		defer resp.Body.Close()

		imageList, err := ioutil.ReadAll(resp.Body)

		iResp := ImageResponse{}
		err = json.Unmarshal(imageList, &iResp)
		if err != nil {
			r.log.Errorf("unable to get next images for url: %v - %v", url, err)
			cancelFunc()
			close(ch)
			return &iResp, err
		}
		//Keep getting the images
		if iResp.Next != "" {
			r.log.Debugf("getting next page of results - %v", iResp.Next)
			//Fan out calls to get the next images.
			go getNextImages(ctx, org, token, iResp.Next, channel, cancelFunc)
		}
		loadImagesFromResult(ctx, iResp.Results)
		return &iResp, nil
	}

	//Intial call to getNextImages this will fan out to retrieve all the values.
	imageResp, err := getNextImages(ctx, org, token, fmt.Sprintf("https://hub.docker.com/v2/repositories/%v/?page_size=100", org), channel, cancelFunc)
	//if there was an issue with the first call, return the error
	if err != nil {
		return nil, err
	}
	//If no results in the fist call then close the channel as nothing will get loaded.
	if len(imageResp.Results) == 0 {
		r.log.Info("canceled retrieval as no items in org")
		close(channel)
	}
	var apbData []*ImageData
	counter := 1
	for imageData := range channel {
		if imageData.Error != nil {
			r.log.Error("Something went wrong loading img data for [ %s ]", imageData.Name)
			r.log.Error(fmt.Sprintf("Error: %s", imageData.Error))
		}

		if imageData.IsPlaybookBundle {
			r.log.Notice("We have a playbook bundle, adding its imagedata")
			apbData = append(apbData, imageData)
		} else {
			r.log.Notice("We did NOT add the imageData for some reason")
		}

		if counter < imageResp.Count {
			counter++
		} else {
			close(channel)
		}
	}

	r.log.Info("Found apbs:")
	for _, dat := range apbData {
		r.log.Info(fmt.Sprintf("%s", dat.Name))
	}
	// check to see if the context had an error
	if ctx.Err() != nil {
		r.log.Error("encountered an error while loading images, we may not have all the apb in the catalog - %v", ctx.Err())
		return apbData, ctx.Err()
	}

	return apbData, nil
}

func (r DockerHubRegistry) loadImageData(ctx context.Context, imageName string, channel chan<- *ImageData) {
	// TODO: Error handling!
	img, err := parseImage(imageName)
	if err != nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &ImageData{Name: imageName, Error: err}
		}
		return
	}
	defer img.Close()

	imgInspect, err := img.Inspect()
	if err != nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &ImageData{Name: imageName, Error: err}
		}
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
	} else {
		outputData.IsPlaybookBundle = false
	}
	select {
	case <-ctx.Done():
		r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
		return
	default:
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
