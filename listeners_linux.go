package main

import (
	"net"

	"github.com/coreos/go-systemd/activation"
)

func listeners(config map[string]interface{}) ([]net.Listener, error) {
	if systemd, ok := config["systemd"].(bool); ok && !systemd {
		return nativeListeners(config)
	} else {
		return activation.Listeners()
	}
}
