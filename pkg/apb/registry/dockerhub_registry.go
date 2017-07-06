package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// DockerHubRegistry - Docker Hub registry
type DockerHubRegistry struct {
	config Config
	log    *logging.Logger
}

// DockerHubImage - Image from a dockerhub registry.
type DockerHubImage struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DockerHubImageResponse - Image response for dockerhub.
type DockerHubImageResponse struct {
	Count   int               `json:"count"`
	Results []*DockerHubImage `json:"results"`
	Next    string            `json:"next"`
}

// DockerHubImageData - used to retrieve specs.
type DockerHubImageData struct {
	Name             string
	Spec             *apb.Spec
	Error            error
	IsPlaybookBundle bool
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
	var specs []*apb.Spec

	if specs, err = r.loadBundleImageData(r.config.Org); err != nil {
		return nil, 0, err
	}
	////////////////////////////////////////////////////////////
	// TODO: DEBUG Remove dump
	////////////////////////////////////////////////////////////
	apb.SpecsLogDump(specs, r.log)
	////////////////////////////////////////////////////////////
	return specs, len(specs), nil
}

// Fail - will determine if this reqistry can cause a failure.
func (r DockerHubRegistry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
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

func (r DockerHubRegistry) loadBundleImageData(org string) ([]*apb.Spec, error) {
	r.log.Debug("DockerHubRegistry::loadBundleImageData")
	r.log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	r.log.Debug("Loading image list for org: [ %s ]", org)

	token, err := r.getDockerHubToken()

	channel := make(chan *DockerHubImageData)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	//Intial call to getNextImages this will fan out to retrieve all the values.
	imageResp, err := r.getNextImages(ctx, org, token,
		fmt.Sprintf("https://hub.docker.com/v2/repositories/%v/?page_size=100", org),
		channel, cancelFunc)
	//if there was an issue with the first call, return the error
	if err != nil {
		return nil, err
	}
	//If no results in the fist call then close the channel as nothing will get loaded.
	if len(imageResp.Results) == 0 {
		r.log.Info("canceled retrieval as no items in org")
		close(channel)
	}
	var apbData []*apb.Spec
	counter := 1
	for imageData := range channel {
		if imageData.Error != nil {
			r.log.Errorf("Something went wrong loading img data name: %v -  %v",
				imageData.Name, imageData.Error)
		}

		if imageData.IsPlaybookBundle {
			r.log.Noticef("We have a playbook bundle, adding its imagedata")
			apbData = append(apbData, imageData.Spec)
		} else {
			r.log.Noticef("We did NOT add the imageData - %v for some reason",
				imageData.Name)
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

// getNextImages - will follow the next URL using go routines.
func (r DockerHubRegistry) getNextImages(ctx context.Context, org, token, url string, ch chan<- *DockerHubImageData,
	cancelFunc context.CancelFunc) (*DockerHubImageResponse, error) {
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

	iResp := DockerHubImageResponse{}
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
		go r.getNextImages(ctx, org, token, iResp.Next, ch, cancelFunc)
	}
	for _, imageName := range iResp.Results {
		r.log.Debugf("Trying to load %v/%v", imageName.Namespace, imageName.Name)
		go r.loadImageData(ctx,
			fmt.Sprintf("%v/%v", imageName.Namespace, imageName.Name),
			ch)
	}
	return &iResp, nil
}

func (r DockerHubRegistry) loadImageData(ctx context.Context, imageName string, channel chan<- *DockerHubImageData) {
	req, err := http.NewRequest("GET", fmt.Sprintf(
		"https://registry.hub.docker.com/v2/%v/manifests/latest", imageName), nil)
	if err != nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &DockerHubImageData{Error: err, Name: imageName}
		}
		return
	}
	token, err := getBearerToken(imageName)
	if err != nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &DockerHubImageData{Error: err, Name: imageName}
		}
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	spec, err := imageToSpec(r.log, req)
	if err != nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &DockerHubImageData{Error: err, Name: imageName}
		}
		return
	}
	if spec == nil {
		select {
		case <-ctx.Done():
			r.log.Debugf("loading images failed due to context err - %v name - %v", ctx.Err(), imageName)
			return
		default:
			channel <- &DockerHubImageData{Name: imageName, IsPlaybookBundle: false}
		}
		return
	}
	spec.RegistryName = r.config.Name
	select {
	case <-ctx.Done():
		r.log.Debugf("loading images failed due to context err - %v name - %v",
			ctx.Err(), imageName)
		return
	default:
		channel <- &DockerHubImageData{IsPlaybookBundle: true, Error: nil, Spec: spec}
	}
}

func getBearerToken(imageName string) (string, error) {
	response, err := http.Get(fmt.Sprintf(
		"https://auth.docker.io/token?service=registry.docker.io&scope=repository:%v:pull",
		imageName))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	t := struct {
		Token string `json:"token"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&t)
	if err != nil {
		return "", err
	}
	return t.Token, nil
}
