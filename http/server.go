package http

import (
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
)

var dialer *net.Dialer

func dial(network, addr string) (net.Conn, error) {
	return dialer.Dial(network, addr)
}

func NewServer(config map[string]interface{}) (*http.Server, error) {
	proxy := goproxy.NewProxyHttpServer()

	if bind, ok := config["bind"].(string); ok {
		localAddr, err := net.ResolveTCPAddr("tcp", bind+":0")
		if err != nil {
			return nil, err
		}

		dialer = &net.Dialer{
			LocalAddr: localAddr,
		}

		proxy.Tr.Dial = dial
		proxy.ConnectDial = nil
	}

	return &http.Server{Handler: proxy}, nil
}
