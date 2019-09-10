// +build !linux
// +build !windows

package service

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"path"
	"syscall"
)

type loggerImpl interface {
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Info(format string, v ...interface{})
}

type genericService struct {
	loggerImpl
	request chan os.Signal
}

func (this *genericService) Ready() {
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

type syslogLogger struct {
	writer *syslog.Writer
}

func (logger *syslogLogger) Error(format string, v ...interface{}) {
	logger.writer.Err(fmt.Sprintf(format, v...))
}

func (logger *syslogLogger) Warning(format string, v ...interface{}) {
	logger.writer.Warning(fmt.Sprintf(format, v...))
}

func (logger *syslogLogger) Info(format string, v ...interface{}) {
	logger.writer.Info(fmt.Sprintf(format, v...))
}

type consoleLogger struct{}

func (*consoleLogger) Error(format string, v ...interface{}) {
	log.Printf(" ERR: "+format, v...)
}

func (*consoleLogger) Warning(format string, v ...interface{}) {
	log.Printf("WARN: "+format, v...)
}

func (*consoleLogger) Info(format string, v ...interface{}) {
	log.Printf("INFO: "+format, v...)
}

func Main(handler func(Service)) {
	var logger loggerImpl
	if writer, err := syslog.New(syslog.LOG_DAEMON, path.Base(os.Args[0])); err == nil {
		defer writer.Close()
		logger = &syslogLogger{writer}
	} else {
		logger = &consoleLogger{}
	}

	service := &genericService{
		logger,
		make(chan os.Signal, 1),
	}

	signal.Notify(service.request, syscall.SIGINT, syscall.SIGTERM)

	handler(service)
}
