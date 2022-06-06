package cli

import (
	"flag"

	"github.com/glerchundi/subcommands"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/saltosystems/winrt-go/internal/codegen"
)

// NewGenerateCommand returns a new subcommand for generating code.
func NewGenerateCommand(logger log.Logger) *subcommands.Command {
	cfg := codegen.NewConfig()
	fs := flag.NewFlagSet("winrt-go-gen", flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")
	fs.StringVar(&cfg.Class, "class", cfg.Class, "The class to generate. This should include the namespace and the class name, e.g. 'System.Runtime.InteropServices.WindowsRuntime.EventRegistrationToken'")
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enables the debug logging.")
	return subcommands.NewCommand(fs.Name(), fs, func() error {
		if cfg.Debug {
			logger = level.NewFilter(logger, level.AllowDebug())
		}
		return codegen.Generate(cfg, logger)
	})
}
