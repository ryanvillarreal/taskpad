package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIURL        string       `json:"api_url"`
	NotesDir      string       `json:"notes_dir"`
	Editor        string       `json:"editor"`
	MigrationsDir string       `json:"migrations_dir"`
	Server        ServerConfig `json:"server"`
	CalDAV        CalDAVConfig `json:"caldav"`
}

type ServerConfig struct {
	Port   string `json:"port"`
	DBPath string `json:"db_path"`
}

type CalDAVConfig struct {
	URL          string `json:"url"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CalendarPath string `json:"calendar_path"`
}

func Load() Config {
	cfg := defaults()

	if configPath := configPath(); configPath != "" {
		if content, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(content, &cfg)
		}
	}

	overrideString(&cfg.APIURL, "TASKPAD_URL")
	overrideString(&cfg.NotesDir, "TASKPAD_NOTES_DIR")
	overrideString(&cfg.Editor, "TASKPAD_EDITOR")
	if cfg.Editor == "" {
		overrideString(&cfg.Editor, "EDITOR")
	}
	overrideString(&cfg.MigrationsDir, "MIGRATIONS_DIR")
	overrideString(&cfg.Server.Port, "PORT")
	overrideString(&cfg.Server.DBPath, "DB_PATH")
	overrideString(&cfg.CalDAV.URL, "TASKPAD_CALDAV_URL")
	overrideString(&cfg.CalDAV.Username, "TASKPAD_CALDAV_USER")
	overrideString(&cfg.CalDAV.Password, "TASKPAD_CALDAV_PASS")
	overrideString(&cfg.CalDAV.CalendarPath, "TASKPAD_CALDAV_CALENDAR")

	return cfg
}

func ConfigPath() string {
	return configPath()
}

func defaults() Config {
	return Config{
		APIURL:        "http://localhost:8080",
		MigrationsDir: "./migrations",
		Server: ServerConfig{
			Port:   "8080",
			DBPath: "./taskpad.db",
		},
	}
}

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "taskpad", "config.json")
}

func overrideString(target *string, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}
