package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// transport includes multiple timeouts for reliable http clients.
var transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

// client contains default client for http requests with timeouts.
var client = &http.Client{
	Timeout:   time.Second * 10,
	Transport: transport,
}

// NewFreshEnergyClient creates new client.
func NewFreshEnergyClient(username, password string, customer string) *FreshEnergyClient {
	c := &FreshEnergyClient{
		Username: username,
		Password: password,
		Customer: customer,
		Token:    FreshEnergyToken{},
		API:      FreshEnergyAPI{},
	}
	// try to load saved state
	if !useNewState {
		c.LoadState()
	}

	return c
}

// LoadState tries to load saved client state.
func (c *FreshEnergyClient) LoadState() {
	// read in state file
	content, err := ioutil.ReadFile(workDir + stateFile)
	if err != nil {
		return
	}

	// unmarshal to FreshEnergyClient struct
	var tmp FreshEnergyClient
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return
	}

	// get some parts from loaded state
	c.API = tmp.API
	c.Token = tmp.Token
}

// PersistState saves current client state.
func (c *FreshEnergyClient) PersistState() error {
	// persist some parts of client
	s := &FreshEnergyClient{
		Username: c.Username,
		Customer: c.Customer,
		Token:    c.Token,
		API: FreshEnergyAPI{
			Links: c.API.Links,
		},
	}

	// marshal to json
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// write to state file
	return ioutil.WriteFile(workDir+stateFile, b, 0640)
}

// Get queries API for energy data.
func (c *FreshEnergyClient) Get() error {
	// get token if current is invalid
	if c.Token.AccessToken == "" || time.Now().Unix() > c.Token.ExpiresDate.Unix() {
		err := c.getToken()
		if err != nil {
			return err
		}
	}

	// get next link
	var next string
	if c.API.Links.Next.Href != "" {
		// have next link in current state
		next = c.API.Links.Next.Href
	} else {
		// don't know where to start, start at t - 1h
		t := time.Now().UTC().Add(time.Hour * -1).Format(queryTimeFormat)
		next = fmt.Sprintf("%s/users/%s/readings/%s", energyAPI, c.Customer, t)
	}

	// build up query URL
	URL, err := url.Parse(next)
	if err != nil {
		return err
	}
	// add access token as value
	values := url.Values{}
	values.Add("access_token", c.Token.AccessToken)
	URL.RawQuery = values.Encode()

	// build up request
	req, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return err
	}
	// we want to get json formatted response from API
	req.Header.Set("Accept", "application/json")

	// do request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// suck it all in
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		// maybe we have a wrong access token or link, remove state and hope for the best
		c.API = FreshEnergyAPI{}
		c.Token = FreshEnergyToken{}
		return fmt.Errorf("could not get data from api, got http status code different than 200: %d", resp.StatusCode)
	}

	// unmatshal body to struct
	return json.Unmarshal(content, &c.API)
}

// getToken retrievs a new access token from API.
func (c *FreshEnergyClient) getToken() error {
	// build up query URL
	URL, err := url.Parse(energyAPI + "/oauth/token")
	if err != nil {
		return err
	}
	// build up request body, see oauth standard for more information
	data := url.Values{}
	data.Add("grant_type", "password")
	data.Add("username", c.Username)
	data.Add("password", c.Password)

	// build up http post request
	req, err := http.NewRequest("POST", URL.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	// set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	// use http basic auth with general username and empty password
	req.SetBasicAuth("fresh-webclient", "")

	// do request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// suck it all in
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		// something went wrong, maybe internal server error
		// hope that some humans read this error and fix it
		return fmt.Errorf("could not get new token, got http status code different than 200: %d", resp.StatusCode)
	}

	// unmarshal body to FreshEnergyToken struct
	var t FreshEnergyToken
	err = json.Unmarshal(content, &t)
	if err != nil {
		return err
	}
	// save token to client
	c.Token = t
	// calculate and set expiry date of token
	c.Token.ExpiresDate = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)

	return nil
}
