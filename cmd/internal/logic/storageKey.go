package logic

import (
	"encoding/hex"
	fpath "path/filepath"
	"strings"
)

// // key2filename はキーからファイル名を変換します（泥縄）
// func key2filename_depl(key string) string {
// 	// パスとして使用できない(かもしれない)文字をエスケープ
// 	key = url.PathEscape(key)
// 	// $ の対策 (環境変数まわり)
// 	key = strings.Replace(key, "$", "%24", -1)
// 	// `` の対策 (環境変数まわり)
// 	//	key = strings.Replace(key, "`", "%60", -1)
// 	// .. といったディレクトリトラバーサルの対策
// 	key = strings.Replace(key, ".", "%2E", -1)
// 	// ドライブレター対策
// 	key = strings.Replace(key, ":", "%3A", -1)
// 	// homeディレクトリ対策
// 	key = strings.Replace(key, "~", "%7E", -1)
// 	// AUX とか使用できないファイル名の対策
// 	key = "_" + key
// 	return key
// }

// // filename2key はファイル名をキーに変換します
// func filename2key_depl(filename string) string {
// 	base := fpath.Base(filename)
// 	base = strings.Split(base, ".")[0]
// 	ret, _ := url.PathUnescape(base)
// 	ret = strings.TrimPrefix(ret, "_")
// 	return ret
// }

// key2filename はキーを拡張子なしファイル名に変換します
func key2filename(key string) string {
	// 圧縮すると視認性もソートも滅茶苦茶になるので十六進変換で
	// キーに長い名前を使う方が悪い
	ret := hex.EncodeToString([]byte(key))
	return ret
}

// filename2key はファイル名をキーに変換します
func filename2key(filename string) string {
	base := fpath.Base(filename)
	base = strings.Split(base, ".")[0]
	bytes, _ := hex.DecodeString(base)
	ret := string(bytes)
	return ret
}
