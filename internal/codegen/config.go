package codegen

import "fmt"

// Config is the configuration for the code generation.
type Config struct {
	Debug bool
	Class string

	SkipFactory bool
	SkipStatics bool
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{}
}

// Validate validates the Config and returns an error if there's any problem.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.Class == "" {
		return fmt.Errorf("generated class may not be empty")
	}

	return nil
}
