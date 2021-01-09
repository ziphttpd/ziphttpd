package handler

import (
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// 広告用URL
const adURL = "/ad"

var adByte []byte

func init() {
	htmlStr := `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width">
		<meta name="description" content="document group login view">
		<title>ZipHttpd - ad</title>
		<style type="text/css">
<!--
div {
	text-align: center;
	font-size: 16pt;
	font-family: serif;
	background-color: wheat;
}
-->
		</style>
	</head>
	<body>
		<div>This space is reserved for<br/>sponsored &amp; advertising.</div>
	</body>
</html>
`
	adByte = []byte(htmlStr)
}

// AdParam はテンプレートパラメータです。
type AdParam interface {
	Logger() common.Logger
	Server() common.Server
}

// AdHandler は広告領域に対するリクエストを処理するハンドラです。
func AdHandler(writer common.ResponseProxy, request common.RequestProxy, param AdParam) {
	writer.SetHeader("Content-Type", "text/html")
	_, err := writer.WriteContentsByte(adByte)
	if err != nil {
		param.Logger().Warnf("writer.WriteContentsByte error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
}
