package config

// APIConfig APIConfig
type APIConfig struct {
	LBHost  string   `json:"lbhost"`
	SvrAddr string   `json:"svraddr"`
	URIPath []string `json:"uripath"`
	// healthcheck url
	HCURL string `json:"hcurl"`
}
