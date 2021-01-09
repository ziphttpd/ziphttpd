package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

var toptpl *template.Template

type topdoc struct {
	// ドキュメント名
	Name string
	// タイトル
	Title string
	// 注釈
	Description string
	// パス
	Path string
	// URL
	URL string
}

func (d *topdoc) JSON() json.Element {
	elem := json.NewElemObject()
	elem.Put("Name", json.NewElemString(d.Name))
	elem.Put("Title", json.NewElemString(d.Title))
	elem.Put("Description", json.NewElemString(d.Description))
	elem.Put("Path", json.NewElemString(d.Path))
	elem.Put("URL", json.NewElemString(d.URL))
	return elem
}

func (d *topdoc) String() string {
	return fmt.Sprintf("Doc: {Name:\"%s\", Title:\"%s\", Description:\"%s\", Path:\"%s\", URL:\"%s\"}", d.Name, d.Title, d.Description, d.Path, d.URL)
}

type topgroup struct {
	// ドキュメントグループ名
	Name string
	// タイトル
	Title string
	// 注釈
	Description string
	// フォルダ
	Documents []*topdoc
}

func (d *topgroup) JSON() json.Element {
	elem := json.NewElemObject()
	elem.Put("Name", json.NewElemString(d.Name))
	elem.Put("Title", json.NewElemString(d.Title))
	elem.Put("Description", json.NewElemString(d.Description))
	arr := json.NewElemArray()
	for _, doc := range d.Documents {
		arr.Append(doc.JSON())
	}
	elem.Put("Documents", arr)
	return elem
}

func (d *topgroup) String() string {
	return fmt.Sprintf("Group: {Name:\"%s\", Title:\"%s\", Description:\"%s\"}", d.Name, d.Title, d.Description)
}

type tophost struct {
	// ホスト名
	Name string
	// タイトル
	Title string
	// 注釈
	Description string
	// フォルダ
	DocumentGroups []*topgroup
}

func (d *tophost) JSON() json.Element {
	elem := json.NewElemObject()
	elem.Put("Name", json.NewElemString(d.Name))
	elem.Put("Title", json.NewElemString(d.Title))
	elem.Put("Description", json.NewElemString(d.Description))
	arr := json.NewElemArray()
	for _, doc := range d.DocumentGroups {
		arr.Append(doc.JSON())
	}
	elem.Put("DocumentGroups", arr)
	return elem
}
func (d *tophost) String() string {
	return fmt.Sprintf("Host: {Name:\"%s\", Title:\"%s\", Description:\"%s\"}", d.Name, d.Title, d.Description)
}

// dirparam はテンプレートに渡す情報です。
type topparam struct {
	// フォルダ
	DocumentHosts []*tophost
	// バージョン
	Version string
	// 広告URL
	AdURL string
}

func (d *topparam) JSON() json.Element {
	elem := json.NewElemObject()
	elem.Put("Version", json.NewElemString(d.Version))
	elem.Put("AdURL", json.NewElemString(d.AdURL))
	arr := json.NewElemArray()
	for _, doc := range d.DocumentHosts {
		arr.Append(doc.JSON())
	}
	elem.Put("DocumentHosts", arr)
	return elem
}

func init() {
	tplStr := `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width">
		<meta name="description" content="document root view">
		<title>ZipHttpd - top</title>
		<style type="text/css">
<!--
#main {
	width: 100%;
	height: 90%;
}
#scr {
	overflow-x: hidden;
	overflow-y: scroll;
}
#list {
	padding: 12px;
	margin-left: 20px;
	min-width: 256px;
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

#main .filler {
	width: 40px;
}
h1 {
	border: #C0C0C0 1px solid;
	background-color: beige;
	padding-left: 10px;
	margin-top: 2px;
	margin-bottom: 2px;
}
h2 {
	color: blue;
	font-weight: bold;
	margin-bottom: 2px;
}
h3 {
	color: green;
	font-weight: bold;
	margin-bottom: 2px;
}
.indent {
	margin-left: 2em;
}
.description {
	border: #C0C0C0 1px solid;
	background-color: beige;
	padding: 12px;
	font-size: x-small;
}
-->
		</style>
	</head>
	<body>
		<table id="main"><tbody>
		<tr><td>
			<h1>Hosted Document</h1>
			<div class="indent">
			{{range .DocumentHosts}}
				<h2>{{.Title}}</h2>
				<div class="indent">
				{{range .DocumentGroups}}
					<h3>{{.Title}}</h3>
					{{if .Description}}<div class="description">{{.Description}}</div>{{end}}
					<div class="indent">
					{{range .Documents}}
						<a href="{{.URL}}" title="{{.Path}}">{{.Title}}</a><br>
						{{if .Description}}<div class="description">{{.Description}}</div>{{end}}
					{{end}}
					</div>
				{{end}}
				</div>
			{{end}}
			</div>
		</div>
		</td><td class="adspace">
			<iframe id="ad" width="100%" height="100%" src="" frameborder="0"></iframe>
		</td></tr>
		</tbody></table>
		<hr/>
		<div id="copyright">Powered by <a href="https://ziphttpd.com/">ZipHttpd</a>.{{.Version}}</div>
		<script>
document.addEventListener("DOMContentLoaded", function() {
	document.all.item("ad").src = "{{.AdURL}}";
});
		</script>
	</body>
</html>
`
	tmpl, err := template.New("top").Parse(tplStr)
	if err != nil {
		panic(err)
	}
	toptpl = tmpl
}

