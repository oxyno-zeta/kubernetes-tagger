package config

// RecommendedConfigFileName Recommended Configuration File Name
const RecommendedConfigFileName = "config"

// Configuration configuration
type Configuration struct {
	Namespace string        `mapstructure:"namespace"`
	AWS       *AWSConfig    `mapstructure:"aws"`
	Rules     []*RuleConfig `mapstructure:"rules"`
}

// AWSConfig AWS Configuration
type AWSConfig struct {
	Region string `mapstructure:"region"`
}

// RuleConfig Rule Configuration
type RuleConfig struct {
	Tag    string             `mapstructure:"tag"`
	Query  string             `mapstructure:"query"`
	Value  string             `mapstructure:"value"`
	Action string             `mapstructure:"action"`
	When   []*ConditionConfig `mapstructure:"when"`
}

// ConditionConfig Condition Configuration
type ConditionConfig struct {
	Condition string `mapstructure:"condition"`
	Value     string `mapstructure:"value"`
	Operator  string `mapstructure:"operator"`
}
