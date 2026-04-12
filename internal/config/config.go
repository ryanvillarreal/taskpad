package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

/*
# exposed
Struct Config - base_url and port
ConfigPath() - externally available config pathing
Load() - creates or loads cfg file

# built-ins
defaults() - creates the Config struct with defaults
configPath() - gets the user's configuration path
writeDefaults() - writes the defaults if non-existent
overrideString() - Gets env variables first?
*/

type Config struct {
	BASE_URL string `json:"base_url"`
	PORT     string `json:"port"`
}

func ConfigPath() string {
	return configPath()
}

func Load() Config {
	cfg := defaults()

	if cp := configPath(); cp != "" {
		content, err := os.ReadFile(cp)
		if err == nil {
			_ = json.Unmarshal(content, &cfg)
		} else if os.IsNotExist(err) {
			_ = writeDefaults(cp, cfg)
		}
	}
	return cfg
}

// Built-ins

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {

		return ""
	}
	return filepath.Join(dir, "taskpad", "config.json")
}

// auto-generate sane defaults
func defaults() Config {
	return Config{
		BASE_URL: "http://localhost:8080",
		PORT:     "8080",
	}
}

func overrideString(target *string, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}

func writeDefaults(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
