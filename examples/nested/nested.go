package nested

import "github.com/bobcob7/sudo-gen/examples/nested/duration"

type Job struct {
	Title    string              `json:"title,omitempty"`
	Company  string              `json:"company,omitempty"`
	Location string              `json:"location,omitempty"`
	Tenure   *duration.Timestamp `json:"tenure,omitempty"`
}

type Home struct {
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
	Age     duration.Duration
}
