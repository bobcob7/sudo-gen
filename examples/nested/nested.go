package nested

//go:generate go run ../../../sudo-gen layerbroker -tests -json
type Config struct {
	Name string `json:"name,omitempty"`
	Jobs []Job  `json:"jobs,omitempty"`
	Home Home   `json:"home,omitempty"`
}

type Job struct {
	Title    string `json:"title,omitempty"`
	Company  string `json:"company,omitempty"`
	Location string `json:"location,omitempty"`
}

type Home struct {
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}
