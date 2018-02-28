package main

import "time"

// FreshEnergyToken contains data about API token.
type FreshEnergyToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresDate time.Time `json:"expires_date"`
	Scope       string    `json:"scope"`
	Jti         string    `json:"jti"`
}

// FreshEnergyAPI contains API response.
type FreshEnergyAPI struct {
	Readings []struct {
		DateTime      time.Time `json:"dateTime"`
		Power         float64   `json:"power"`
		PowerPhase1   float64   `json:"powerPhase1"`
		PowerPhase2   float64   `json:"powerPhase2"`
		PowerPhase3   float64   `json:"powerPhase3"`
		EnergyReading float64   `json:"energyReading"`
	} `json:"readings"`
	Links struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

// FreshEnergyClient contains client data for accessing API.
type FreshEnergyClient struct {
	Username string
	Password string
	Customer string
	Token    FreshEnergyToken
	API      FreshEnergyAPI
}

// queryTimeFormat contains time format used in API query.
var queryTimeFormat = "2006-01-02_15-04-05"
