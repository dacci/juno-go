// +build !linux

package main

import (
	"fmt"
	"net"

	"github.com/dacci/juno-go/service"
)

func listeners(config map[string]interface{}) ([]net.Listener, error) {
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

func main() {
	service.Main(func(service service.Service) {
		if *configFile == "" {
			service.Error("No configuration")
			return
		}

		config, err := loadConfig(*configFile)
		if err != nil {
			service.Error("Failed to load config: %s", err.Error())
			return
		}

		server, err := newServer(config)
		if err != nil {
			service.Error("Failed to create server: %s", err.Error())
			return
		}
		defer server.Close()

		if listeners, err := listeners(config); err == nil {
			for _, listener := range listeners {
				if listener != nil {
					go server.Serve(listener)
				}
			}
		} else {
			service.Warning("Failed to get listeners: %s", err.Error())
			return
		}

		service.Ready()
	})
}
