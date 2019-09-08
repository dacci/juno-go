package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/dacci/juno-go/http"
	"github.com/dacci/juno-go/socks"
)

var (
	configFile = flag.String("c", "", "")
)

func init() {
	flag.Parse()
}

type StreamServer interface {
	Close() error
	Serve(net.Listener) error
}

func loadConfig(path string) (config map[string]interface{}, err error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return
	}

	return
}

func newServer(config map[string]interface{}) (service StreamServer, err error) {
	switch config["provider"] {
	case "socks":
		service, err = socks.NewServer(config)

	case "http":
		service, err = http.NewServer(config)

	default:
		err = fmt.Errorf("invalid provider")
	}

	return
}
