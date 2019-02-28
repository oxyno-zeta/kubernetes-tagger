package config

// RecommendedConfigFileName Recommended Configuration File Name
const RecommendedConfigFileName = "config"

// MainConfiguration Main Configuration
type MainConfiguration struct {
	Config *Configuration `mapstructure:"config"`
	Rules  []*RuleConfig  `mapstructure:"rules"`
}

// Configuration configuration
type Configuration struct {
	AWS *AWSConfig `mapstructure:"aws"`
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
