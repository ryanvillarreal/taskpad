package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	BaseURL    string       `json:"base_url"`
	Host       string       `json:"host"`
	Port       string       `json:"port"`
	NotesDir   string       `json:"notes_dir"`
	TasksDir   string       `json:"tasks_dir"`
	LinksDir   string       `json:"links_dir"`
	FetchLimit int          `json:"fetch_limit"`
	TLS        TLSConfig    `json:"tls"`
	Notify     NotifyConfig `json:"notify"`
}

type NotifyConfig struct {
	Backend       string `json:"backend"`
	URL           string `json:"url"`
	Topic         string `json:"topic"`
	RepeatMinutes int    `json:"repeat_minutes"`
}

type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
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

	if cfg.NotesDir == "" {
		cfg.NotesDir = defaultNotesDir()
	}
	if cfg.TasksDir == "" {
		cfg.TasksDir = defaultTasksDir(cfg.NotesDir)
	}
	if cfg.LinksDir == "" {
		cfg.LinksDir = defaultLinksDir(cfg.NotesDir)
	}
	if cfg.FetchLimit == 0 {
		cfg.FetchLimit = 5000
	}
	_ = os.MkdirAll(cfg.NotesDir, 0o755)
	_ = os.MkdirAll(cfg.TasksDir, 0o755)
	_ = os.MkdirAll(cfg.LinksDir, 0o755)

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

func defaults() Config {
	notesDir := defaultNotesDir()
	return Config{
		BaseURL:  "http://localhost:8080",
		Host:     "",
		Port:     "8080",
		NotesDir:   notesDir,
		TasksDir:   defaultTasksDir(notesDir),
		LinksDir:   defaultLinksDir(notesDir),
		FetchLimit: 5000,
		TLS: TLSConfig{
			Enabled:  false,
			CertFile: "",
			KeyFile:  "",
		},
		Notify: NotifyConfig{
			Backend:       "ntfy",
			URL:           "https://ntfy.sh",
			Topic:         "",
			RepeatMinutes: 60,
		},
	}
}

func defaultTasksDir(notesDir string) string {
	return filepath.Join(notesDir, "tasks")
}

func defaultLinksDir(notesDir string) string {
	return filepath.Join(notesDir, "links")
}

func defaultNotesDir() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "taskpad", "notes")
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
