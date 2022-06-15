package metadata

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/saltosystems/winrt-go/winmd"
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

// ClassNotFoundError is returned when a class is not found.
type ClassNotFoundError struct {
	Class string
}

func (e *ClassNotFoundError) Error() string {
	return fmt.Sprintf("class %s was not found", e.Class)
}

// Store holds the windows metadata contexts. It can be used to get the metadata across multiple files.
type Store struct {
	contexts map[string]*types.Context
	logger   log.Logger
}

// NewStore loads all windows metadata files and returns a new Store.
func NewStore(logger log.Logger) (*Store, error) {
	contexts := make(map[string]*types.Context)

	winmdFiles, err := winmd.AllFiles()
	if err != nil {
		return nil, err
	}

	// parse and store all files in memory
	for _, f := range winmdFiles {
		winmdCtx, err := parseWinMDFile(f.Name())
		if err != nil {
			return nil, err
		}
		contexts[f.Name()] = winmdCtx
	}

	return &Store{
		contexts: contexts,
		logger:   logger,
	}, nil
}

func parseWinMDFile(path string) (*types.Context, error) {
	f, err := winmd.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return types.FromPE(f)
}

// TypeDefByName returns a type definition that matches the given name.
func (mds *Store) TypeDefByName(class string) (*TypeDef, error) {
	// the type can belong to any of the contexts
	for _, ctx := range mds.contexts {
		if td := mds.typeDefByNameAndCtx(class, ctx); td != nil {
			return td, nil // return the first match
		}
	}
	return nil, &ClassNotFoundError{Class: class}
}

func (mds *Store) typeDefByNameAndCtx(class string, ctx *types.Context) *TypeDef {
	typeDefTable := ctx.Table(md.TypeDef)
	for i := uint32(0); i < typeDefTable.RowCount(); i++ {
		var typeDef types.TypeDef
		if err := typeDef.FromRow(typeDefTable.Row(i)); err != nil {
			continue // keep searching instead of failing
		}

		if typeDef.TypeNamespace+"."+typeDef.TypeName == class {
			return &TypeDef{
				TypeDef:    typeDef,
				HasContext: HasContext{ctx},
				logger:     mds.logger,
			}
		}
	}

	return nil
}
