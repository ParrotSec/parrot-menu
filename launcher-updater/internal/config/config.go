package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	systemConfigDir = "/usr/share/parrot-menu/config.d"
	localConfigPath = "/etc/parrot-menu/launcher.conf"
	optionName      = "ShowInstallableLaunchers"
)

type Options struct {
	ShowInstallableLaunchers bool
}

func Load() (Options, error) {
	return load(systemConfigDir, localConfigPath)
}

func load(configDir, overridePath string) (Options, error) {
	var options Options

	entries, err := os.ReadDir(configDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return options, fmt.Errorf("read config directory %s: %w", configDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}

		path := filepath.Join(configDir, entry.Name())
		options, err = applyFile(path, options, false)
		if err != nil {
			return Options{}, err
		}
	}

	options, err = applyFile(overridePath, options, true)
	if err != nil {
		return Options{}, err
	}
	return options, nil
}

func applyFile(path string, options Options, optional bool) (Options, error) {
	data, err := os.ReadFile(path)
	if optional && errors.Is(err, os.ErrNotExist) {
		return options, nil
	}
	if err != nil {
		return options, fmt.Errorf("read config file %s: %w", path, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return options, fmt.Errorf("%s:%d: expected key=value", path, lineNumber)
		}
		if strings.TrimSpace(key) != optionName {
			return options, fmt.Errorf(
				"%s:%d: unknown option %q", path, lineNumber, strings.TrimSpace(key))
		}

		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true":
			options.ShowInstallableLaunchers = true
		case "false":
			options.ShowInstallableLaunchers = false
		default:
			return options, fmt.Errorf(
				"%s:%d: %s must be true or false", path, lineNumber, optionName)
		}
	}

	if err := scanner.Err(); err != nil {
		return options, fmt.Errorf("read config file %s: %w", path, err)
	}
	return options, nil
}
