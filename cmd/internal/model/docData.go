package model

import (
	"fmt"
	"os"
	fpath "path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/util/pkg/zip"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

const (
	// ドキュメントグループが指定されていない場合のデフォルト
	defaultDocPortGroup = "common"
)

// docInst はホストしているzipを管理します。
type docInst struct {
	conf common.Config
	// 親の設定
	typeer common.ContentTypeer
	// 設定ファイルのエレメント
	element json.Element
	// 設定ファイルのパス
	conffile string
	// zipファイルのパス
	zipfile string
	// 静的ファイルパス
	staticPath string
	// ドキュメントグループ
	docGroupName common.DocGroupName
	// パス内の zip ドキュメントを特定する識別 (eg. /docid/index.html)
	docid common.DocID
	// text/xxx の Content-Encoding
	encoding string
	// 静的ファイル利用
	useStaticFiles bool
	// Zipファイル
	zipDic zip.Dictionary
	// ドキュメントルート
	docroot string
	// 拡張子 - Content-Type 辞書
	contentTypes map[string]string
	// タイトル
	title string
	// 説明
	description string
}

// JSON はJSONオブジェクトを返します。
func (d *docInst) JSON() json.ElemObject {
	elem := json.NewElemObject()
	elem.Put("conffile", json.NewElemString(d.conffile))
	elem.Put("zipfile", json.NewElemString(d.zipfile))
	elem.Put("docGroupName", json.NewElemString(d.docGroupName))
	elem.Put("docid", json.NewElemString(d.docid))
	elem.Put("encoding", json.NewElemString(d.encoding))
	elem.Put("useStaticFiles", json.NewElemBool(d.useStaticFiles))
	elem.Put("staticPath", json.NewElemString(d.staticPath))
	elem.Put("docroot", json.NewElemString(d.docroot))
	elem.Put("title", json.NewElemString(d.title))
	elem.Put("description", json.NewElemString(d.description))
	return elem
}

// OpenDocConfig はドキュメントの設定を読みだします。
func OpenDocConfig(conf common.Config, confFile, host, group, doc string) (common.DocData, error) {
	var err error
	sfilePath := fpath.Join(conf.ConfigPath(), "static", host, group, doc)
	// 生成
	dd := &docInst{
		conf:         conf,
		typeer:       conf,
		docGroupName: defaultDocPortGroup,
		conffile:     confFile,
		contentTypes: map[string]string{},
		staticPath:   sfilePath,
	}

	// ドキュメント設定ファイル
	elem, err := json.LoadFromJSONFile(confFile)
	if err != nil {
		// 定義ファイルがない
		return nil, err
	}

	if o, ok := elem.AsObject(); ok {
		if host != "" {
			o.Put("host", json.NewElemString(host))
		}
		if group != "" {
			o.Put("docgroup", json.NewElemString(group))
		}
		if doc != "" {
			o.Put("name", json.NewElemString(doc))
		}
	}

	dd.element = elem
	dd.setup()

	return dd, err
}

// ポートグループ名禁止文字の変換
func mask(str string) string {
	// TODO: 適当に禁止しているので再検討
	rarr := []rune(str)
	ret := make([]rune, 0, cap(rarr))
	for _, ch := range rarr {
		switch ch {
		case '<', '{', '[':
			ret = append(ret, '(')
		case '>', '}', ']':
			ret = append(ret, ')')
		case '?', '!', ':', ',', '%', '&', '|', '+', '/', '*', '\\', '"', '\'':
			ret = append(ret, ' ')
		default:
			ret = append(ret, ch)
		}
	}
	return string(ret)
}

// setup は docData.element の内容を読みだします。
func (d *docInst) setup() {
	// zipファイルのパス
	if pathcont, ok := json.QueryElemString(d.element, docpathPath); ok {
		d.zipfile = pathcont.Text()
	}
	if false == fpath.IsAbs(d.zipfile) {
		// 相対パスを絶対パスに変換
		d.zipfile = fpath.Join(d.conf.ConfigPath(), d.zipfile)
	}

	// ドキュメントグループ名
	if nameE, ok := json.QueryElemString(d.element, docpathDocGroup); ok {
		d.docGroupName = mask(nameE.Text())
	}

	// ドキュメント識別子
	if nameE, ok := json.QueryElemString(d.element, docpathName); ok {
		d.docid = strings.ToLower(nameE.Text())
	} else {
		// 定義がなければ拡張子を外してドキュメント識別子とする
		filename := d.zipfile
		d.docid = strings.ToLower(fpath.Base(filename[:len(filename)-len(fpath.Ext(filename))]))
	}

	// 固定の Content-Encoding
	d.encoding = "; charset=utf8"
	if encE, ok := json.QueryElemString(d.element, docpathEncoding); ok {
		d.encoding = "; charset=" + encE.Text()
	}

	// ドキュメントルート
	d.docroot = ""
	if nameD, ok := json.QueryElemString(d.element, docpathDocRoot); ok {
		d.docroot = nameD.Text()
	}

	// 拡張子 - Content-Type 辞書
	if elem, ok := json.QueryElemObject(d.element, docpathContentType); ok {
		keys := elem.Keys()
		for _, exts := range keys {
			// Content-Type
			typestr := elem.Child(exts).Text()
			for _, ext := range strings.Split(strings.ToLower(exts), ",") {
				// 拡張子
				nmlExt := strings.TrimSpace(ext)
				d.contentTypes[nmlExt] = typestr
			}
		}
	}

	// 静的ファイルを利用するか
	d.useStaticFiles = false
	if use, ok := json.QueryElemBool(d.element, docpathUseStaticFiles); ok {
		d.useStaticFiles = use.Bool()
	}

}

// Name はドキュメントの文字列を返します。
func (d *docInst) DocID() common.DocID {
	return d.docid
}

// DocRoot はドキュメントルートの文字列を返します。
func (d *docInst) DocRoot() string {
	return d.docroot
}

// DocGroup は所属しているグループを取得します。
func (d *docInst) DocGroupName() common.DocGroupName {
	return d.docGroupName
}

// Encoding は text/xxx の Content-Encoding　に設定する文字列を返します。
func (d *docInst) Encoding() string {
	return d.encoding
}

// UseStaticFiles は静的ファイルを利用するかを取得します。
func (d *docInst) UseStaticFiles() bool {
	return true
}

// ZipDic はZipファイル辞書を返します。
func (d *docInst) ZipDic() zip.Dictionary {
	if d.zipDic == nil {
		// 必要あるまでzipファイル読み込みは遅延
		var err error
		var file string
		if fpath.IsAbs(d.zipfile) {
			file = d.zipfile
		} else {
			file = fpath.Join(d.conf.DocPath(), d.zipfile)
		}
		d.zipDic, err = zip.OpenDictionary(file, false)
		if err != nil {
			return nil
		}
	}
	return d.zipDic
}

// ConfPath は設定ファイルのパスを返します。
func (d *docInst) ConfPath() string {
	return d.conffile
}

// ZipPath はzipファイルのパスを返します。
func (d *docInst) ZipPath() string {
	return d.zipfile
}

func fileList(base string) []string {
	ret := []string{}
	if f, err := os.Open(base); err == nil {
		defer f.Close()
		if fi, err := f.Stat(); err != nil || fi.IsDir() == false {
			// エラーかディレクトリ
			return ret
		}
		list, err := f.Readdir(-1)
		if err != nil {
			return ret
		}
		for _, fi := range list {
			name := fpath.Join(base, fi.Name())
			ret = append(ret, fpath.ToSlash(name))
			if fi.IsDir() {
				ret = append(ret, fileList(name)...)
			}
		}
	}
	return ret
}

// FilePaths はドキュメント内のファイルパスの一覧(zip, static)を返します。
func (d *docInst) FilePaths() []string {
	dic := map[string]int{}
	paths := d.ZipDic().FilePaths()
	for i, str := range paths {
		dic[str] = i
	}
	// 静的ファイル(static/ホスト/ドキュメントグループ/ドキュメント/パス)
	sfilePath := fpath.Join(d.staticPath, d.docGroupName, d.docid)
	list := fileList(sfilePath)
	slen := len(sfilePath)
	for i, str := range list {
		s := str[slen+1:]
		dic[s] = i
	}
	// マージ
	ret := []string{}
	for k := range dic {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

// FileInfo はファイルパスの FileInfo を返します。
func (d *docInst) FileInfo(filepath string) (common.DocFileInfo, error) {
	sfilePath := fpath.Join(d.staticPath, d.docGroupName, d.docid)
	sfile := fpath.Join(sfilePath, filepath)
	if f, err := os.Open(sfile); err == nil {
		defer f.Close()
		if fi, err := f.Stat(); err == nil {
			info := &fileinfo{
				isdir:   fi.IsDir(),
				modtime: fi.ModTime(),
				size:    strconv.FormatInt(fi.Size(), 10),
			}
			return info, nil
		}
	}
	zf := d.ZipDic().File(filepath)
	if zf == nil {
		return nil, fmt.Errorf("not found")
	}
	f := zf.File()
	fi := &fileinfo{
		isdir:   f.FileInfo().IsDir(),
		modtime: f.Modified,
		size:    strconv.FormatUint(f.UncompressedSize64, 10),
	}
	return fi, nil
}

// ContentType はファイルの拡張子から Content-Type を取得します。
func (d *docInst) ContentType(filepath string) string {
	ext := strings.ToLower(strings.TrimPrefix(fpath.Ext(filepath), "."))
	if val, ok := d.contentTypes[ext]; ok {
		return val
	}
	// 固有の設定が無かったので親設定のコンテントタイプ
	return d.typeer.ContentType(filepath)
}

// Close はドキュメントをクローズします。
func (d *docInst) Close() {
	if d.zipDic != nil {
		d.zipDic.Close()
		d.zipDic = nil
	}
}

// SetTitleInfo はタイトル情報をセットします。
func (d *docInst) SetTitleInfo(title, description string) {
	d.title = title
	d.description = description
}

// Title はタイトルを返します。
func (d *docInst) Title() string {
	return d.title
}

// Description は説明を返します。
func (d *docInst) Description() string {
	return d.description
}
