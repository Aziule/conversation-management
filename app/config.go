package app

import (
	"encoding/json"
	"io/ioutil"
)

// Config is the struct that will hold the runtime configuration
type Config struct {
	Debug             bool   `json:"debug"`
	ListeningPort     int    `json:"listening_port"`
	FbVerifyToken     string `json:"fb_verify_token"`
	FbApiVersion      string `json:"fb_api_version"`
	FbPageAccessToken string `json:"fb_page_access_token"`
	DbHost            string `json:"db_host"`
	DbName            string `json:"db_name"`
	DbUser            string `json:"db_user"`
	DbPass            string `json:"db_pass"`
	WitBearerToken    string `json:"wit_bearer_token"` // @todo: move it to the DB (bot's config)
}

// LoadConfig loads the configuration located at the given path
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	if err = json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
