package handler

import (
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// FilesHandler はディレクトリに対するファイル取得リクエストを処理するハンドラです。
func FilesHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	// logger
	log := param.Logger()
	defer func() {
		if err := recover(); err != nil {
			log.Infof("Runtime Error:", err)
		}
	}()
	conf := param.Config()
	// TODO: リクエストURLパスと、内部的なパスがごっちゃになってるので、もっと統一的に扱う
	// リクエストのパス (/docGroup/docname/zip内のパス/)
	urlpath, _ := url.QueryUnescape(request.URLString())
	log.Infof("urlpath: %s", urlpath)
	paths := param.Paths()
	pathLen := len(paths)
	//docHost := param.DocHost()
	//docGroup := param.DocGroup()
	//doc := param.DocData()
	result := json.NewElemObject()
	dirArr := []string{}
	fileArr := []string{}
	if paths[pathLen-1] == "" {
		// 最後が / で終わっていた
		pathLen -= 1
	}
	switch pathLen {
	case 1, 2:
		// http://localhost:8823/files
		//  -> ホストの一覧を出力
		for _, name := range conf.HostNames() {
			dirArr = append(dirArr, name)
		}
		break
	case 3:
		// http://localhost:8823/files/hostname
		//  -> ドキュメントグループの一覧を出力
		docHost := conf.DocHost(paths[2])
		for _, name := range docHost.Ids() {
			dirArr = append(dirArr, name)
		}
		break
	case 4:
		// http://localhost:8823/files/hostname/groupname
		//  -> ドキュメントの一覧を出力
		docHost := conf.DocHost(paths[2])
		docGroup := docHost.Get(paths[3])
		for _, name := range docGroup.Ids() {
			dirArr = append(dirArr, path.Base(name))
		}
		break
	default:
		// http://localhost:8823/files/hostname/groupname/document
		//  -> ドキュメント内のファイル一覧を出力
		// url上の親 (/docname/zip内のパス)
		docHost := conf.DocHost(paths[2])
		docGroup := docHost.Get(paths[3])
		doc := docGroup.Get(paths[4])
		hostName := docHost.Name()
		docGroupName := docGroup.Name()
		docName := doc.DocID()
		//baseurlpath := path.Dir(urlpath)
		baseDirPath := strings.Join(paths[5:pathLen], "/")
		docPath := "/" + hostName + "/" + docGroupName + "/" + docName
		baseFullPath := path.Join(docPath, baseDirPath)
		log.Infof("hostName: %s", hostName)
		log.Infof("docGroupName: %s", docGroupName)
		log.Infof("docName: %s", docName)
		log.Infof("targetFullPath: %s", baseFullPath)
		dirSet := map[string]bool{}
		fileSet := map[string]bool{}
		for _, filepath := range doc.FilePaths() {
			fileFullPath := path.Join(docPath, filepath)
			log.Infof("fullpath: %s", fileFullPath)
			if false == strings.HasPrefix(fileFullPath, baseFullPath+"/") { //nolint:gosimple
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
			baseDir := path.Dir(fileFullPath)
			log.Infof("baseDir: %s", baseDir)

			if baseFullPath != baseDir {
				// 直下ではなくサブディレクトリのファイル情報だった
				continue
			}

			if isDir {
				name := path.Base(filepath)
				if _, ok := dirSet[name]; false == ok {
					dirArr = append(dirArr, name)
					dirSet[name] = true
				}
			} else {
				name := path.Base(filepath)
				if _, ok := fileSet[name]; false == ok {
					fileArr = append(fileArr, path.Base(name))
					fileSet[name] = true
				}
			}
		}
		break
	}
	// ディレクトリ
	sort.Sort(sort.StringSlice(dirArr))
	dirs := json.NewElemArray()
	result.Put("dirs", dirs)
	for _, dir := range dirArr {
		dirs.Append(json.NewElemString(dir))
	}
	// ファイル
	sort.Sort(sort.StringSlice(fileArr))
	files := json.NewElemArray()
	result.Put("files", files)
	for _, file := range fileArr {
		files.Append(json.NewElemString(file))
	}

	resultJson := result.Text()
	log.Infof("result: %s", resultJson)

	// 終了
	writer.SetHeader("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.WriteContentsByte([]byte(resultJson))
}
