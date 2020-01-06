package main

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	// The directory containing notes. Defaults to ~/Notes.
	NotesDirectory string
	// The file to write logs to, if omitted no logs will be written.
	LogFile string
}

func loadConfigFromFile(homeDir string, config *Config) error {
	file, err := os.Open(path.Join(homeDir, ".go-notes"))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return json.NewDecoder(file).Decode(config)
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	homeDir, errHomeDir := os.UserHomeDir()
	if errHomeDir == nil {
		if err := loadConfigFromFile(homeDir, config); err != nil {
			return nil, err
		}
	}
	if config.NotesDirectory == "" {
		if errHomeDir != nil {
			return nil, errHomeDir
		}
		config.NotesDirectory = path.Join(homeDir, "Notes")
	}
	return config, nil
}
