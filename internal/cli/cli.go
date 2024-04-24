package cli

import (
	"flag"

	"github.com/glerchundi/subcommands"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/saltosystems/winrt-go/internal/codegen"
)

const methodFilterUsage = `The filter to use when generating the methods. This option can be set several times, 
the given filters will be applied in order, and the first that matches will determine the result. The generator
will allow any method by default. The filter uses the overloaded method name to discriminate between overloaded
methods.

You can use the '!' character to negate a filter. For example, to generate all methods except the 'Add' method:
    -method-filter !Add

You can also use the '*' character to match any method, so if you want to generate only the 'Add' method, you can do:
    -method-filter Add -method-filter !*`

// NewGenerateCommand returns a new subcommand for generating code.
func NewGenerateCommand(logger log.Logger) *subcommands.Command {
	cfg := codegen.NewConfig()
	fs := flag.NewFlagSet("winrt-go-gen", flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")
	fs.BoolVar(&cfg.ValidateOnly, "validate", cfg.ValidateOnly, "validate the existing code instead of generating it")
	fs.StringVar(&cfg.Class, "class", cfg.Class, "The class to generate. This should include the namespace and the class name, e.g. 'System.Runtime.InteropServices.WindowsRuntime.EventRegistrationToken'.")
	fs.Func("method-filter", methodFilterUsage, func(m string) error {
		cfg.AddMethodFilter(m)
		return nil
	})
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enables the debug logging.")
	return subcommands.NewCommand(fs.Name(), fs, func() error {
		if cfg.Debug {
			logger = level.NewFilter(logger, level.AllowDebug())
		} else {
			logger = level.NewFilter(logger, level.AllowInfo())
		}
		return codegen.Generate(cfg, logger)
	})
}
