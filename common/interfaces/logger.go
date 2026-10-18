package commoninterfaces

type Logger interface {
	Info(msg string, args ...any)
	Warning(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}
