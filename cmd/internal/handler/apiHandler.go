package handler

import (
	"net/http"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

const (
	// CSRFTOKEN はトークンのキーです
	CSRFTOKEN = "X-Requested-With"
)

// APIHandler はwebapiリクエストを処理するハンドラです。
// api は /api/ で始まるurlです。
// TODO まずは作り込み優先で検討はあと
func APIHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	log := param.Logger()

	// まずは POST であることが必須
	if request.Method() != http.MethodPost {
		ErrorHandler(writer, request, param, http.StatusForbidden)
		return
	}

	// ポートからドキュメントグループを特定
	//docGroup := param.Port2DocGroup(request.Port())

	// OSRF (Own Site Request Forgeries) トークン
	token := request.GetHeader(CSRFTOKEN)
	docHost := param.DocHost()
	log.Infof("[%s] token from client:%s", docHost.Name(), token)
	if docHost.Token() != token {
		// 認証エラー
		ErrorHandler(writer, request, param, http.StatusUnauthorized)
		return
	}

	// APIのパラメータ
	jsonRequestStr := request.GetPostForm("data")
	// グループごとの Api をキック
	//api := logic.GetApi(docGroupName, docGroup.GetApiPath(), param.Conf)
	api := docHost.GetAPI()
	res, err := api.Execute(jsonRequestStr)
	if err != nil {
		// 不正なリクエスト
		ErrorHandler(writer, request, param, http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.WriteContentsByte([]byte(res))
}
