package nested

import "time"

//go:generate go run ../../../sudo-gen layerbroker -tests -json
type Config struct {
	Name      string    `json:"name,omitempty"`
	Jobs      []Job     `json:"jobs,omitempty"`
	Home      Home      `json:"home,omitempty"`
	OtherHome *Home     `json:"home,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
