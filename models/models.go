package models

type OauthState struct {
	State    string `json:"state"`
	Code     string `json:"code"`
	Verifier string
}
