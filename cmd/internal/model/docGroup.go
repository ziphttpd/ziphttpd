package model

import (
	"sort"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type docGroupInst struct {
	// ポートグループ名称
	name common.DocGroupName
	// ホスト名
	host common.HostName
	// ドキュメント
	docsDic map[common.DocID]common.DocData
	// タイトル
	title string
	// 説明
	description string
}

// JSON はJSONオブジェクトを返します。
func (d *docGroupInst) JSON() json.ElemObject {
	elem := json.NewElemObject()
	elem.Put("name", json.NewElemString(d.name))
	elem.Put("host", json.NewElemString(d.host))
	docsDic := json.NewElemObject()
	for _, id := range d.Ids() {
		doc := d.docsDic[id]
		docsDic.Put(doc.DocID(), doc.JSON())
	}
	elem.Put("docs", docsDic)
	elem.Put("title", json.NewElemString(d.title))
	elem.Put("description", json.NewElemString(d.description))
	return elem
}

// NewDocGroup はドキュメントグループを作成します。
func NewDocGroup(host common.HostName, group common.DocGroupName) common.DocGroup {
	return &docGroupInst{
		name:    group,
		host:    host,
		docsDic: map[common.DocID]common.DocData{},
	}
}

// DocGroup はドキュメントグループ名称を取得します。
func (d *docGroupInst) Name() common.DocGroupName {
	return d.name
}

// Put はドキュメントを追加します。
func (d *docGroupInst) Put(docid common.DocID, doc common.DocData) {
	d.docsDic[docid] = doc
}

// DocData は zip ドキュメントを取得します。
func (d *docGroupInst) Get(docid common.DocID) common.DocData {
	return d.docsDic[docid]
}

// Ids はホストしている zip ドキュメントの名前の一覧を取得します。
func (d *docGroupInst) Ids() []common.DocID {
	keys := make([]common.DocID, 0, len(d.docsDic))
	for key := range d.docsDic {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Close はホストしている zip ドキュメントをクローズします。
func (d *docGroupInst) Close() {
	// すべてのドキュメントをクローズ
	for _, v := range d.docsDic {
		v.Close()
	}
	d.docsDic = map[common.DocID]common.DocData{}
}

// SetTitleInfo はタイトル情報をセットします。
func (d *docGroupInst) SetTitleInfo(title, description string) {
	d.title = title
	d.description = description
}

// Title はタイトルを返します。
func (d *docGroupInst) Title() string {
	return d.title
}

// Description は説明を返します。
func (d *docGroupInst) Description() string {
	return d.description
}
