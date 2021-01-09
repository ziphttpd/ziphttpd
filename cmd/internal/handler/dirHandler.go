package handler

import (
	"html/template"
	"net/url"
	"path"
	"strings"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// fileEntry は圧縮されているファイルの情報です
type fileEntry struct {
	Name string
	Size string
	Date string
}

// dirParm はテンプレートに渡す情報です。
type dirParm struct {
	// タイトル
	Title string
	// フォルダ
	Dirs []*fileEntry
	// ファイル
	Files []*fileEntry
	// バージョン
	Version string
	// 広告URL
	AdURL string
}

var dirTmplate *template.Template

func init() {
	tplStr := `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width">
		<meta name="description" content="directory view">
		<title>{{.Title}}</title>
		<style type="text/css">
<!--
#main {
	width: 100%;
	height: 100%;
}
#scr {
	overflow-x: hidden;
//	overflow-y: scroll;
}
#list {
	padding: 12px;
	mergin-left: 20px;
	min-width: 256px;
}
#list thead, tbody {
	display: block;
	width: 100%;
}
#list tbody {
    overflow-y: scroll;
    height: 500px;
}
#list tr {
	width: 100%;
}
#list thead th {
	border: #C0C0C0 1px solid;
	background-color: beige;
	padding: 12px;
}
#copyright {
	padding: 12px;
	text-align: center;
	vertical-align: text-top;
}
.linkspace {
//	padding-left: 40px;
//	font-family: monospace;
}
.adspace {
	width: 512px;
	min-width: 512px;
	background-color: wheat;
	vertical-align: top;
}
#ad {
	width: 100%;
}

#list td:nth-child(3n-1) {
	background-color: beige;
}
.sizespace {
	font-size: 8pt;
	font-family: monospace;
}
.datespace {
	font-size: 8pt;
	font-family: monospace;
}
-->
		</style>
	</head>
	<body>
		<table id="main"><tbody>
			<tr><td>
				<div id="scr">
					<table id="list">
					<thead><tr><th>Name</th><th>Size</th><th>Date</th></tr></thead>
					<tbody>
						{{range .Dirs}}
							<tr><td class="linkspace"><a href="./{{.Name}}">{{.Name}}</a></td><td class="sizespace">{{.Size}}</td><td class="datespace">{{.Date}}</td></tr>
						{{end}}
						{{range .Files}}
							<tr><td class="linkspace"><a href="./{{.Name}}">{{.Name}}</a></td><td class="sizespace">{{.Size}}</td><td class="datespace">{{.Date}}</td></tr>
						{{end}}
					</tbody>
					</table>
				</div>
			</td><td class="adspace">
				<iframe id="ad" width="100%" height="100%" src="" frameborder="0"></iframe>
			</td></tr>
			<tr><td colspan="2">
				<hr/>
				<div id="copyright">Powered by <a href="https://ziphttpd.com/">ZipHttpd</a>.{{.Version}}</div>
			</td></tr>
		</tbody></table>
		<script>
document.addEventListener("DOMContentLoaded", function() {
	document.all.item("ad").src = "{{.AdURL}}";
});
		</script>
	</body>
</html>
`
	tmpl, err := template.New("dir").Parse(tplStr)
	if err != nil {
		panic(err)
	}
	dirTmplate = tmpl
}

// DirHandler はディレクトリに対するリクエストを処理するハンドラです。
func DirHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	// logger
	log := param.Logger()
	// リクエストのパス (/docGroup/docname/zip内のパス/)
	urlpath, _ := url.QueryUnescape(request.URLString())
	log.Infof("urlpath: %s", urlpath)
	// テンプレートパラメータ
	tmplParam := &dirParm{
		Title:   urlpath,
		Dirs:    []*fileEntry{},
		Files:   []*fileEntry{},
		Version: param.Version(),
		AdURL:   adURL,
	}
	// TODO : パスの編集ロジックを整理
	// zip のドキュメント名
	docHost := param.DocHost()
	docGroup := param.DocGroup()
	doc := param.DocData()
	// 親ディレクトリへのリンク
	tmplParam.Dirs = append(tmplParam.Dirs, &fileEntry{Name: "../", Size: "", Date: ""})

	if docGroup == nil {
		// ■ グループのリスト
		dirSet := map[string]bool{}
		for _, hostName := range docHost.Ids() {
			name := path.Base(hostName) + "/"
			if _, ok := dirSet[name]; false == ok {
				entry := &fileEntry{Name: name, Size: "", Date: ""}
				tmplParam.Dirs = append(tmplParam.Dirs, entry)
				dirSet[name] = true
			}
		}
	} else {
		if doc == nil {
			// ■ ドキュメントのリスト
			dirSet := map[string]bool{}
			for _, filepath := range docGroup.Ids() {
				name := path.Base(filepath) + "/"
				if _, ok := dirSet[name]; false == ok {
					entry := &fileEntry{Name: name, Size: "", Date: ""}
					tmplParam.Dirs = append(tmplParam.Dirs, entry)
					dirSet[name] = true
				}
			}
		} else {
			// ■ ファイルのリスト
			// url上の親 (/docname/zip内のパス)
			hostName := docHost.Name()
			docGroupName := docGroup.Name()
			docname := doc.DocID()
			baseurlpath := path.Dir(urlpath)
			//zipDic := doc.ZipDic()
			dirSet := map[string]bool{}
			for _, filepath := range doc.FilePaths() {
				fullpath := "/" + hostName + "/" + docGroupName + "/" + docname + "/" + filepath
				if false == strings.HasPrefix(fullpath, urlpath) { //nolint:gosimple
					// "/docname/zip内のパス/" で始まらない、つまり他のディレクトリだった
					continue
				}

				// ディレクトリか？
				f, err := doc.FileInfo(filepath)
				if err != nil {
					// 存在しないファイル
					continue
				}
				isDir := f.IsDir()

				// ファイル情報の親ディレクトリ
				var basedir string
				if isDir {
					// "/basedir/dirname/" -> "/basedir"
					basedir = path.Dir(path.Dir(fullpath))
				} else {
					// "/basedir/filename" -> "/basedir"
					basedir = path.Dir(fullpath)
					// 途中ディレクトリのzipエントリが省略される問題の対策
					// urlpath / xxxx / ... の xxxx を Dir として登録
					urllen := len(urlpath)
					if urllen < len(basedir) {
						name := strings.Split(basedir[urllen:], "/")[0] + "/"
						if _, ok := dirSet[name]; false == ok {
							entry := &fileEntry{Name: name, Size: "", Date: ""}
							tmplParam.Dirs = append(tmplParam.Dirs, entry)
							dirSet[name] = true
						}
					}
				}

				if baseurlpath != basedir {
					// 直下ではなくサブディレクトリのファイル情報だった
					continue
				}

				if isDir {
					name := path.Base(filepath) + "/"
					if _, ok := dirSet[name]; false == ok {
						entry := &fileEntry{Name: name, Size: "", Date: ""}
						tmplParam.Dirs = append(tmplParam.Dirs, entry)
						dirSet[name] = true
					}
				} else {
					// 表示情報
					dateStr := f.ModTime().String()
					name := path.Base(filepath)
					entry := &fileEntry{Name: name, Size: f.Size(), Date: dateStr}
					tmplParam.Files = append(tmplParam.Files, entry)
				}
			}
		}
	}
	writer.SetHeader("Content-Type", "text/html")
	// https://golang.org/pkg/html/template/ によるとコードインジェクションされないはず
	err := writer.ParseContents(dirTmplate, tmplParam)
	if err != nil {
		param.Logger().Warnf("writer.ParseContents error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
}
