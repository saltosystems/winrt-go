package codegen

import (
	"fmt"
)

// Config is the configuration for the code generation.
type Config struct {
	Debug         bool
	Class         string
	ValidateOnly  bool
	methodFilters []string
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{}
}

// AddMethodFilter adds a method to the list of methodFilters to generate.
func (cfg *Config) AddMethodFilter(methodFilter string) {
	cfg.methodFilters = append(cfg.methodFilters, methodFilter)
}

// MethodFilter creates and returns a new method filter for the current config.
func (cfg *Config) MethodFilter() *MethodFilter {
	return NewMethodFilter(cfg.methodFilters)
}

// Validate validates the Config and returns an error if there's any problem.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.Class == "" {
		return fmt.Errorf("generated classes may not be empty")
	}

	return nil
}
