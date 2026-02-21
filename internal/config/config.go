// Package config handles Flynn configuration loading and management.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Default returns the default configuration.
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".flynn")

	return &Config{
		Instance: InstanceConfig{
			ID:        "flynn-local",
			MaxAgents: 3,
		},
		Tenant: TenantConfig{
			ID:   "default",
			Name: "Default Tenant",
			Members: []TeamMember{
				{ID: "user-local", Name: "Local User", Role: "admin"},
			},
		},
		User: UserConfig{
			Name:                 "",
			Timezone:             "UTC",
			Language:             "en",
			ResponseStyle:        string(ResponseStyleBalanced),
			CostSensitivity:      string(CostSensitivityBalanced),
			ProactiveSuggestions: true,
		},
		Models: ModelConfig{
			Local: LocalModelConfig{
				PrimaryModel: "qwen-2.5-7b",
				IntentModel:  "qwen-2.5-3b",
				Threads:      4,
				ContextSize:  4096,
				GPULayers:    0,
				ModelsDir:    filepath.Join(dataDir, "models"),
			},
			Cloud: CloudModelConfig{
				Provider:      "openrouter",
				DefaultModel:  "deepseek/deepseek-r1",
				Mode:          string(CloudModeSmart),
				MonthlyBudget: 10.0,
			},
		},
		Features: Features{
			Calendar:      false,
			Email:         false,
			Notes:         false,
			Browser:       false,
			CostDashboard: true,
		},
		Paths: PathsConfig{
			DataDir:    dataDir,
			LogsDir:    filepath.Join(dataDir, "logs"),
			CacheDir:   filepath.Join(dataDir, "cache"),
			PersonalDB: filepath.Join(dataDir, "personal.db"),
			TeamDB:     filepath.Join(dataDir, "team.db"),
		},
		Privacy: PrivacyConfig{
			AllowCloudFor:   []string{"research", "coding", "analysis"},
			SensitiveTopics: []string{"health", "finance", "passwords"},
			Anonymize:       true,
		},
	}
}

// Load loads the configuration from the given path.
// If the file doesn't exist, returns defaults.
func Load(configPath string) (*Config, error) {
	cfg := Default()

	// Try to read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return defaults
			return cfg, nil
		}
		return nil, err
	}

	// Parse TOML
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Expand paths
	cfg = expandPaths(cfg)

	return cfg, nil
}

// Save saves the configuration to the given path.
func (c *Config) Save(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write TOML
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}

// SaveConfig is a convenience function to save a config.
func SaveConfig(configPath string, cfg *Config) error {
	return cfg.Save(configPath)
}

// DefaultInstanceConfig returns a default instance config for CLI use.
func DefaultInstanceConfig() *Config {
	return Default()
}

// expandPaths expands ~ and environment variables in paths.
func expandPaths(cfg *Config) *Config {
	homeDir, _ := os.UserHomeDir()

	if cfg.Paths.DataDir == "" || cfg.Paths.DataDir[0] == '~' {
		cfg.Paths.DataDir = filepath.Join(homeDir, cfg.Paths.DataDir[1:])
	}
	if cfg.Paths.LogsDir == "" || cfg.Paths.LogsDir[0] == '~' {
		cfg.Paths.LogsDir = filepath.Join(homeDir, cfg.Paths.LogsDir[1:])
	}
	if cfg.Paths.CacheDir == "" || cfg.Paths.CacheDir[0] == '~' {
		cfg.Paths.CacheDir = filepath.Join(homeDir, cfg.Paths.CacheDir[1:])
	}
	if cfg.Paths.PersonalDB == "" || cfg.Paths.PersonalDB[0] == '~' {
		cfg.Paths.PersonalDB = filepath.Join(homeDir, cfg.Paths.PersonalDB[1:])
	}
	if cfg.Paths.TeamDB == "" || cfg.Paths.TeamDB[0] == '~' {
		cfg.Paths.TeamDB = filepath.Join(homeDir, cfg.Paths.TeamDB[1:])
	}
	if cfg.Models.Local.ModelsDir == "" || cfg.Models.Local.ModelsDir[0] == '~' {
		cfg.Models.Local.ModelsDir = filepath.Join(homeDir, cfg.Models.Local.ModelsDir[1:])
	}

	return cfg
}

// GetTenantID returns the tenant ID.
func (c *Config) GetTenantID() string {
	return c.Tenant.ID
}

// GetPersonalDBPath returns the path to the personal database.
func (c *Config) GetPersonalDBPath() string {
	return c.Paths.PersonalDB
}

// GetTeamDBPath returns the path to the team database.
func (c *Config) GetTeamDBPath() string {
	return c.Paths.TeamDB
}

// IsCloudEnabled returns true if cloud mode is enabled.
func (c *Config) IsCloudEnabled() bool {
	return CloudMode(c.Models.Cloud.Mode) != CloudModeNever
}

// CanUseCloudFor returns true if the given category can use cloud.
func (c *Config) CanUseCloudFor(category string) bool {
	if !c.IsCloudEnabled() {
		return false
	}
	for _, allowed := range c.Privacy.AllowCloudFor {
		if allowed == category {
			return true
		}
	}
	return false
}

// IsSensitiveTopic returns true if the topic is sensitive.
func (c *Config) IsSensitiveTopic(topic string) bool {
	for _, sensitive := range c.Privacy.SensitiveTopics {
		if sensitive == topic {
			return true
		}
	}
	return false
}
