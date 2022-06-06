package winmd

import (
	"bytes"
	"debug/pe"
	"embed"
	"io/fs"
	"io/ioutil"
)

//go:embed *.winmd
var files embed.FS

// AllFiles returns all winmd files embedded in the binary.
func AllFiles() ([]fs.DirEntry, error) {
	return files.ReadDir(".")
}

// Open reads the given file and returns a pe.File instance.
// The user should close the returned instance once he is done working with it.
func Open(path string) (*pe.File, error) {
	f, err := files.Open(path)
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
