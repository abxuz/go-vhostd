package fs

import (
	"io/fs"
	"net/http"
)

type NoAutoIndexFileSystem struct {
	http.FileSystem
}

type NoAutoIndexFile struct {
	http.File
}

func (f *NoAutoIndexFileSystem) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return &NoAutoIndexFile{File: file}, nil
}

func (f *NoAutoIndexFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fs.ErrPermission
}
