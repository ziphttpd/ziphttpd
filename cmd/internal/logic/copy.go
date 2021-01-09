package logic

import (
	"io"
	"io/ioutil"
	"os"
	fpath "path/filepath"
)

// copyFile は指定のファイルをコピーします。
func copyFile(dst, src string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	return err
}

// copyFiles はディレクトリ内のファイルをコピーします。
func copyFiles(dstDir, srcDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		src := fpath.Join(srcDir, file.Name())
		dst := fpath.Join(dstDir, file.Name())
		if err := copyFile(dst, src); err != nil {
			return err
		}
	}
	return nil
}
