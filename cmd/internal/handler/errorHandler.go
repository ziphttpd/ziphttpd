package handler

import (
	"github.com/xorvercom/util/pkg/calledcheck"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// ErrorHandler はエラーを処理するハンドラです。
func ErrorHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param, errorcode int) {
	param.Logger().Warnf("errorcode: %d (calledby:%s)", errorcode, calledcheck.GetCallerPC().String())
	writer.WriteHeader(errorcode)
}
