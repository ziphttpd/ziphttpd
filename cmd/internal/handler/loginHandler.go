package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

var logintmpl *template.Template

// loginparam はテンプレートに渡す情報です。
type loginparam struct {
	// ドキュメントグループ名
	HostName string
	// バージョン
	Version string
	// 広告URL
	AdURL string
	// リダイレクト先
	RedirectTo string
	// トークン
	Token string
	// sessionStorage/localStorage
	Storage string
}

func init() {
	tplStr := `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width">
		<meta name="description" content="document group login view">
		<title>ZipHttpd - login</title>
		<style type="text/css">
<!--
#main {
	width: 100%
}
#list {
	padding: 12px;
	mergin-left: 20px;
	min-width: 256px;
}
#list th {
	border: #C0C0C0 1px solid;
	background-color: beige;
	padding: 12px;
}
#copyright {
	padding: 12px;
	text-align: center;
	vertical-align: text-top;
}
.hostname {
	color: green;
	font-weight: bold;
}
#main .filler {
	width: 40px;
}
.login {
	text-align: right;
}
.login input {
	padding-right: 10px;
	padding-left: 10px;
}
.adspace {
	width: 512px;
	min-width: 512px;
	background-color: wheat;
}
.adspace div {
	text-align: center;
	font-size: 16pt;
	font-family: serif;
}
-->
		</style>
		</head>
	<body>
		<table id="main"><tbody>
			<tr><td>
				<p class="attention">
				表示中のドキュメントで ZipHttpd の WebAPI が実行されました。<br/>
				ローカルテキストファイルに記録されているパスワードでの認証を行います。<br/>
				これは記録されたデータへの、第三者のスクリプトによるアクセスを防ぎます。<br/>
				</p>
				<form method="POST" action="">
				host: <span class="hostname">{{.HostName}}</span><br/>
				password: <input type="password" autocomplete="current-password" name="password" required id="password"/><br/>
				<input type="submit" value="login"/>
				<input type="hidden" name="redirectto" value="{{.RedirectTo}}"/>
				</form>
			</td><td class="adspace"><iframe id="ad" width="100%" height="100%" src="" frameborder="0"></iframe>
			</td></tr>
		</tbody></table>
		<hr/>
		<div id="copyright">Powered by <a href="https://ziphttpd.com/">ZipHttpd</a>.{{.Version}}</div>
		<script>
		document.addEventListener("DOMContentLoaded", function() {
			let token = "{{.Token}}";
			window['localStorage'].removeItem('token')
			window['sessionStorage'].removeItem('token')
			window['{{.Storage}}'].setItem('token', token);
			if (token) {
				location.href = "{{.RedirectTo}}";
			} else {
				document.all.item("ad").src = "{{.AdURL}}";
			}
		});
		</script>
	</body>
</html>
`
	tmpl, err := template.New("login").Parse(tplStr)
	if err != nil {
		panic(err)
	}
	logintmpl = tmpl
}

// LoginHandler はログインに対するリクエストを処理するハンドラです。
func LoginHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	// 呼び出し元URL
	redirectTo := request.GetPostForm("redirectto")
	// パスワード
	password := request.GetPostForm("password")

	// localhost:8823
	_, port := SplitHost(request.Host())
	reqPort := port // 8823
	iPort, _ := strconv.Atoi(reqPort)
	// ポート番号からドキュメントグループ名称を取得
	hostName := param.PortMan().HostName(iPort)
	if hostName == "" {
		// 404 file not found
		ErrorHandler(writer, request, param, http.StatusNotFound)
		return
	}
	token := ""
	sec := param.SecurityMan()
	if password != "" {
		if sec.IsValid(hostName, password) {
			token = sec.Token(hostName)
		} else {
			// 第三者によってパスワードが試行されている可能性があるので、間違ったパスワードにはディレイを入れる
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
	var storage string
	if sec.UseLocalStorage(hostName) {
		storage = "localStorage"
	} else {
		storage = "sessionStorage"
	}

	// テンプレートパラメータ
	tmplParam := &loginparam{
		HostName:   hostName,
		Version:    param.Version(),
		RedirectTo: redirectTo,
		AdURL:      adURL,
		Token:      token,
		Storage:    storage,
	}
	writer.SetHeader("Content-Type", "text/html")
	// https://golang.org/pkg/html/template/ によるとコードインジェクションされないはず
	if err := writer.ParseContents(logintmpl, tmplParam); err != nil {
		param.Logger().Warnf("writer.ParseContents error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
}
