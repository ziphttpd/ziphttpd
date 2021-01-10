package logic

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xorvercom/util/pkg/json"
)

// backgroundLogic は窓口となるバックグラウンド処理です。
func backgroundLogic(a *api) {
	for {
		var param *apiParam
		select {
		case <-a.kick:
			// API がリクエストされたので、キューから逐次実行
			for param = a.pop(); param != nil; param = a.pop() {
				// 実行中に追加されたチャネルへのキックをクリア
				chk := true
				for chk == true {
					select {
					case <-a.kick:
					default:
						chk = false
					}
				}
				if a.terminated {
					param.result = json.NewElemNull()
					param.done <- -1
				} else {
					// APIのロジックを実行
					execLogic(a, param)
				}
			}

		case <-a.done:
			// 即時全終了
			return
		}
	}
}

// APIのロジックを実行
func execLogic(a *api, param *apiParam) bool {
	log := a.config.Logger()
	// パラメータをJSONオブジェクトに
	var jsonObj json.ElemObject
	var ok bool
	if jsonObj, ok = param.elem.AsObject(); false == ok {
		// JSONオブジェクトでなかったのでエラー
		a.sendError(param, "not object")
		param.result = json.NewElemNull()
		param.done <- -1
		return false
	}

	// APIバージョンチェック
	var version json.ElemString
	if version, ok = jsonObj.Child("version").AsString(); false == ok {
		// バージョン指定が異常なのでエラー
		a.sendError(param, "invalid version")
		param.result = json.NewElemNull()
		param.done <- -1
		return false
	}
	// 現在は ver.1
	if version.Text() != "1" {
		// バージョン指定が異常なのでエラー
		a.sendError(param, "unknown version")
		param.result = json.NewElemNull()
		param.done <- -1
		return false
	}

	// apiメソッド
	var apiStr json.ElemString
	if apiStr, ok = jsonObj.Child("api").AsString(); false == ok {
		// APIが指定されていないのでエラー
		a.sendError(param, "no api")
		param.result = json.NewElemNull()
		param.done <- -1
		return false
	}

	// データ記録フォルダ
	storagePath := a.storagePath
	if nameElem, ok := jsonObj.Child("name").AsString(); ok {
		// name 別の格納先
		storagePath = filepath.Join(storagePath, key2filename(nameElem.Text()))
	}
	// フォルダがない場合には作る
	os.MkdirAll(storagePath, 0755)

	// API 別の処理
	apiMethod := strings.ToLower(apiStr.Text())
	a.sendMessage(apiMethod, param, "start")
	log.Infof(apiMethod)

	switch apiMethod {
	case "noop":
		// ログ
		log.Infof("%s: done", apiMethod)

		// 要求終了を通知
		param.result = json.NewElemNull()
		param.done <- 0
		return false

	case "list":
		res := listKeys(storagePath)
		keys := make([]interface{}, 0)
		for i := 0; i < res.Size(); i++ {
			if key, ok := res.Child(i).AsString(); ok {
				keys = append(keys, key.String())
			}
		}
		// イベント通知
		a.sendArray(apiMethod, param, keys)
		// ログ
		log.Infof("%s: done %+v", apiMethod, keys)

		// 要求終了を通知
		param.result = res
		param.done <- 0
		return false

	case "write":
		var items json.ElemObject
		if items, ok = jsonObj.Child("items").AsObject(); false == ok {
			// items がキーと値のオブジェクトでなかったのでエラー
			a.sendError(param, "write items must object")
			param.result = json.NewElemNull()
			param.done <- -1
			return false
		}

		// 一時的に .temp に書き込み
		keys, err := addData(storagePath, items)
		if err != nil {
			a.sendError(param, "dont write")
			param.result = json.NewElemNull()
			param.done <- -1
			return false
		}
		// .temp を .txt に置き換え
		err = commitData(storagePath)
		if err != nil {
			a.sendError(param, "dont rename")
			param.result = json.NewElemNull()
			param.done <- -1
			return false
		}
		// イベント通知
		a.sendArray(apiMethod, param, keys)
		// ログ
		log.Infof("%s: done %+v", apiMethod, keys)

		// 要求終了を通知
		param.result = json.Parse(keys)
		param.done <- 0
		return false

	case "read":
		var items json.ElemArray
		if items, ok = jsonObj.Child("items").AsArray(); false == ok {
			// items がキーの配列でなかったのでエラー
			a.sendError(param, "read items must array")
			param.result = json.NewElemNull()
			param.done <- -1
			return false
		}

		// データを読み出す
		keys := make([]interface{}, 0)
		ret := json.NewElemObject()
		for idx := 0; idx < items.Size(); idx++ {
			item := items.Child(idx)
			key := item.Text()
			ret.Put(key, loadItemValue(storagePath, key))
			keys = append(keys, key)
		}
		// イベント通知
		a.sendArray(apiMethod, param, keys)
		// ログ
		log.Infof("%s: done %+v", apiMethod, ret)

		// 要求終了を通知
		param.result = ret
		param.done <- 0
		return false

	case "delete":
		var items json.ElemArray
		if items, ok = jsonObj.Child("items").AsArray(); false == ok {
			// items がキーの配列でなかったのでエラー
			a.sendError(param, "read items must array")
			param.result = json.NewElemNull()
			param.done <- -1
			return false
		}

		// データを削除
		keys := make([]interface{}, 0)
		for idx := 0; idx < items.Size(); idx++ {
			item := items.Child(idx)
			key := item.Text()
			if err := deleteItem(storagePath, key); err == nil {
				keys = append(keys, key)
			}
		}
		// イベント通知
		a.sendArray(apiMethod, param, keys)
		// ログ
		log.Infof("%s: done %+v", apiMethod, keys)

		// 要求終了を通知
		param.result = json.Parse(keys)
		param.done <- 0
		return false

	default:
		// 未知のAPIが指定されていたのでエラー
		a.sendError(param, "unknown api")
		param.result = json.NewElemNull()
		param.done <- -1
		return false

	}
}
