package model

import (
	srand "crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type securityManInst struct {
	mu sync.Mutex
	// パスワード
	pass map[string]string
	// ローカルストレージを使用
	localstorage map[string]bool
	// ドキュメントグループに割り当てたトークンを探す
	tokenDic map[string]common.Token
}

// NewSecurityMan はコンストラクタです。
func NewSecurityMan(passwordfile string) common.SecurityMan {
	s := &securityManInst{tokenDic: map[string]common.Token{}}
	s.LoadPassword(passwordfile)
	return s
}

// LoadPassword はパスワードを読み込みます。
func (s *securityManInst) LoadPassword(passwordfile string) {
	var pass = make(map[string]string)
	var localstorage = make(map[string]bool)
	// password ファイルを読む
	if e, err := json.LoadFromJSONFile(passwordfile); err == nil {
		if eo, ok := e.AsObject(); ok {
			for _, key := range eo.Keys() {
				if edo, ok := eo.Child(key).AsObject(); ok {
					if es, ok := edo.Child("password").AsString(); ok {
						pass[key] = es.Text()
					}
					if es, ok := edo.Child("localstorage").AsBool(); ok {
						localstorage[key] = es.Bool()
					}
				}
			}
		}
	}
	s.pass = pass
	s.localstorage = localstorage
}

// Token はホストにトークンを振り出します。
func (s *securityManInst) Token(hostName common.HostName) common.Token {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.tokenDic[hostName]
	if false == ok { // nolint:gosimple
		r := make([]byte, 256/8)
		// secure random number generator を使用して256ビットのトークンを振り出す
		// これで推測されるならば、鍵を生成しておいてAES256で暗号化
		// それでも足りなければサイズは増えるけど公開鍵暗号
		_, err := srand.Read(r)
		if err != nil {
			panic(fmt.Errorf("%+v", err))
		}
		token = base64.StdEncoding.EncodeToString(r)
		s.tokenDic[hostName] = token
	}
	return token
}

// IsValid はドキュメントグループのパスワードをチェックします。
func (s *securityManInst) IsValid(hostName common.HostName, password string) bool {
	if pass, ok := s.pass[hostName]; ok {
		return pass == password
	}
	return false
}

func (s *securityManInst) UseLocalStorage(hostName common.HostName) bool {
	if localstorage, ok := s.localstorage[hostName]; ok {
		return localstorage
	}
	return false
}
