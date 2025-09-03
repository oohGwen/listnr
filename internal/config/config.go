package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	MusicRoutes     []string `json:"music_routes"`
	Volume          float64  `json:"volume"`
	LastPath        string   `json:"last_path"`
	AutoplayEnabled bool     `json:"autoplay_enabled"`
	RepeatMode      bool     `json:"repeat_mode"`
}

func Load() (*Config, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(usr.HomeDir, ".config", "listnr.json")

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &Config{
			MusicRoutes:     []string{filepath.Join(usr.HomeDir, "Music")},
			Volume:          0.5,
			LastPath:        "",
			AutoplayEnabled: true,
			RepeatMode:      false,
		}

		// Create .config directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}

		// Write default config
		if err := Save(defaultConfig, configPath); err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}

	// Load existing config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Save(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
