package config

import (
	fpath "path/filepath"

	"github.com/ziphttpd/zhsig/pkg/zhsig"
)

// FileMan はファイル名管理
type FileMan struct {
	baseDir string
}

// NewFileMan は baseDir を起点としたファイル名管理を生成します
func NewFileMan(baseDir string) *FileMan {
	return &FileMan{baseDir: baseDir}
}

// DocConfig はドキュメント設定ファイル(json)のパスを生成します
//
// TODO: 自動ダウンロードされた store 以下のドキュメントの設定ファイルをどうするのか非常に困っている
// 将来的には変更することにした暫定対応
func (f *FileMan) DocConfig(host, group, docid string) string {
	var ret string
	if host == "" {
		host = "common"
		// docs/{group}/{docid}.json
		ret = fpath.Join(f.baseDir, defaultDocument, group, docid+extConf)
	} else {
		// store/{host}/{group}/{docid}.json
		h := zhsig.NewHost(f.baseDir, host)
		ret = fpath.Join(h.StorePath(), group, docid+extConf)
	}
	return ret
}
