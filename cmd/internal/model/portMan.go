package model

import (
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

const (
	systemDocGroup = "system"
)

type portManInst struct {
	nextPort  int
	listeners map[int]*net.TCPListener
	// ポートグループ名 -> ポート
	portMap     map[common.HostName]int
	hostNames   []common.HostName
	lockinPorts map[common.HostName]int
	lockinHosts map[int]common.HostName
}

// NewPortMan はコンストラクタです。
func NewPortMan(start int) common.PortMan {
	return &portManInst{
		nextPort:    start,
		listeners:   map[int]*net.TCPListener{},
		portMap:     map[common.HostName]int{},
		hostNames:   []common.HostName{},
		lockinPorts: map[common.HostName]int{},
		lockinHosts: map[int]common.HostName{},
	}
}

// OpenLockIn は
func (p *portManInst) OpenLockIn() error {
	for host, port := range p.portMap {
		if _, ok := p.listeners[port]; ok {
			err := p.Put(host, port)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PutLockIn は以前に利用していたポート番号を予約します。
func (p *portManInst) PutLockIn(host common.HostName, port int) {
	p.lockinPorts[host] = port
	p.lockinHosts[port] = host
}

// Port はグループ名のポートを返します。未登録ならば空いているポートを探して確保します。
func (p *portManInst) Port(host common.HostName) int {
	port, ok := p.portMap[host]
	if false == ok { // nolint:gosimple
		// 未登録ならば空いているポートを探して登録します。
		p.assign(host)
		port = p.portMap[host]
	}
	return port
}

// HostName はポート番号のドキュメントグループ名称を返します。
func (p *portManInst) HostName(port int) common.HostName {
	if docGroupName, ok := p.lockinHosts[port]; ok {
		return docGroupName
	}
	return ""
}

// LockInPorts はグループ名-ポートのマップを返します。
func (p *portManInst) LockInPorts() map[common.HostName]int {
	return p.lockinPorts
}

// HostNames はグループ名を返します。
func (p *portManInst) HostNames() []common.HostName {
	return p.hostNames
}

// assign は空いているポートを探して登録します。
func (p *portManInst) assign(host common.HostName) {
	// そのポートグループが以前に使われていたら、そのときのポートに割り当てる
	if port, ok := p.lockinPorts[host]; ok {
		err := p.Put(host, port)
		if err == nil {
			return
		}
	}
	// 空いているポートを探す
	for {
		port := p.nextPort
		p.nextPort++
		if _, ok := p.lockinHosts[port]; ok {
			// 予約されているのでスキップ
			continue
		}
		err := p.Put(host, port)
		if err == nil {
			return
		}
	}
}

// Listeners はリスナーを返します。
func (p *portManInst) Listener(host common.HostName) *net.TCPListener {
	return p.listeners[p.Port(host)]
}

// Put はポートを登録します。固定ポートの登録時に使用します。
func (p *portManInst) Put(host common.HostName, port int) error {
	laddr, _ := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	p.listeners[port] = listener
	p.portMap[host] = port
	p.hostNames = append(p.hostNames, host)
	sort.Strings(p.hostNames)
	// 使用ポート記録
	p.PutLockIn(host, port)
	return nil
}

// Close は全てのリスナをクローズします。
func (p *portManInst) Close() {
	for _, listener := range p.listeners {
		listener.Close()
	}
}

// Load はグループで使用するポートをポートロックインファイルから読みだします。
func (p *portManInst) Load(portsfile string) {
	element, err := json.LoadFromJSONFile(portsfile)
	if err != nil {
		// ポートロックインファイルが読めなかった
		return
	}
	if elemObj, ok := element.AsObject(); ok {
		for _, host := range elemObj.Keys() {
			if host == systemDocGroup {
				// system ドキュメントのポートは環境変数で得るため
				continue
			}
			if elemFlo, ok := elemObj.Child(host).AsFloat(); ok {
				port := int(elemFlo.Float())
				// ポートロックインを設定
				p.PutLockIn(host, port)
			}
		}
	}
}

// Save はポートグループで使用しているポートをポートロックインファイルに書き出します。
func (p *portManInst) Save(portsfile string) {
	// ポートの使用情報を収集
	obj := json.NewElemObject()
	for _, host := range p.HostNames() {
		if host == systemDocGroup {
			// system ドキュメントのポートは環境変数で得るため
			continue
		}
		port := p.Port(host)
		obj.Put(host, json.NewElemFloat(float64(port)))
	}
	// ポートロックインファイルに書き込み
	err := json.SaveToJSONFile(portsfile, obj, true)
	if err != nil {
		panic(fmt.Errorf("error json.SaveToJSONFile(%s, %+v, true) : %v", portsfile, obj, err))
	}
}
