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

func nativeListeners(config map[string]interface{}) ([]net.Listener, error) {
	switch value := config["listenStream"].(type) {
	case []string:
		listeners := make([]net.Listener, len(value))

		for i, address := range value {
			if listener, err := net.Listen("tcp", address); err == nil {
				listeners[i] = listener
			}
		}

		return listeners, nil

	case string:
		if listener, err := net.Listen("tcp", value); err == nil {
			return []net.Listener{listener}, nil
		} else {
			return []net.Listener{}, nil
		}

	case nil:
		return []net.Listener{}, nil
	}

	return nil, fmt.Errorf("illegal type for listenStream: %T", config["listenStream"])
}
