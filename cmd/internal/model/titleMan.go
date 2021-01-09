package model

import (
	"sort"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
	"github.com/ziphttpd/zhsig/pkg/zhsig"
)

// DocTitle はドキュメントのタイトル情報です。
type DocTitle struct {
	// title はドキュメントのタイトルです。
	title string
	// description はドキュメントの注釈です。
	description string
}

// JSON はJSONオブジェクトを返します。
func (d *DocTitle) JSON() json.ElemObject {
	elem := json.NewElemObject()
	elem.Put("title", json.NewElemString(d.title))
	elem.Put("description", json.NewElemString(d.description))
	return elem
}

// Title はドキュメントのタイトルです。
func (d *DocTitle) Title() string {
	return d.title
}

// Description はドキュメントの注釈です。
func (d *DocTitle) Description() string {
	return d.description
}

// GroupTitle はドキュメントグループのタイトル情報です。
type GroupTitle struct {
	// title はドキュメントグループのタイトルです
	title string
	// description はドキュメントグループの注釈です
	description string
	// Docs はドキュメントの辞書です
	Docs map[string]*DocTitle
}

// JSON はJSONオブジェクトを返します。
func (g *GroupTitle) JSON() json.ElemObject {
	elem := json.NewElemObject()
	elem.Put("title", json.NewElemString(g.title))
	elem.Put("description", json.NewElemString(g.description))
	docsDic := json.NewElemObject()
	for _, id := range g.Ids() {
		group := g.Docs[id]
		docsDic.Put(id, group.JSON())
	}
	elem.Put("Docs", docsDic)
	return elem
}

// Ids はホストしているドキュメントグループのタイトル名の一覧を取得します。
func (g *GroupTitle) Ids() []string {
	keys := make([]common.DocGroupName, 0, len(g.Docs))
	for key := range g.Docs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// AddDoc はドキュメントを追加します。
func (g *GroupTitle) AddDoc(name, title, description string) common.DocTitle {
	doc := &DocTitle{
		title:       title,
		description: description,
	}
	g.Docs[name] = doc
	return doc
}

// Title はドキュメントグループのタイトルです。
func (g *GroupTitle) Title() string {
	return g.title
}

// Description はドキュメントグループの注釈です。
func (g *GroupTitle) Description() string {
	return g.description
}

// Doc はドキュメントのタイトル情報を返します。
func (g *GroupTitle) Doc(name string) common.DocTitle {
	doc, ok := g.Docs[name]
	if false == ok {
		// 無かったので空で返す
		return &DocTitle{}
	}
	return doc
}

// HostTitle はホストのタイトル情報を管理します。
type HostTitle struct {
	peer *zhsig.PeerInfo
	// Docs はドキュメントグループの辞書です
	Groups map[string]*GroupTitle
}

// JSON はJSONオブジェクトを返します。
func (h *HostTitle) JSON() json.ElemObject {
	elem := json.NewElemObject()
	groupsDic := json.NewElemObject()
	for _, id := range h.Ids() {
		group := h.Groups[id]
		groupsDic.Put(id, group.JSON())
	}
	elem.Put("Groups", groupsDic)
	return elem
}

// Ids はホストしているドキュメントグループのタイトル名の一覧を取得します。
func (h *HostTitle) Ids() []string {
	keys := make([]common.DocGroupName, 0, len(h.Groups))
	for key := range h.Groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// AddGroup はドキュメントグループを追加します。
func (h *HostTitle) AddGroup(name, title, description string) common.GroupTitle {
	group := &GroupTitle{
		title:       title,
		description: description,
		Docs:        map[string]*DocTitle{},
	}
	h.Groups[name] = group
	return group
}

// Group はドキュメントグループのタイトル情報を返します。
func (h *HostTitle) Group(name string) common.GroupTitle {
	group, ok := h.Groups[name]
	if false == ok {
		// 無かったので空で返す
		return &GroupTitle{Docs: map[string]*DocTitle{}}
	}
	return group
}

// Peer はピア情報を返します。
func (h *HostTitle) Peer() *zhsig.PeerInfo {
	return h.peer
}

// TitleMan はタイトル情報を管理します。
type TitleMan struct {
	hosts map[string]*HostTitle
}

// JSON はJSONオブジェクトを返します。
func (t *TitleMan) JSON() json.ElemObject {
	elem := json.NewElemObject()
	hosts := json.NewElemObject()
	for _, id := range t.Ids() {
		group := t.hosts[id]
		hosts.Put(id, group.JSON())
	}
	elem.Put("hosts", hosts)
	return elem
}

// Ids はホストしているドキュメントグループのタイトル名の一覧を取得します。
func (t *TitleMan) Ids() []string {
	keys := make([]common.DocGroupName, 0, len(t.hosts))
	for key := range t.hosts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// NewTitleMan はタイトル情報を管理するマネージャを作成します。
func NewTitleMan() *TitleMan {
	return &TitleMan{hosts: map[string]*HostTitle{}}
}

// AddHost はホストのタイトル情報を追加します。
func (t *TitleMan) AddHost(name string, peer *zhsig.PeerInfo) *HostTitle {
	h := &HostTitle{
		peer:   peer,
		Groups: map[string]*GroupTitle{},
	}
	t.hosts[name] = h
	return h
}

// Host はホストのタイトル情報を返します。
func (t *TitleMan) Host(name string) common.HostTitle {
	host, ok := t.hosts[name]
	if false == ok {
		// 無かったので空で返す
		return &HostTitle{Groups: map[string]*GroupTitle{}}
	}
	return host
}
