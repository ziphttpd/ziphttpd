package model

import "time"

type fileinfo struct {
	isdir   bool
	modtime time.Time
	size    string
}

func (f *fileinfo) IsDir() bool {
	return f.isdir
}

func (f *fileinfo) ModTime() time.Time {
	return f.modtime
}

func (f *fileinfo) Size() string {
	return f.size
}
