package codegen

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/saltosystems/winrt-go/winmd"
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

type classNotFoundError struct {
	class string
}

func (e *classNotFoundError) Error() string {
	return fmt.Sprintf("class %s was not found", e.class)
}

type generator struct {
	logger log.Logger
}

// Generate generates the code for the given config.
func Generate(cfg *Config, logger log.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	g := &generator{
		logger: logger,
	}
	return g.run(cfg)
}

func (g *generator) run(cfg *Config) error {
	_ = level.Debug(g.logger).Log("msg", "starting code generation", "class", cfg.Class)

	fs, err := winmd.AllFiles()
	if err != nil {
		return err
	}

	// we don't know which winmd file contains the class, so we have to iterate over all of them
	for _, f := range fs {
		if err := g.generateType(f.Name(), cfg.Class); err != nil {
			// class not found errors are ok
			if _, ok := err.(*classNotFoundError); ok {
				continue
			}

			return err
		}

		_ = level.Debug(g.logger).Log("msg", "found class", "class", cfg.Class, "file", f.Name())
		return nil
	}

	return fmt.Errorf("class %s was not found", cfg.Class)

}

func (g *generator) generateType(path, class string) error {
	f, err := winmd.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	c, err := types.FromPE(f)
	if err != nil {
		return err
	}

	tt := c.Table(md.TypeDef)
	for i := uint32(0); i < tt.RowCount(); i++ {
		var t types.TypeDef
		if err := t.FromRow(tt.Row(i)); err != nil {
			return err
		}

		if t.TypeNamespace+"."+t.TypeName == class {
			// TODO
			return nil
		}
	}
	return &classNotFoundError{class: class}
}
