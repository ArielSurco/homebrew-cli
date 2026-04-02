package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

// Config holds the global CLI configuration.
type Config struct {
	Projects []Project `toml:"projects"`
}

// Project represents a registered project.
type Project struct {
	Name      string `toml:"name"`
	Path      string `toml:"path"`
	DevScript string `toml:"dev_script,omitempty"`
}

// ActiveModules holds the active module list.
type ActiveModules struct {
	Modules ModulesSection `toml:"modules"`
}

// ModulesSection holds active module names.
type ModulesSection struct {
	Active []string `toml:"active"`
}

// localConfig parses .arielsurco-cli.toml in CWD.
type localConfig struct {
	Project localProject `toml:"project"`
}

type localProject struct {
	Name      string `toml:"name"`
	DevScript string `toml:"dev_script,omitempty"`
}

// GlobalConfigPath returns the path to the global config file.
// Calls xdg.Reload() to pick up any XDG_CONFIG_HOME changes (e.g. in tests).
func GlobalConfigPath() (string, error) {
	xdg.Reload()
	return filepath.Join(xdg.ConfigHome, "arielsurco-cli", "config.toml"), nil
}

// ActiveModulesPath returns the path to the active modules config file.
// Calls xdg.Reload() to pick up any XDG_CONFIG_HOME changes (e.g. in tests).
func ActiveModulesPath() (string, error) {
	xdg.Reload()
	return filepath.Join(xdg.ConfigHome, "arielsurco-cli", "active-modules.toml"), nil
}

// Load reads the global config, then merges any local .arielsurco-cli.toml overrides.
// A missing global config is not an error — returns an empty Config.
func Load() (*Config, error) {
	globalPath, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}

	cfg, err := loadTOMLFile[Config](globalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = &Config{}
		} else {
			return nil, err
		}
	}

	// Check for local .arielsurco-cli.toml in CWD
	local, err := loadLocalConfig()
	if err == nil && local != nil && local.Project.Name != "" {
		mergeLocalConfig(cfg, local)
	}

	return cfg, nil
}

// Save atomically writes the config to the global config path.
func Save(cfg *Config) error {
	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}
	return atomicSaveTOML(path, cfg)
}

// LoadActive reads the active modules config. A missing file returns empty ActiveModules.
func LoadActive() (*ActiveModules, error) {
	path, err := ActiveModulesPath()
	if err != nil {
		return nil, err
	}

	activeModules, err := loadTOMLFile[ActiveModules](path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ActiveModules{}, nil
		}
		return nil, err
	}
	return activeModules, nil
}

// SaveActive atomically writes the active modules config.
func SaveActive(activeModules *ActiveModules) error {
	path, err := ActiveModulesPath()
	if err != nil {
		return err
	}
	return atomicSaveTOML(path, activeModules)
}

// loadTOMLFile reads and decodes a TOML file into T.
func loadTOMLFile[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result T
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// atomicSaveTOML marshals value to TOML and writes it atomically via tempfile + rename.
func atomicSaveTOML(path string, value any) error {
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	data, err := toml.Marshal(value)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(configDir, "*.toml.tmp")
	if err != nil {
		return err
	}
	tempFilePath := tempFile.Name()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFilePath)
		return err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFilePath)
		return err
	}

	return os.Rename(tempFilePath, path)
}

// loadLocalConfig reads .arielsurco-cli.toml from the current working directory.
func loadLocalConfig() (*localConfig, error) {
	data, err := os.ReadFile(".arielsurco-cli.toml")
	if err != nil {
		return nil, err
	}
	var override localConfig
	if err := toml.Unmarshal(data, &override); err != nil {
		return nil, err
	}
	return &override, nil
}

// mergeLocalConfig merges local project overrides into the global config.
// Local dev_script wins for matching project name.
func mergeLocalConfig(cfg *Config, local *localConfig) {
	for i := range cfg.Projects {
		if cfg.Projects[i].Name == local.Project.Name && local.Project.DevScript != "" {
			cfg.Projects[i].DevScript = local.Project.DevScript
			return
		}
	}
}
