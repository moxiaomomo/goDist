package config

// APIConfig APIConfig
type APIConfig struct {
	ServiceName string   `json:"servicename"`
	LBHost      string   `json:"lbhost"`
	SvrAddr     string   `json:"svraddr"`
	URIPath     []string `json:"uripath"`
	// healthcheck url
	HCURL string `json:"hcurl"`
}
