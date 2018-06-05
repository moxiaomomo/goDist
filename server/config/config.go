package config

// ServerConfig ServerConfig
type ServerConfig struct {
	LBHost  string   `json:"lbhost"`
	SvrAddr string   `json:"svraddr"`
	URIPath []string `json:"uripath"`
}
