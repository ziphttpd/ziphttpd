package logic

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	fpath "path/filepath"
	"sort"
	"strings"
)

// listFiles は全ファイルの一覧を返します
func listFiles(folder string) []string {
	res := []string{}
	if files, err := ioutil.ReadDir(folder); err == nil {
		for _, file := range files {
			res = append(res, file.Name())
		}
	}
	sort.Sort(sort.StringSlice(res))
	return res
}

// listFilesByExt は特定の拡張子を持つファイルの一覧を返します
func listFilesByExt(folder, ext string) []string {
	res := []string{}
	for _, filename := range listFiles(folder) {
		if strings.HasSuffix(filename, ext) {
			res = append(res, filename)
		}
	}
	return res
}

// removeFilesByExt は特定の拡張子のファイルを抹消します。
func removeFilesByExt(folder, ext string) error {
	errorFiles := []string{}
	for _, filename := range listFilesByExt(folder, ext) {
		srcName := fpath.Join(folder, filename)
		err := os.Remove(srcName)
		if err != nil {
			// エラーは一覧化する
			errorFiles = append(errorFiles, filename)
		}
	}
	if len(errorFiles) != 0 {
		// TODO: エラー処理再検討
		return fmt.Errorf("can't remove [%s]", strings.Join(errorFiles, ","))
	}
	return nil
}

func baseName(filename, ext string) string {
	return filename[:len(filename)-len(ext)]
}

// changeExt はフォルダにあるファイルのold拡張子をnew拡張子に置換します。
func changeExt(folder, srcExt, dstExt string) error {
	errorFiles := []string{}
	for _, filename := range listFilesByExt(folder, srcExt) {
		srcName := fpath.Join(folder, filename)
		dstName := baseName(srcName, srcExt) + dstExt
		err := os.Rename(srcName, dstName)
		if err != nil {
			// エラーは一覧化する
			errorFiles = append(errorFiles, filename)
		}
	}
	if len(errorFiles) != 0 {
		// TODO: エラー処理再検討
		return fmt.Errorf("can't rename [%s]", strings.Join(errorFiles, ","))
	}
	return nil
}

// saveString は文字列を保存します。
func saveString(filename, savedata string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(savedata)
	if err != nil {
		return err
	}
	return nil
}

// loadString は文字列を読み込みます。
func loadString(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func deleteFile(filename string) error {
	return os.Remove(filename)
}

// tempSpace は一時領域を作成し、runner を実行した後に一時領域を削除します。
func tempSpace(runner func(tempdir string) error) (err error) {
	// 一時領域名を乱数で生成
	randBtres := make([]byte, 8)
	rand.Read(randBtres)
	tempdir := fpath.Join(os.TempDir(), "tempspace_"+hex.EncodeToString(randBtres))

	// 一時領域を作成
	err = os.MkdirAll(tempdir, 0600)
	if err != nil {
		return err
	}
	// 一時領域を削除
	defer os.RemoveAll(tempdir)
	// パニックをキャッチ
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("Run() return %+v", r)
		}
	}()

	// runner を実行
	err = runner(tempdir)

	return err
}
