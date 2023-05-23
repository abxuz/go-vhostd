package fs

import (
	"io/fs"
	"net/http"
	"strings"
)

type NoAutoIndexFileSystem struct {
	http.FileSystem
}

func (f *NoAutoIndexFileSystem) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if !stat.IsDir() {
		return file, nil
	}

	file.Close()

	// check index.html
	index := strings.TrimSuffix(name, "/") + "/index.html"
	file, err = f.FileSystem.Open(index)
	if err != nil {
		return file, err
	}

	stat, err = file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if !stat.IsDir() {
		return file, nil
	}
	return nil, fs.ErrPermission
}
