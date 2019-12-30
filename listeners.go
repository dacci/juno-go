// +build !linux

package main

import "net"

func listeners(config map[string]interface{}) ([]net.Listener, error) {
	return nativeListeners(config)
}
