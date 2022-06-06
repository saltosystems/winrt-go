package main

import (
	"flag"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/peterbourgon/ff/v3"
	"github.com/saltosystems/winrt-go/internal/cli"
)

func main() {
	logger := createLogger()
	winrtGoGenCmd := cli.NewGenerateCommand(logger)

	err := winrtGoGenCmd.Execute(os.Args[1:], func(fs *flag.FlagSet, args []string) error {
		return ff.Parse(fs, args,
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithEnvVarPrefix(strings.ToUpper(strings.ReplaceAll(winrtGoGenCmd.Name(), "-", "_"))),
		)
	})
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}

func createLogger() log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = level.NewInjector(logger, level.InfoValue())

	return logger
}
