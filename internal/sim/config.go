package sim

import (
	"encoding/json"
	"fmt"
	"os"
)


type InputEvent struct {
	Tick uint64 `json:"tick"`
	Token interface{} `json:"token"`
}
type Config struct {
	MaxTicks uint64 `json:"max_ticks"`
	Events []InputEvent `json:"events"`
}

func (e InputEvent) TokenAsWord() (int32, error) {
	switch v := e.Token.(type) {
	case float64:
		return int32(v), nil
	case string:
		r := []rune(v)
		if len(r) == 1 {
			return int32(r[0]), nil
		}
		return 0, fmt.Errorf("token string must contain exactly one rune, got %q", v)
	default:
		return 0, fmt.Errorf("unsupported token type %T", v)
	}

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
