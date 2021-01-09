package handler

import (
	"net/url"
)

// SplitHost は Host をアドレスとポートに分割します。
func SplitHost(host string) (hostname, port string) {
	u, _ := url.Parse("http://" + host + "/")
	hostname = u.Hostname()
	port = u.Port()
	if port == "" {
		port = "80"
	}
	return hostname, port
}
