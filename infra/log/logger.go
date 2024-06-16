package log

type Logger interface {
	Info(format string, a ...any)
	InfoJson(obj interface{})
}
