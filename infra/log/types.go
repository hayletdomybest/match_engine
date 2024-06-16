package log

type LogLevel int

const (
	Debug LogLevel = iota
	// InfoLevel defines info log level.
	Info
	// WarnLevel defines warn log level.
	Warn
	// ErrorLevel defines error log level.
	Error
	// FatalLevel defines fatal log level.
	Fatal
	// PanicLevel defines panic log level.
	Panic
	// NoLevel defines an absent log level.
	No
	// Disabled disables the logger.
	Disabled
)
