package logic

import (
	azip "archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	fpath "path/filepath"

	"github.com/xorvercom/util/pkg/json"
)

const (
	extData = ".txt"
	extTemp = ".temp"
)

// addData は配列として渡ってきたオブジェクトのキーごとの文字列値を一時ファイルとして保存します
func addData(folder string, items json.ElemObject) ([]interface{}, error) {
	// .temp ファイルが残っていたら消去
	removeFilesByExt(folder, extTemp)

	keys := make([]interface{}, 0)
	for _, key := range items.Keys() {
		keys = append(keys, key)
		var item json.ElemString
		var ok bool
		if item, ok = items.Child(key).AsString(); false == ok {
			// 文字列以外の値が挿入されていたのでエラー
			return keys, fmt.Errorf("err key=%s is not string", key)
		}

		// 格納先の基本名称
		basename := key2filename(key)
		filename := fpath.Join(folder, basename+extTemp)

		// 値
		value := item.Text()

		// 保存
		err := saveString(filename, value)
		if err != nil {
			return keys, err
		}
	}
	return keys, nil
}

// commitData は一時ファイルをデータファイルとしてコミットします
func commitData(folder string) error {
	return changeExt(folder, extTemp, extData)
}

// loadItemValue は特定キーに対する値を返します。
func loadItemValue(folder, key string) json.ElemString {
	filename := key2filename(key) + extData
	value := ""
	if res, err := loadString(fpath.Join(folder, filename)); err == nil {
		value = res
	}
	return json.NewElemString(value)
}

func deleteItem(folder, key string) error {
	filename := key2filename(key) + extData
	return deleteFile(fpath.Join(folder, filename))
}

// listKeys はキーの一覧を返します。
func listKeys(folder string) json.ElemArray {
	arr := json.NewElemArray()
	for _, filename := range listFilesByExt(folder, extData) {
		key := filename2key(baseName(filename, extData))
		arr.Append(json.NewElemString(key))
	}
	return arr
}

// backupData は全データファイルをzip化したリーダーを返します。
// 実ファイル構成を意識しないでバックアップできるようにするための試験実装。
// TODO: 未使用
func backupData(folder string) *bytes.Buffer {
	// TODO: エラーハンドリング
	buffer := &bytes.Buffer{}
	zip := azip.NewWriter(buffer)
	for _, filename := range listFilesByExt(folder, extData) {
		if reader, err := os.Open(fpath.Join(folder, filename)); err == nil {
			if ent, err := zip.Create(filename); err == nil {
				if bytes, err := ioutil.ReadAll(reader); err == nil {
					ent.Write(bytes)
					zip.Flush()
				}
			}
		}
	}
	zip.Close()
	return buffer
}
