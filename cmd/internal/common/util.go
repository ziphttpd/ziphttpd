package common

import (
	"os"
	fpath "path/filepath"
)

// FileExists はファイルが存在するか判定します。
func FileExists(filename string) bool {
	f, err := os.Stat(filename)
	return err == nil && false == f.IsDir()
}

// DirExists はディレクトリが存在するか判定します。
func DirExists(filename string) bool {
	f, err := os.Stat(filename)
	return err == nil && f.IsDir()
}

// BaseName はファイルの拡張子以外を返します
func BaseName(filename string) string {
	_, filename = fpath.Split(filename)
	basename := fpath.Base(filename[:len(filename)-len(fpath.Ext(filename))])
	return basename
}
