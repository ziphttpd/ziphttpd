package model

import (
	fpath "path/filepath"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/util/pkg/zip"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

const (
	// zip 内に梱包しておく設定ファイル
	documentConfigFile = "ziphttpd/config.json"
	// ドキュメントの実体のパス
	docpathPath = json.PathJSON("path")
	// 配布サイトのホスト名(https://ホスト名/)
	docpathHost = json.PathJSON("host")
	// ドキュメント表示名
	docpathName = json.PathJSON("name")
	// euc-8 などのコンテンツのエンコード
	docpathEncoding = json.PathJSON("contentencoding")
	// index.html などの初期表示ファイル
	docpathDocRoot = json.PathJSON("docroot")
	// ドキュメントグループ名
	docpathDocGroup = json.PathJSON("docgroup")
	// 拡張子別のコンテンツタイプ
	docpathContentType = json.PathJSON("contenttype")
	// 静的ファイルを利用するか (ドキュメント開発時のデバッグ用)
	docpathUseStaticFiles = json.PathJSON("usestaticfiles")
)

// NewDocConfig は簡易なドキュメント要素を構築します。
func NewDocConfig(c common.Config, zipPath, displayname string) (json.Element, error) {
	elem := json.NewElemObject()
	var dic zip.Dictionary
	var err error

	// zip ファイルのパス
	zipPath = fpath.Clean(zipPath)
	if dic, err = zip.OpenDictionary(zipPath, false); nil != err {
		// zip として開けなかった
		return nil, err
	}
	defer dic.Close()

	// zipファイルのパスを相対化して記録
	zipRelPath, err := fpath.Rel(c.ConfigPath(), zipPath)
	if err != nil {
		zipRelPath = zipPath
	}
	elem.Put(docpathPath, json.NewElemString(zipRelPath))

	// 提供者設定を反映
	confElem, err := loadDefinedConfig(dic)
	if nil != err {
		return nil, err
	}

	// ドキュメント表示名
	if str, ok := json.QueryElemString(confElem, docpathName); ok {
		elem.Put(docpathName, str.Clone())
	} else {
		elem.Put(docpathName, json.NewElemString(displayname))
	}

	// 配布サイトのホスト名
	if str, ok := json.QueryElemString(confElem, docpathHost); ok {
		elem.Put(docpathHost, str.Clone())
	}

	// utf-8 などのコンテンツのエンコード
	if str, ok := json.QueryElemString(confElem, docpathEncoding); ok {
		elem.Put(docpathEncoding, str.Clone())
	}

	// index.html などの初期表示ファイル
	if str, ok := json.QueryElemString(confElem, docpathDocRoot); ok {
		elem.Put(docpathDocRoot, str.Clone())
	}

	// ドキュメントグループ名
	// TODO: 第三者がドキュメントグループを詐称した場合の対策を考える
	if str, ok := json.QueryElemString(confElem, docpathDocGroup); ok {
		elem.Put(docpathDocGroup, str.Clone())
	}

	// 拡張子別のコンテンツタイプ
	if obj, ok := json.QueryElemObject(confElem, docpathContentType); ok {
		elem.Put(docpathContentType, obj.Clone())
	}

	// 静的ファイルを利用するか
	if obj, ok := json.QueryElemBool(confElem, docpathUseStaticFiles); ok {
		elem.Put(docpathUseStaticFiles, obj.Clone())
	}
	return elem, nil
}

func loadDefinedConfig(dic zip.Dictionary) (json.ElemObject, error) {
	// 指定された設定があれば読み出す
	if false == dic.Contains(documentConfigFile) {
		return nil, nil
	}
	bytes, err := dic.Get(documentConfigFile)
	if nil != err {
		// エラーではあるが無視するだけ
		return nil, nil
	}
	elem, err := json.LoadFromJSONByte(bytes)
	if nil != err {
		// エラーではあるが無視するだけ
		return nil, nil
	}
	if ret, ok := elem.AsObject(); ok {
		return ret, nil
	}
	// 無視するだけ
	return nil, nil
}
