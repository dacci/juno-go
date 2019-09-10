package service

type Service interface {
	Ready()
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Info(format string, v ...interface{})
}
