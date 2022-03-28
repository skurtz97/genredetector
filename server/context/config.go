package context

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// If successful, returns a pointer to a new Config initialized with the values
// contained the file at $USER_CONFIG_DIR/genredetector/config.json.
func NewConfig() (*Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path.Join(dir, "genredetector/config.json"))
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
