package model

import (
	"encoding/json"
	"io"
)

type SocialProfileData struct {
	UserID    int64            `json:"userId,omitempty"`
	Username  string           `json:"username,omitempty"`
	Followers []*SocialProfile `json:"followers,omitempty"`
	Following []*SocialProfile `json:"following,omitempty"`
}

type SocialProfileDatas []*SocialProfileData

func (sp *SocialProfileDatas) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(sp)
}

func (sp *SocialProfileData) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(sp)
}
