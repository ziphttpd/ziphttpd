package logic

import (
	"fmt"
	"sync"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type api struct {
	config common.Config
	// 同時実行防止
	mu *sync.Mutex
	// ドキュメントグループ
	docGroupName string
	// ストレージのパス
	storagePath string
	// 要求のキュー先頭
	first *apiParam
	// 要求のキュー末尾
	last *apiParam
	// 要求受信時にキックされるチャネル
	kick chan int
	// 停止要求でキックされるチャネル
	done chan int
	// 状態通知
	// TODO: デタッチできるように再検討
	att []chan json.ElemObject
	// 停止
	terminated bool
}

// apiParam は単一のAPI処理です。
type apiParam struct {
	elem   json.Element
	next   *apiParam
	done   chan int
	result json.Element
}

var apiinstance map[string]*api

func init() {
	apiinstance = map[string]*api{}
}

// GetApi は保存領域別の Api のシングルトンです。
func GetApi(docGroupName, storagePath string, config common.Config) *api {
	if a, ok := apiinstance[storagePath]; ok {
		return a
	}

	a := &api{
		config:       config,
		mu:           &sync.Mutex{},
		docGroupName: docGroupName,
		storagePath:  storagePath,
		first:        nil,
		last:         nil,
		kick:         make(chan int, 100),
		done:         make(chan int),
		att:          []chan json.ElemObject{},
		terminated:   false,
	}
	apiinstance[storagePath] = a

	// バックグラウンド処理開始
	go backgroundLogic(a)
	return a
}

// Execute は API ロジックを同期実行する
func (a *api) Execute(jsonRequestStr string) (string, error) {
	log := a.config.Logger()
	log.Infof("[%s] json:%s", a.docGroupName, jsonRequestStr)
	if jsonRequestStr == "" {
		// ログインチェックの空打ち時
		return "", nil
	}
	requestElem, err := json.LoadFromJSONByte([]byte(jsonRequestStr))
	if err != nil {
		return "", err
	}

	// 非同期実行
	param := &apiParam{
		elem: requestElem,
		done: make(chan int),
	}
	a.push(param)

	// API 完了待ち
	ret := <-param.done
	if ret != 0 {
		return "", fmt.Errorf("err")
	}

	// 終了
	return param.result.Text(), nil
}

// Terminate はバックグラウンド処理を強制停止させます。
func (a *api) Terminate() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if false == a.terminated {
		a.terminated = true
		a.done <- 0
	}
}

// push は要求を待ちキューにプッシュします。
func (a *api) push(param *apiParam) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if nil == a.first {
		a.first = param
	} else {
		a.last.next = param
	}
	a.last = param
	a.kick <- 1
}

// pop は要求を待ちキューからポップします。
func (a *api) pop() *apiParam {
	a.mu.Lock()
	defer a.mu.Unlock()

	res := a.first
	if nil != res {
		a.first = res.next
	}
	return res
}

func (a *api) Attach() <-chan json.ElemObject {
	a.mu.Lock()
	defer a.mu.Unlock()

	ch := make(chan json.ElemObject)
	a.att = append(a.att, ch)
	return ch
}

func (a *api) fireEvent(data json.ElemObject) {
	for _, att := range a.att {
		att <- data
	}
}

func (a *api) sendMessage(kind string, param *apiParam, mes string) {
	event := json.NewElemObject()
	event.Put("kind", json.NewElemString(kind))
	event.Put("param", param.elem)
	event.Put("data", json.NewElemString(mes))
	a.fireEvent(event)
}

func (a *api) sendArray(kind string, param *apiParam, arr []interface{}) {
	event := json.NewElemObject()
	event.Put("kind", json.NewElemString(kind))
	event.Put("param", param.elem)
	keyArray := json.Parse(arr)
	event.Put("data", keyArray)
	a.fireEvent(event)
}

// エラーログ
func (a *api) sendError(param *apiParam, mes string) {
	a.sendMessage("error", param, mes)
}
