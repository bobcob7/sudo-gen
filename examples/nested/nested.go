package nested

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
