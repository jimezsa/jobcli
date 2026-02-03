package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

const (
	DirName         = "jobcli"
	ConfigFileName  = "config.json"
	ProxiesFileName = "proxies.txt"
	CookiesFileName = "cookies.json"
)

// Config contains default search settings.
type Config struct {
	DefaultLocation string `json:"default_location"`
	DefaultCountry  string `json:"default_country"`
	DefaultLimit    int    `json:"default_limit"`
}

func DefaultConfig() Config {
	return Config{
		DefaultLocation: envString("JOBCLI_DEFAULT_LOCATION", ""),
		DefaultCountry:  envString("JOBCLI_DEFAULT_COUNTRY", "usa"),
		DefaultLimit:    envInt("JOBCLI_DEFAULT_LIMIT", 20),
	}
}

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, DirName), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func ProxiesPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ProxiesFileName), nil
}

func Load() (Config, error) {
	cfg := DefaultConfig()
	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return cfg, nil
	}

	if err := json5.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Init writes default config.json and proxies.txt if they don't already exist.
func Init() ([]string, error) {
	var created []string

	dir, err := ConfigDir()
	if err != nil {
		return created, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return created, err
	}

	configPath := filepath.Join(dir, ConfigFileName)
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		if err := writeConfig(configPath, DefaultConfig()); err != nil {
			return created, err
		}
		created = append(created, configPath)
	}

	proxiesPath := filepath.Join(dir, ProxiesFileName)
	if _, err := os.Stat(proxiesPath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(proxiesPath, []byte(""), 0o644); err != nil {
			return created, err
		}
		created = append(created, proxiesPath)
	}

	return created, nil
}

func writeConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func LoadProxies(flagValue string) ([]string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return splitCSV(flagValue), nil
	}

	if env := strings.TrimSpace(os.Getenv("JOBCLI_PROXIES")); env != "" {
		return splitCSV(env), nil
	}

	path, err := ProxiesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var proxies []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		proxies = append(proxies, line)
	}
	return proxies, nil
}

func envString(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func envInt(key string, fallback int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
