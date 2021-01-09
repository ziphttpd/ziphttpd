package httpd

import (
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// Param は handler の各機能へのパラメータです。
type param struct {
	conf     common.Config
	server   common.Server
	docHost  common.DocHost
	docGroup common.DocGroup
	docData  common.DocData
	paths    []string
}

func (p *param) Config() common.Config {
	return p.conf
}

func (p *param) PortMan() common.PortMan {
	return p.conf.PortMan()
}

func (p *param) DocHost() common.DocHost {
	if p.docHost == nil {
		hostName := p.conf.PortMan().HostName(p.server.Port())
		p.docHost = p.conf.DocHost(hostName)
	}
	return p.docHost
}

func (p *param) DocGroup() common.DocGroup {
	if p.docGroup == nil && len(p.paths) > 3 {
		docHost := p.DocHost()
		// paths : [0]:"" / [1]:{ホスト} / [2]:{グループ} / [3]:{ドキュメント}
		groupName := p.paths[2]
		p.docGroup = docHost.Get(groupName)
	}
	return p.docGroup
}

func (p *param) DocData() common.DocData {
	if p.docData == nil && len(p.paths) > 4 {
		docGroup := p.DocGroup()
		// paths : [0]:"" / [1]:{ホスト} / [2]:{グループ} / [3]:{ドキュメント}
		docID := p.paths[3]
		p.docData = docGroup.Get(docID)
	}
	return p.docData
}

func (p *param) Logger() common.Logger {
	return p.conf.Logger()
}

func (p *param) SecurityMan() common.SecurityMan {
	return p.conf.SecurityMan()
}

func (p *param) Version() string {
	return p.conf.Version()
}

func (p *param) ConfigPath() string {
	return p.conf.ConfigPath()
}

func (p *param) Server() common.Server {
	return p.server
}

func (p *param) ListenPort() int {
	return p.conf.ListenPort()
}

func (p *param) Favicon() []byte {
	return p.conf.Favicon()
}

func (p *param) Paths() []string {
	return p.paths
}

func (p *param) AdURL() string {
	return "/ad"
}
