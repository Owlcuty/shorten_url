package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Parse(r io.ReadCloser) (*Config, error) {
	var cfg Config
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decoder json: %w", err)
	}

	return &cfg, nil
}

func ParseFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()
	return Parse(file)
}
