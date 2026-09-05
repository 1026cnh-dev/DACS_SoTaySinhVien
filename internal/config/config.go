package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Runtime describes which configuration file was selected at startup.
// Environment variables already present in the operating system always win
// over values in the selected file, which keeps Render secrets out of Git.
type Runtime struct {
	Environment string
	ConfigFile  string
}

// Load selects and loads the environment configuration without requiring code
// changes between local development and Render production.
//
// Selection priority:
//  1. CONFIG_FILE environment variable (explicit file path)
//  2. Render runtime detection (RENDER is set) -> production
//  3. APP_ENV environment variable
//  4. config/environment file
//  5. local
//
// The selected file is config/local.env or config/production.env unless an
// explicit CONFIG_FILE is supplied. Existing OS environment variables are not
// overwritten by file values.
func Load() (Runtime, error) {
	envName := detectEnvironment()
	configFile := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	if configFile == "" {
		configFile = filepath.Join("config", envName+".env")
	}

	if err := loadEnvFile(configFile); err != nil {
		if os.IsNotExist(err) {
			return Runtime{}, fmt.Errorf("không tìm thấy file cấu hình %s: %w", configFile, err)
		}
		return Runtime{}, err
	}

	// Keep the resolved environment available to every existing package that
	// reads APP_ENV. Do not replace an explicit APP_ENV supplied by Render/user.
	if strings.TrimSpace(os.Getenv("APP_ENV")) == "" {
		_ = os.Setenv("APP_ENV", envName)
	}
	_ = os.Setenv("DACS_CONFIG_FILE", configFile)

	rt := Runtime{Environment: normalizedEnvironment(os.Getenv("APP_ENV")), ConfigFile: configFile}
	log.Printf("config: environment=%s file=%s", rt.Environment, rt.ConfigFile)
	return rt, nil
}

func detectEnvironment() string {
	if strings.TrimSpace(os.Getenv("RENDER")) != "" {
		return "production"
	}
	if v := strings.TrimSpace(os.Getenv("APP_ENV")); v != "" {
		return normalizedEnvironment(v)
	}
	if b, err := os.ReadFile(filepath.Join("config", "environment")); err == nil {
		if v := strings.TrimSpace(string(b)); v != "" {
			return normalizedEnvironment(v)
		}
	}
	return "local"
}

func normalizedEnvironment(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "prod", "production", "render":
		return "production"
	case "dev", "development", "local", "localhost", "":
		return "local"
	default:
		// Custom environments can reuse config/<name>.env.
		return strings.ToLower(strings.TrimSpace(v))
	}
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: cấu hình phải có dạng KEY=VALUE", path, lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("%s:%d: tên biến môi trường rỗng", path, lineNo)
		}
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("%s:%d: không thể đặt %s: %w", path, lineNo, key, err)
		}
	}
	return s.Err()
}
