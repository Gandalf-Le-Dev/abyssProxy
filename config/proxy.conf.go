package config

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
)

type Location struct {
	ProxyPass    string `json:"proxyPass"`
	RequiredAuth bool   `json:"requiredAuth,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	Auth         struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth,omitempty"`
}

type ServerConfig struct {
	Location Location `json:"location"`
}

type HTTPConfig struct {
	Servers map[string]ServerConfig `json:"servers"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(file string, cfg *HTTPConfig) error {
	if !regexp.MustCompile(`^.*\.json$`).MatchString(file) {
		return errors.New("invalid file path. Must be a .json file")
	}

	jsonFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteFile, cfg)
	if err != nil {
		return err
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(file string, cfg *HTTPConfig) error {
	if !regexp.MustCompile(`^.*\.json$`).MatchString(file) {
		return errors.New("invalid file path. Must be a .json file")
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
