package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
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

	if config.Secret == "" {
		return Config{}, errors.New("Fatal: Secret is empty")
	}

	return config, nil
}
