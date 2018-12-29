package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
)

const (
	defaultPort         = 40000
	defaultBaseURL      = "http://localhost"
	defaultTemplatesDir = "./templates"
)

// Result holds the correct result hash for each challenge.
type Result struct {
	Name   string `toml:"name"`
	Output string `toml:"output"`
}

// Config holds the main configuration items.
type Config struct {
	// Our secret. This is used to calculate the hash. Once set, it
	// cannot be changed easily. Also, this should be kept secret,
	// as the name implies (i.e, DO NOT add the production version
	// of this to the repo).
	Secret  string   `toml:"secret"`
	Results []Result `toml:"result"`

	// HTTP port to listen on.
	Port int `toml:"port"`

	// Base URL for the XMLHttpRequests (from JS).
	BaseURL string `toml:"base_url"`

	// Directory for templates.
	TemplatesDir string `toml:"templates_dir"`
}

// parseConfig parses the configuration string from the slice of bytes
// containing the TOML config read from disk and performs basic sanity checking
// of configuration items.
func parseConfig(r io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	if _, err := toml.Decode(string(data), &config); err != nil {
		return Config{}, err
	}

	// Defaults
	if config.Port == 0 {
		config.Port = defaultPort
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.TemplatesDir == "" {
		config.BaseURL = defaultTemplatesDir
	}

	if config.Secret == "" {
		return Config{}, errors.New("Fatal: Secret is empty")
	}

	// Make input consistent.
	config.BaseURL = trimSlash(config.BaseURL)

	return config, nil
}
