package service

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

type windowsService struct {
	handler func(Service)
	args    []string
	request <-chan svc.ChangeRequest
	status  chan<- svc.Status
	log     *eventlog.Log
}

func (this *windowsService) Args() []string {
	return this.args
}

func (this *windowsService) Ready() {
	this.status <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}
	defer func() {
		this.status <- svc.Status{State: svc.StopPending}
	}()

	for {
		select {
		case c := <-this.request:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				return
			}
		}
	}
}

func (this *windowsService) Error(format string, v ...interface{}) {
	this.log.Error(0, fmt.Sprintf(format, v...))
}

func (this *windowsService) Warning(format string, v ...interface{}) {
	this.log.Warning(0, fmt.Sprintf(format, v...))
}

func (this *windowsService) Info(format string, v ...interface{}) {
	this.log.Info(0, fmt.Sprintf(format, v...))
}

func (this *windowsService) Execute(args []string, request <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	this.args = args
	this.request = request
	this.status = status

	status <- svc.Status{State: svc.StartPending}
	this.handler(this)
	status <- svc.Status{State: svc.Stopped}

	return
}

func Main(handler func(Service)) {
	log, err := eventlog.Open("Application")
	if err != nil {
		panic(err)
	}
	defer log.Close()

	service := &windowsService{
		handler: handler,
		log:     log,
	}

	svc.Run("", service)
}
