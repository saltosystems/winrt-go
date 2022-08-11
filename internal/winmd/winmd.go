package winmd

import (
	"bytes"
	"debug/pe"
	"embed"
	"io/fs"
	"io/ioutil"

	"github.com/tdakkota/win32metadata/types"
)

// Custom Attributes
const (
	AttributeTypeGUID                 = "Windows.Foundation.Metadata.GuidAttribute"
	AttributeTypeExclusiveTo          = "Windows.Foundation.Metadata.ExclusiveToAttribute"
	AttributeTypeStaticAttribute      = "Windows.Foundation.Metadata.StaticAttribute"
	AttributeTypeActivatableAttribute = "Windows.Foundation.Metadata.ActivatableAttribute"
	AttributeTypeDefaultAttribute     = "Windows.Foundation.Metadata.DefaultAttribute"
	AttributeTypeOverloadAttribute    = "Windows.Foundation.Metadata.OverloadAttribute"
)

// HasContext is a helper struct that holds the original context of a metadata element.
type HasContext struct {
	originalCtx *types.Context
}

// Ctx return the original context of the element.
func (hctx *HasContext) Ctx() *types.Context {
	return hctx.originalCtx
}

//go:embed metadata/*.winmd
var files embed.FS

// allFiles returns all winmd files embedded in the binary.
func allFiles() ([]fs.DirEntry, error) {
	return files.ReadDir("metadata")
}

// open reads the given file and returns a pe.File instance.
// The user should close the returned instance once he is done working with it.
func open(path string) (*pe.File, error) {
	f, err := files.Open("metadata/" + path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	pef, err := pe.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return pef, nil
}
