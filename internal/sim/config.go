package sim

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	MaxTicks uint64 `json:"max_ticks"`
}

func LoadConfig(path string) (Config, error) {
	if path == "" {
		return Config{MaxTicks: 200_000}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if c.MaxTicks == 0 {
		c.MaxTicks = 200_000
	}
	return c, nil
}
