package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type languageConfig struct {
	BuildCmd         []string      `json:"build"`
	RunCmd           []string      `json:"run"`
	BuildMemoryLimit int           `json:"build_memory_limit"`
	BuildTimeout     time.Duration `json:"build_timeout"`
	BuildMaxFileSize int           `json:"build_max_file_size"`
	CodeFile string `json:"codefile"`
}

func NewLangConfigFromFile(path string) (*languageConfig, error) {
	file, err := os.Open(filepath.Join(path, "config.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	cfg := new(languageConfig)
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}
	cfg.BuildTimeout *= time.Millisecond
	return cfg, nil
}
