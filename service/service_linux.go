package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	"github.com/coreos/go-systemd/journal"
)

type linuxService struct {
	request chan os.Signal
}

func (*linuxService) Args() []string {
	return os.Args[1:]
}

func (this *linuxService) Ready() {
	daemon.SdNotify(false, daemon.SdNotifyReady)
	defer daemon.SdNotify(false, daemon.SdNotifyStopping)

	for {
		select {
		case s := <-this.request:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				return
			}
		}
	}
}

func (*linuxService) Error(format string, v ...interface{}) {
	journal.Print(journal.PriErr, format, v...)
}

func (*linuxService) Warning(format string, v ...interface{}) {
	journal.Print(journal.PriWarning, format, v...)
}

func (*linuxService) Info(format string, v ...interface{}) {
	journal.Print(journal.PriInfo, format, v...)
}

func Main(handler func(Service)) {
	service := &linuxService{
		make(chan os.Signal, 1),
	}

	signal.Notify(service.request, syscall.SIGINT, syscall.SIGTERM)

	handler(service)
}
