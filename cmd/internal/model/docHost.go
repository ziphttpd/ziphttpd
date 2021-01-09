package model

import (
	"sort"
	"strconv"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type docHostInst struct {
	// ホスト名
	name common.HostName
	// ポート番号
	port string
	// WebAPI のフォルダ
	apiPath string
	// api
	api common.API
	// セキュリティマネージャ
	sec common.SecurityMan
	// ドキュメントグループ
	groupDic map[common.HostName]common.DocGroup
	// タイトル
	title string
	// 説明
	description string
}

// JSON はJSONオブジェクトを返します。
func (h *docHostInst) JSON() json.ElemObject {
	elem := json.NewElemObject()
	elem.Put("name", json.NewElemString(h.name))
	elem.Put("port", json.NewElemString(h.port))
	elem.Put("apiPath", json.NewElemString(h.apiPath))
	groupsDic := json.NewElemObject()
	for _, id := range h.Ids() {
		group := h.groupDic[id]
		groupsDic.Put(group.Name(), group.JSON())
	}
	elem.Put("groups", groupsDic)
	elem.Put("title", json.NewElemString(h.title))
	elem.Put("description", json.NewElemString(h.description))
	return elem
}

// NewDocHost はドキュメントホストを作成します。
func NewDocHost(conf common.Config, host common.HostName) common.DocHost {
	docport := strconv.Itoa(conf.PortMan().Port(host))
	return &docHostInst{
		port:     docport,
		name:     host,
		apiPath:  conf.APIPath(host),
		sec:      conf.SecurityMan(),
		groupDic: map[common.DocGroupName]common.DocGroup{},
	}
}

// Name はホスト名称を取得します。
func (h *docHostInst) Name() common.HostName {
	return h.name
}

// Port はポート番号を取得します。
func (h *docHostInst) Port() string {
	return h.port
}

// Put はドキュメントグループを追加します。
func (h *docHostInst) Put(groupid common.DocGroupName, group common.DocGroup) {
	h.groupDic[groupid] = group
}

// Get はドキュメントグループを取得します。
func (h *docHostInst) Get(group common.DocGroupName) common.DocGroup {
	if g, ok := h.groupDic[group]; ok {
		return g
	}
	return nil
}

// Ids はホストしているドキュメントグループの名前の一覧を取得します。
func (h *docHostInst) Ids() []common.DocGroupName {
	keys := make([]common.DocGroupName, 0, len(h.groupDic))
	for key := range h.groupDic {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Token はCSRFトークンを返します
func (h *docHostInst) Token() common.Token {
	return h.sec.Token(h.name)
}

// GetAPIPath は WebAPI ロジックで使用するフォルダを返します。
func (h *docHostInst) GetAPIPath() string {
	return h.apiPath
}

// GetAPI はgetter
func (h *docHostInst) GetAPI() common.API {
	return h.api
}

// SetAPI はsetter
func (h *docHostInst) SetAPI(api common.API) {
	h.api = api
}

// Close はホストしているドキュメントグループをクローズします。
func (h *docHostInst) Close() {
	for _, v := range h.groupDic {
		v.Close()
	}
	h.groupDic = map[common.HostName]common.DocGroup{}
}

// SetTitleInfo はタイトル情報を設定します。
func (h *docHostInst) SetTitleInfo(title, description string) {
	h.title = title
	h.description = description
}

// Title はタイトルを返します。
func (h *docHostInst) Title() string {
	return h.title
}

// Description は説明を返します。
func (h *docHostInst) Description() string {
	return h.description
}
