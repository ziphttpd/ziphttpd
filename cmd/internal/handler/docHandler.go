package handler

import (
	"io"
	"net/http"
	"net/url"
	"os"
	fpath "path/filepath"
	"strings"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// DocHandler はドキュメントに対するリクエストを処理するハンドラです。
func DocHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	// localhost:8823
	// 要求されたポート
	_, reqPort := SplitHost(request.Host())

	// 要求されたドキュメントをホストしているポート
	docport := param.DocHost().Port()
	if reqPort != docport {
		// ポートが合っていないのでリダイレクト
		// パスを合成
		requrl, _ := url.Parse(request.RequestURI())
		baseurl, _ := url.Parse("http://localhost:" + docport)
		redirectto := baseurl.ResolveReference(requrl).String()
		param.Logger().Info("redirect to " + redirectto)
		//writer.Redirect(request, redirectto, http.StatusMovedPermanently)
		// 永続的(301)にすると、ブラウザは記憶していて永続的にリダイレクトする(Chromeで発生)
		// これは portlockin などで転送先が変わっているときにも記憶に従ってしまう
		// よって、一時的なリダイレクト(302)に変更して記憶しないようにする
		writer.Redirect(request, redirectto, http.StatusFound)
		return
	}

	docHostName := param.DocHost().Name()
	docGroupName := param.DocGroup().Name()
	doc := param.DocData()
	docID := doc.DocID()
	// ファイルのパス
	filepath := strings.Join(param.Paths()[4:], "/")
	var reader io.ReadCloser
	reader = nil
	if doc.UseStaticFiles() {
		// 静的ファイル(static/ホスト/ドキュメントグループ/ドキュメント/パス)を優先
		sfilePath := fpath.Join(param.ConfigPath(), "static", docHostName, docGroupName, docID)
		sfile, _ := fpath.Abs(fpath.Join(sfilePath, filepath))
		if strings.HasPrefix(sfile, sfilePath) {
			// 静的ファイルアクセスではドキュメントからトラバーサルされていないことが条件
			if sfile, err := os.Open(sfile); err == nil {
				reader = sfile
			}
		}
	}
	if reader == nil {
		// 通常にzipから取得
		zipdic := doc.ZipDic()
		if !zipdic.Contains(filepath) {
			ErrorHandler(writer, request, param, http.StatusNotFound)
			return
		}
		// ファイルの中身を用意
		reader, _ = zipdic.GetReader(filepath)
	}
	// 予約クローズ
	defer reader.Close()

	// レスポンスヘッダ
	ct := doc.ContentType(filepath)
	if ct == "" {
		// 未定義の拡張子はダウンロードさせる
		ct = "application/octet-stream"
	}
	if strings.HasPrefix(ct, "text/") {
		// テキスト
		ct += doc.Encoding()
	}
	writer.SetHeader("Content-Type", ct)

	logger := param.Logger()
	logger.Infof("    type:%s", ct)

	// ヘッダの決定
	writer.WriteHeader(http.StatusOK)

	// サーバとして確保するメモリを削減するため、中身を小分けで送信する
	_, err := writer.WriteContents(reader)
	if err != nil {
		logger.Warnf("io.Copy error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
	logger.Infof("[done]    type:%s", ct)
}
