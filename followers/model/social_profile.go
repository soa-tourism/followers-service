package model

import (
	"encoding/json"
	"io"
)

type SocialProfile struct {
	UserID   int64  `json:"userId,omitempty"`
	Username string `json:"username,omitempty"`
}

type SocialProfiles []*SocialProfile

func (sp *SocialProfiles) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(sp)
}

func (sp *SocialProfile) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(sp)
}
