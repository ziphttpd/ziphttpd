package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/xorvercom/util/pkg/easywork"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
	iconfig "github.com/xorvercom/ziphttpd/cmd/internal/config"
	"github.com/xorvercom/ziphttpd/cmd/internal/httpd"
	"github.com/xorvercom/ziphttpd/cmd/internal/logic"
)

func main() {

	// TODO: ドキュメントのホットデプロイ(動的抜挿)を考えると代表ポートのサーバとドキュメントのサーバは分離すべき
	// 反論: 内部で全部落として再初期化のほうが単純で良いのでは？
	//       セッションの保持は難しいが良いのか？リブートとしては正しいが

	// TODO: 全般的に設定情報とかモデル間のデータ授受がダサい
	// 反論: とりあえず計画性よりも実際に動くものを構築したため
	//       いずれリニューアルする

	util := common.NewUtil()
	var (
		confPath     = flag.String("config", "", "configuration directory")
		logPath      = flag.String("log", "", "logging directory")
		listenPort   = flag.Int("port", common.DefaultListenPort, "listen port")
		firstDocPort = flag.Int("docport", common.DefaultFirstDocPort, "document listen port")
	)
	flag.Parse()
	util.SetConfigDir(*confPath)
	util.SetLogDir(*logPath)
	util.SetListenPort(*listenPort)
	util.SetFirstDocPort(*firstDocPort)

	// 設定読み込み
	conf, err := iconfig.OpenConfig(util)
	if err != nil {
		panic(err)
	}
	defer conf.Close()
	fmt.Println(conf)

	log := conf.Logger()
	log.Info("---- server start ----")

	wg := easywork.NewGroup()
	defer wg.Wait()

	// サーバ生成
	for _, hostName := range conf.PortMan().HostNames() {
		// API 設定
		docHost := conf.DocHost(hostName)
		docHost.SetAPI(logic.GetApi(hostName, docHost.GetAPIPath(), conf))
		// サーバ起動
		wg.Start(httpd.NewServer(conf, hostName))
	}

	// Interrupt検知
	interuptChan := make(chan os.Signal, 1)
	go func() {
		signal.Notify(interuptChan)
		go func() {
			// シグナルを同期して待つ
			func() {
				for {
					s := <-interuptChan
					conf.Logger().Infof("signal: %v", s)
					switch s {
					case os.Interrupt:
						return
					case os.Kill:
						return
					}
				}
			}()
			// リスナを閉じる
			conf.PortMan().Close()
		}()
	}()

	// 標準入力からのコマンド
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		text := stdin.Text()
		if strings.ToLower(text) == "quit" {
			interuptChan <- os.Interrupt
			break
		}
	}
	log.Info("---- server stop ----")
}
