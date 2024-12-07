package isolate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type languageConfig struct {
	BuildScript      string        `json:"build_script"`
	RunScript        string        `json:"run_script"`
	BuildMemoryLimit int           `json:"build_memory_limit"`
	BuildTimeout     time.Duration `json:"build_timeout"`
	BuildMaxFileSize int           `json:"build_max_file_size"`
}

func NewLangConfigFromFile(path string) (*languageConfig, error) {
	file, err := os.Open(filepath.Join(path, "config.json"))
	if err != nil {
		return nil, err
	}
	cfg := new(languageConfig)
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}
	cfg.BuildScript = filepath.Join(path, cfg.BuildScript)
	cfg.RunScript = filepath.Join(path, cfg.RunScript)
	cfg.BuildTimeout *= 1000000
	return cfg, nil
}
