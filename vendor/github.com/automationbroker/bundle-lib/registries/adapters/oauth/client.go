//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package oauth

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// authResponse - holds the response data from an oauth2 token request
type authResponse struct {
	Token string `json:"access_token"`
}

// NewClient - creates and returns a *Client ready to use. If skipVerify is
// true, it will skip verification of the remote TLS certificate.
func NewClient(user, pass string, skipVerify bool, url *url.URL) *Client {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify}}
	if skipVerify == true {
		log.Warn("skipping verification of registry TLS certificate per adapter configuration")
	}

	return &Client{
		user:   user,
		pass:   pass,
		url:    url,
		mutex:  &sync.Mutex{},
		client: &http.Client{Timeout: time.Second * 60, Transport: transport},
	}
}

// Client - may be used as an HTTP client that automatically authenticates using oauth.
type Client struct {
	user   string
	pass   string
	token  string
	mutex  *sync.Mutex
	client *http.Client
	url    *url.URL
}

// NewRequest - creates and returns a *http.Request assuming the GET method.
// The base URL configured on the Client gets used with its Path component
// replaced by the path argument. If a token is available, it is added to the
// request automatically. "Accept: application/json" is added to all requests.
// The caller should customize the request as necessary before using it.
func (c *Client) NewRequest(path string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = path
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	req.Header.Add("Accept", "application/json")
	return req, nil
}

// Do - passes through to the underlying http.Client instance
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Getv2 - makes a GET request to the registry's /v2/ endpoint. If a 401
// Unauthorized response is received, this method attempts to obtain an oauth
// token and tries again with the new token. If a username and password are
// available, they are used with Basic Auth in the request to the token
// service. This method is goroutine-safe.
func (c *Client) Getv2() error {
	// lock to prevent multiple goroutines from retrieving, and especially
	// writing, a new token at the same time
	c.mutex.Lock()
	defer c.mutex.Unlock()
	req, err := c.NewRequest("/v2/")
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		h := resp.Header.Get("www-authenticate")
		c.getToken(h)

		// try the new token
		tokenReq, err := c.NewRequest("/v2/")
		if err != nil {
			return err
		}
		tokenResp, err := c.Do(tokenReq)
		if err != nil {
			return err
		}
		defer tokenResp.Body.Close()

		if tokenResp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("Token not accepted by /v2/ - %s", tokenResp.Status)
			log.Warn(msg)
			return errors.New(msg)
		}
		log.Debug("GET /v2/ successful with new token")

	case http.StatusOK:
		if c.token == "" {
			log.Debug("GET /v2/ successful without token")
		} else {
			log.Debug("GET /v2/ successful with existing token")
		}

	default:
		msg := fmt.Sprintf("Bad response from /v2/ - %s", resp.Status)
		log.Warn(msg)
		return errors.New(msg)
	}
	return nil
}

// getToken - parses a www-authenticate header and uses the information to
// retrieve an oauth token. The token is stored on the Client and automatically
// added to future requests.
func (c *Client) getToken(wwwauth string) error {
	// compute the URL
	u, err := parseAuthHeader(wwwauth)
	if err != nil {
		return err
	}

	// form the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Errorf("could not form request: %s", err.Error())
		return err
	}
	if c.pass != "" {
		log.Debug("adding basic auth to token request")
		req.SetBasicAuth(c.user, c.pass)
	}

	// make the request
	resp, err := c.client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("error obtaining token: %s", err.Error())
		log.Warn(msg)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("token service responded: %s", resp.Status)
		log.Warn(msg)
		return errors.New(msg)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("failed to read token body: %s", err.Error())
		return err
	}

	authResp := authResponse{}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return err
	}
	c.token = authResp.Token
	log.Debugf("new token: %s", c.token)
	return nil
}

// parseAuthHeader - parses the text from a www-authenticate header and uses it
// to construct a url.URL that can be used to retrieve a token.
func parseAuthHeader(value string) (*url.URL, error) {
	rrealm, err := regexp.Compile("realm=\"([^\"]+)\"")
	if err != nil {
		return nil, err
	}
	rservice, err := regexp.Compile("service=\"([^\"]+)\"")
	if err != nil {
		return nil, err
	}

	rmatch := rrealm.FindStringSubmatch(value)
	if len(rmatch) != 2 {
		msg := fmt.Sprintf("Could not parse www-authenticate header: %s", value)
		log.Warn(msg)
		return nil, errors.New(msg)
	}
	realm := rmatch[1]

	u, err := url.Parse(realm)
	if err != nil {
		msg := fmt.Sprintf("realm is not a valid URL: %s", realm)
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	smatch := rservice.FindStringSubmatch(value)
	if len(smatch) == 2 {
		service := smatch[1]
		q := u.Query()
		q.Set("service", service)
		u.RawQuery = q.Encode()
	}

	return u, nil
}
