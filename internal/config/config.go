package config

import (
	"os"
	"strconv"
)

// Config holds server and tool defaults (no WORKSPACE_ROOT).
type Config struct {
	Port            int
	ShellTimeoutSec int
	MaxRequestBody  int64
}

// Load reads config from env with defaults.
func Load() Config {
	c := Config{
		Port:            8000,
		ShellTimeoutSec: 120,
		MaxRequestBody:  10 << 20, // 10 MiB
	}
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			c.Port = v
		}
	}
	if t := os.Getenv("SHELL_TIMEOUT_SEC"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 {
			c.ShellTimeoutSec = v
		}
	}
	if m := os.Getenv("MAX_REQUEST_BODY"); m != "" {
		if v, err := strconv.ParseInt(m, 10, 64); err == nil && v > 0 {
			c.MaxRequestBody = v
		}
	}
	return c
}
