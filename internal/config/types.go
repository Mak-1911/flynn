// Package config provides configuration types for Flynn.
package config

// Config represents the main Flynn configuration.
type Config struct {
	Instance InstanceConfig `toml:"instance"`
	Tenant   TenantConfig   `toml:"tenant"`
	User     UserConfig     `toml:"user"`
	Models   ModelConfig    `toml:"models"`
	Features Features       `toml:"features"`
	Paths    PathsConfig    `toml:"paths"`
	Privacy  PrivacyConfig  `toml:"privacy"`
	Graph    GraphConfig    `toml:"graph"`
}

// InstanceConfig contains instance-level settings.
type InstanceConfig struct {
	ID        string `toml:"id"`
	MaxAgents int    `toml:"max_agents"`
}

// TenantConfig contains team/tenant settings.
type TenantConfig struct {
	ID      string       `toml:"id"`
	Name    string       `toml:"name"`
	Members []TeamMember `toml:"members"`
}

// TeamMember represents a team member.
type TeamMember struct {
	ID   string `toml:"id"`
	Name string `toml:"name"`
	Role string `toml:"role"` // admin, member
}

// UserConfig contains user preferences.
type UserConfig struct {
	Name                 string `toml:"name"`
	Timezone             string `toml:"timezone"`
	Language             string `toml:"language"`
	ResponseStyle        string `toml:"response_style"`   // concise, balanced, detailed
	CostSensitivity      string `toml:"cost_sensitivity"` // aggressive, balanced, quality
	ProactiveSuggestions bool   `toml:"proactive_suggestions"`
}

// ModelConfig contains model-related settings.
type ModelConfig struct {
	Local LocalModelConfig `toml:"local"`
	Cloud CloudModelConfig `toml:"cloud"`
}

// LocalModelConfig configures local model inference.
type LocalModelConfig struct {
	PrimaryModel string `toml:"primary_model"`
	IntentModel  string `toml:"intent_model"`
	Threads      int    `toml:"threads"`
	ContextSize  int    `toml:"context_size"`
	GPULayers    int    `toml:"gpu_layers"`
	ModelsDir    string `toml:"models_dir"`
}

// CloudModelConfig configures cloud model usage.
type CloudModelConfig struct {
	Provider      string  `toml:"provider"`
	DefaultModel  string  `toml:"default_model"`
	Mode          string  `toml:"mode"` // never, smart, always
	MonthlyBudget float64 `toml:"monthly_budget"`
	APIKey        string  `toml:"api_key"` // API key for cloud provider
	Enabled       bool    `toml:"enabled"` // Whether cloud is enabled
}

// Features contains feature flags.
type Features struct {
	Calendar      bool `toml:"calendar"`
	Email         bool `toml:"email"`
	Notes         bool `toml:"notes"`
	Browser       bool `toml:"browser"`
	CostDashboard bool `toml:"cost_dashboard"`
}

// PathsConfig contains file path settings.
type PathsConfig struct {
	DataDir    string `toml:"data_dir"`
	LogsDir    string `toml:"logs_dir"`
	CacheDir   string `toml:"cache_dir"`
	PersonalDB string `toml:"personal_db"`
	TeamDB     string `toml:"team_db"`
}

// PrivacyConfig contains privacy settings.
type PrivacyConfig struct {
	AllowCloudFor   []string `toml:"allow_cloud_for"`
	SensitiveTopics []string `toml:"sensitive_topics"`
	Anonymize       bool     `toml:"anonymize"`
}

// GraphConfig contains knowledge graph settings.
type GraphConfig struct {
	Enabled       bool `toml:"enabled"`
	UseLLM        bool `toml:"use_llm"`
	MaxEntities   int  `toml:"max_entities"`
	MaxRelations  int  `toml:"max_relations"`
	MaxChunkBytes int  `toml:"max_chunk_bytes"`
}

// ThreadMode represents the visibility of a conversation.
type ThreadMode string

const (
	ThreadModePersonal ThreadMode = "personal"
	ThreadModeTeam     ThreadMode = "team"
)

// ResponseStyle represents the desired response format.
type ResponseStyle string

const (
	ResponseStyleConcise  ResponseStyle = "concise"
	ResponseStyleBalanced ResponseStyle = "balanced"
	ResponseStyleDetailed ResponseStyle = "detailed"
)

// CostSensitivity represents the cost vs quality preference.
type CostSensitivity string

const (
	CostSensitivityAggressive CostSensitivity = "aggressive"
	CostSensitivityBalanced   CostSensitivity = "balanced"
	CostSensitivityQuality    CostSensitivity = "quality"
)

// CloudMode represents when to use cloud models.
type CloudMode string

const (
	CloudModeNever  CloudMode = "never"
	CloudModeSmart  CloudMode = "smart"
	CloudModeAlways CloudMode = "always"
)
