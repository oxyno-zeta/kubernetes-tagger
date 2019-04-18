package config

import (
	"errors"

	"github.com/thoas/go-funk"
)

// RecommendedConfigFileName Recommended Configuration File Name
const RecommendedConfigFileName = "config"

// AWSProviderName AWS provider name
const AWSProviderName = "aws"

// SupportedProviders List of supported providers
var SupportedProviders = []string{AWSProviderName}

// ErrNoProviderSelected No provider selected error
var ErrNoProviderSelected = errors.New("no provider selected")

// ErrProviderNotSupported Provider not supported error
var ErrProviderNotSupported = errors.New("provider not supported")

// ErrEmptyAWSConfiguration Error Empty AWS Configuration
var ErrEmptyAWSConfiguration = errors.New("aws configuration is empty")

// ErrEmptyAWSRegionConfiguration Error Empty AWS Region Configuration
var ErrEmptyAWSRegionConfiguration = errors.New("aws region is empty in configuration")

// Configuration configuration
type Configuration struct {
	Namespace  string        `mapstructure:"namespace"`
	Kubeconfig string        `mapstructure:"kubeconfig"`
	Address    string        `mapstructure:"address"`
	LogLevel   string        `mapstructure:"loglevel"`
	LogFormat  string        `mapstructure:"logformat"`
	AWS        *AWSConfig    `mapstructure:"aws"`
	Rules      []*RuleConfig `mapstructure:"rules"`
	Provider   string        `mapstructure:"provider"`
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

// IsValid Checks if the configuration is valid
func (cfg *Configuration) IsValid() error {
	// Check if provider is empty
	if cfg.Provider == "" {
		return ErrNoProviderSelected
	}

	// Check if provider is supported
	if !funk.Contains(SupportedProviders, cfg.Provider) {
		return ErrProviderNotSupported
	}

	// Check AWS configuration is ok if provider is aws
	if cfg.Provider == AWSProviderName {
		// Check that aws configuration block exists
		if cfg.AWS == nil {
			return ErrEmptyAWSConfiguration
		}
		// Check that region is set in AWS configuration block
		if cfg.AWS.Region == "" {
			return ErrEmptyAWSRegionConfiguration
		}
	}

	return nil
}
