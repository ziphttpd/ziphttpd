package handler

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

// アイコンにセンシティブな情報がないか一応確認
var faviconStrArr = []string{
	"0000", // reserved
	"0100", // type (icon: 1)
	"0100", // icon count (==1)
	// directory 1
	"20",       // width (32)
	"20",       // height (32)
	"10",       // color count (16)
	"00",       // reserve
	"0000",     // color plane count
	"0000",     // bits per pixel
	"e8020000", // byte count (==0x02e8==744)
	"16000000", // byte offset (==0x16==22)
	// dib header (BITMAPINFOHEADER)
	"28000000", // bit size
	"20000000", // bit width
	"40000000", // bit height
	"0100",     // plane
	"0400",     // bit count
	"00000000", // compression
	"80020000", // size image
	"00000000", // XPelsPerMeter
	"00000000", // YPelsPerMeter
	"00000000", // ClrUsed
	"00000000", // ClrImportant
	// RGBQUAD
	"000000", "00", // 0
	"000080", "00", // 1
	"008000", "00", // 2
	"008080", "00", // 3
	"800000", "00", // 4
	"800080", "00", // 5
	"808000", "00", // 6
	"808080", "00", // 7
	"c0c0c0", "00", // 8
	"0000ff", "00", // 9
	"00ff00", "00", // a
	"00ffff", "00", // b
	"ff0000", "00", // c
	"ff00ff", "00", // d
	"ffff00", "00", // e
	"ffffff", "00", // f
	// bit pattern
	"ffffffffffffffffffffffffffffffff",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000000fffffffffffffffffffff000f",
	"f0000000ffffffffffffffffffff000f",
	"f00000000fffffffffffffffffff000f",
	"f000000000ffffffffffffffffff000f",
	"f0000000000fffffffffffffffff000f",
	"f00000000000ffffffffffffffff000f",
	"f000000000000fffffffffffffff000f",
	"f0000000000000ffffffffffffff000f",
	"f00000000000000fffffffffffff000f",
	"f000000000000000ffffffffffff000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000ffffffffffff000000000000000f",
	"f000fffffffffffff00000000000000f",
	"f000ffffffffffffff0000000000000f",
	"f000fffffffffffffff000000000000f",
	"f000ffffffffffffffff00000000000f",
	"f000fffffffffffffffff0000000000f",
	"f000ffffffffffffffffff000000000f",
	"f000fffffffffffffffffff00000000f",
	"f000ffffffffffffffffffff0000000f",
	"f000fffffffffffffffffffff000000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"f000000000000000000000000000000f",
	"ffffffffffffffffffffffffffffffff",
	//
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
	"00000000000000000000000000000000",
}
var faviconArr []byte

func init() {
	faviconStr := strings.Join(faviconStrArr, "")
	faviconArr, _ = hex.DecodeString(faviconStr)
}

// FaviconHandler はfavicon.icoファイルに対するリクエストを処理するハンドラです。
func FaviconHandler(writer common.ResponseProxy, request common.RequestProxy, param common.Param) {
	buf := param.Favicon()
	if buf == nil {
		// 定義ファイルにアイコンの指定が無かったので標準のアイコンを返す
		buf = faviconArr
	}
	writer.SetHeader("Content-Type", "image/ico")
	cl := strconv.Itoa(len(buf))
	writer.SetHeader("Content-Length", cl)
	// ヘッダの決定
	writer.WriteHeader(http.StatusOK)
	// 中身
	_, err := writer.WriteContentsByte(buf)
	if err != nil {
		param.Logger().Warnf("writer.ParseContents error : %+v", err)
		// エラーの時に、http.Server の ConnState ハンドルが呼ばれず現接続数の計算でミスする
		param.Server().ConnDone()
	}
}
