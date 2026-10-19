package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct{ Addr, Host, Port, User, Password, From, Pepper string }

// LoadDotEnv reads a local .env file without overwriting existing environment variables.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}
func Load() Config {
	get := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}
	return Config{get("HTTP_ADDR", ":8080"), get("SMTP_HOST", "smtp.gmail.com"), get("SMTP_PORT", "587"), get("SMTP_USER", ""), get("SMTP_PASSWORD", ""), get("SMTP_FROM", ""), get("OTP_PEPPER", "")}
}
