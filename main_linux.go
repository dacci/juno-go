package main

import (
	"github.com/coreos/go-systemd/activation"
	"github.com/dacci/juno-go/service"
)

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

		listeners, err := activation.Listeners()
		if err != nil {
			service.Error("Failed to get listeners: %s", err.Error())
			return
		}

		for _, listener := range listeners {
			if listener != nil {
				go server.Serve(listener)
			}
		}

		service.Ready()
	})
}
