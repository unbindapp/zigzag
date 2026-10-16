package log

import (
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	cblog "github.com/charmbracelet/log"
)

type Logger struct{ *cblog.Logger }

func (l *Logger) Printf(format string, v ...any) { l.Infof(format, v...) }

func (l *Logger) Fatalf(format string, v ...any) { l.Logger.Fatalf(format, v...) }

var (
	logger     *Logger
	initLogger sync.Once
)

func GetLogger() *Logger {
	initLogger.Do(func() {
		styles := cblog.DefaultStyles()
		styles.Levels[cblog.FatalLevel] = lipgloss.NewStyle().SetString("FATAL")
		styles.Levels[cblog.ErrorLevel] = lipgloss.NewStyle().SetString("ERROR")
		styles.Levels[cblog.WarnLevel] = lipgloss.NewStyle().SetString("WARN")
		styles.Levels[cblog.InfoLevel] = lipgloss.NewStyle().SetString("INFO")

		base := cblog.New(os.Stderr)
		base.SetStyles(styles)
		base.SetReportTimestamp(false)

		logger = &Logger{base}
	})
	return logger
}

func Info(msg any, keyvals ...any)   { GetLogger().Info(msg, keyvals...) }
func Infof(format string, v ...any)  { GetLogger().Infof(format, v...) }
func Warn(msg any, keyvals ...any)   { GetLogger().Warn(msg, keyvals...) }
func Warnf(format string, v ...any)  { GetLogger().Warnf(format, v...) }
func Error(msg any, keyvals ...any)  { GetLogger().Error(msg, keyvals...) }
func Errorf(format string, v ...any) { GetLogger().Errorf(format, v...) }
func Fatal(msg any, keyvals ...any)  { GetLogger().Fatal(msg, keyvals...) }
func Fatalf(format string, v ...any) { GetLogger().Fatalf(format, v...) }

func SetLevel(level string) {
	parsed, err := cblog.ParseLevel(level)
	if err != nil {
		GetLogger().Warn("Unknown log level, keeping info", "level", level)
		return
	}
	GetLogger().SetLevel(parsed)
}