// TopHandler はトップディレクトリに対するリクエストを処理するハンドラです。
func TopHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	// localhost:8823
	reqAddr, reqPort := SplitHost(request.Host())
	intPort, err := strconv.Atoi(reqPort)
	if err != nil {
		reqPort = "0"
		intPort = 0
	}
	// http://localhost
	domain := "http://" + reqAddr
	listenPort := param.ListenPort()
	logger := param.Logger()
	if intPort != listenPort {
		// トップページなのにポートが代表ポートではないのでリダイレクト
		redirectto := fmt.Sprintf("%s:%d", domain, listenPort)
		logger.Info("redirect to " + redirectto)
		//writer.Redirect(request, redirectto, http.StatusMovedPermanently)
		// 永続的(301)では、ブラウザは記憶していて永続的にリダイレクトする(Chromeで発生)
		// これは portlockin などで転送先が変わっているときにも記憶に従う
		// よって、一時的なリダイレクト(302)を使用する
		writer.Redirect(request, redirectto, http.StatusFound)
		return
	}
	// テンプレートパラメータ
	tmpParam := &topparam{
		DocumentHosts: []*tophost{},
		Version:       param.Version(),
		AdURL:         adURL,
	}
	conf := param.Config()
	for _, hostName := range conf.HostNames() {
		// ホスト
		docHost := conf.DocHost(hostName)
		if docHost == nil {
			// ホストがない
			continue
		}
		if len(docHost.Ids()) == 0 {
			// ドキュメントグループが無い
			continue
		}
		hostTitle := docHost.Title()
		if hostTitle == "" {
			hostTitle = docHost.Name()
		}
		tmpHost := &tophost{
			Name:           docHost.Name(),
			Title:          hostTitle,
			Description:    docHost.Description(),
			DocumentGroups: []*topgroup{},
		}
		logger.Info(tmpHost.String())
		tmpParam.DocumentHosts = append(tmpParam.DocumentHosts, tmpHost)
		for _, docGroupName := range docHost.Ids() {
			// ドキュメントグループ
			docGroup := docHost.Get(docGroupName)
			if docGroup == nil {
				// ドキュメントグループが無い
				continue
			}
			if len(docGroup.Ids()) == 0 {
				// ドキュメントが無い
				continue
			}

			var baseurl *url.URL
			if docGroupName == reqPort {
				baseurl, _ = url.Parse(domain)
			} else {
				baseurl, _ = url.Parse(domain + ":" + docHost.Port())
			}
			// ドキュメントグループ
			groupTitle := docGroup.Title()
			if groupTitle == "" {
				groupTitle = docGroup.Name()
			}
			tmpParamPortGroup := &topgroup{
				Name:        docGroup.Name(),
				Title:       groupTitle,
				Description: docGroup.Description(),
				Documents:   []*topdoc{},
			}
			logger.Info(tmpParamPortGroup.String())
			for _, docid := range docGroup.Ids() {
				// zip ドキュメント
				docData := docGroup.Get(docid)
				// パスを合成 (トラバーサル予防)
				ustr := hostName + "/" + docGroupName + "/" + docid + "/" + docData.DocRoot()
				requrl, _ := url.Parse(ustr)
				req := baseurl.ResolveReference(requrl).String()
				docTitle := docData.Title()
				if docTitle == "" {
					docTitle = docid
				}
				td := &topdoc{
					URL:         req,
					Name:        docid,
					Title:       docTitle,
					Description: docData.Description(),
					Path:        docData.ZipPath(),
				}
				logger.Info(td.String())
				tmpParamPortGroup.Documents = append(tmpParamPortGroup.Documents, td)
			}
			tmpHost.DocumentGroups = append(tmpHost.DocumentGroups, tmpParamPortGroup)
		}
	}

	hosts := tmpParam.JSON()
	logger.Info(json.ToJSON(hosts, true))

	writer.SetHeader("Content-Type", "text/html")
	// https://golang.org/pkg/html/template/ によるとコードインジェクションされないはず
	err = writer.ParseContents(toptpl, tmpParam)
	if err != nil {
		param.Logger().Warnf("writer.ParseContents error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
}
