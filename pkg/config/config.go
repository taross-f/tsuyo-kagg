package config

import (
	"encoding/json"
	"os"
)

// Config holds the application configuration
type Config struct {
	SplashURL      string `json:"splash_url"`
	KaggleBaseURL  string `json:"kaggle_base_url"`
	OutputFile     string `json:"output_file"`
	PageSize       int    `json:"page_size"`
	MaxPages       int    `json:"max_pages"`
	TargetCountry  string `json:"target_country"`
	RequestTimeout int    `json:"request_timeout"`
	WaitTime       int    `json:"wait_time"`
	MinDelay       int    `json:"min_delay"`
	MaxDelay       int    `json:"max_delay"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		SplashURL:      "http://localhost:8050",
		KaggleBaseURL:  "https://www.kaggle.com",
		OutputFile:     "output.csv",
		PageSize:       20,
		MaxPages:       250,
		TargetCountry:  "Japan",
		RequestTimeout: 10,
		WaitTime:       5,
		MinDelay:       1,
		MaxDelay:       5,
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (Config, error) {
	config := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
