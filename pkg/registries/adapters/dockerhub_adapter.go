package adapters

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

const dockerhubName = "docker.io"
const dockerHubLoginURL = "https://hub.docker.com/v2/users/login/"
const dockerHubRepoImages = "https://hub.docker.com/v2/repositories/%v/?page_size=100"
const dockerHubManifestURL = "https://registry.hub.docker.com/v2/%v/manifests/latest"

// DockerHubAdapter - Docker Hub Adapter
type DockerHubAdapter struct {
	Config Configuration
	Log    *logging.Logger
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

// RegistryName - Retrieve the registry name
func (r DockerHubAdapter) RegistryName() string {
	return dockerhubName
}

// GetImages - retrieve the images
func (r DockerHubAdapter) GetImages() ([]string, error) {
	r.Log.Debug("DockerHubAdapter::GetImages")
	r.Log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	r.Log.Debug("Loading image list for org: [ %s ]", r.Config.Org)

	token, err := r.getDockerHubToken()

	channel := make(chan string)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	//Intial call to getNextImages this will fan out to retrieve all the values.
	imageResp, err := r.getNextImages(ctx, r.Config.Org, token,
		fmt.Sprintf(dockerHubRepoImages, r.Config.Org),
		channel, cancelFunc)
	//if there was an issue with the first call, return the error
	if err != nil {
		return nil, err
	}
	//If no results in the fist call then close the channel as nothing will get loaded.
	r.Log.Info("%#v", imageResp)
	if len(imageResp.Results) == 0 {
		r.Log.Info("canceled retrieval as no items in org")
		close(channel)
	}
	var apbData []string
	counter := 1
	for imageData := range channel {
		r.Log.Infof("%v", apbData)
		apbData = append(apbData, imageData)
		if counter < imageResp.Count {
			r.Log.Infof("%v", apbData)
			counter++
		} else {
			r.Log.Infof("%v", apbData)
			close(channel)
		}
	}
	// check to see if the context had an error
	if ctx.Err() != nil {
		r.Log.Error("encountered an error while loading images, we may not have all the apb in the catalog - %v", ctx.Err())
		return apbData, ctx.Err()
	}

	return apbData, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r DockerHubAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	specs := []*apb.Spec{}
	for _, imageName := range imageNames {
		spec, err := r.loadSpec(imageName)
		if err != nil {
			r.Log.Errorf("unable to retrieve spec data for image - %v", err)
			return specs, err
		}
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// getDockerHubToken - will retrieve the docker hub token.
func (r DockerHubAdapter) getDockerHubToken() (string, error) {
	type Payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type TokenResponse struct {
		Token string `json:"token"`
	}
	data := Payload{
		Username: r.Config.User,
		Password: r.Config.Pass,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", dockerHubLoginURL, body)
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

// getNextImages - will follow the next URL using go routines.
func (r DockerHubAdapter) getNextImages(ctx context.Context,
	org, token, url string,
	ch chan<- string,
	cancelFunc context.CancelFunc) (*DockerHubImageResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		r.Log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("JWT %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.Log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}
	defer resp.Body.Close()

	imageList, err := ioutil.ReadAll(resp.Body)

	iResp := DockerHubImageResponse{}
	err = json.Unmarshal(imageList, &iResp)
	if err != nil {
		r.Log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return &iResp, err
	}
	//Keep getting the images
	if iResp.Next != "" {
		r.Log.Debugf("getting next page of results - %v", iResp.Next)
		//Fan out calls to get the next images.
		go r.getNextImages(ctx, org, token, iResp.Next, ch, cancelFunc)
	}
	for _, imageName := range iResp.Results {
		r.Log.Debugf("Trying to load %v/%v", imageName.Namespace, imageName.Name)
		go func(image *DockerHubImage) {
			select {
			case <-ctx.Done():
				r.Log.Debugf(
					"loading images failed due to context err - %v name - %v",
					ctx.Err(), image.Name)
				return
			default:
				ch <- fmt.Sprintf("%v/%v", image.Namespace, image.Name)
			}
		}(imageName)
	}
	return &iResp, nil
}

func (r DockerHubAdapter) loadSpec(imageName string) (*apb.Spec, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(dockerHubManifestURL, imageName), nil)
	if err != nil {
		return nil, err
	}
	token, err := getBearerToken(imageName)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	return imageToSpec(r.Log, req)
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
