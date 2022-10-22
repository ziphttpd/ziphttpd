package logic

import (
	"fmt"
	"os"
	fpath "path/filepath"
	"strings"

	"github.com/xorvercom/util/pkg/fileutil"
)

// listFilesByExt は特定の拡張子を持つファイルの一覧を返します
func listFilesByExt(folder, ext string) []string {
	res := []string{}
	for _, filename := range fileutil.FilesList(folder) {
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
	return os.WriteFile(filename, []byte(savedata), 0655)
}

// loadString は文字列を読み込みます。
func loadString(filename string) (str string, err error) {
	bytes, err := os.ReadFile(filename)
	if err == nil {
		str = string(bytes)
	}
	return
}
