package winmd

import (
	"bytes"
	"debug/pe"
	"embed"
	"io/fs"
	"io/ioutil"
)

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
