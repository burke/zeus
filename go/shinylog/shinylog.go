package shinylog

import (
	"log"
	"os"
	"strings"
	"sync"
	"io"
)

type ShinyLogger struct {
	mu             sync.Mutex
	happyLogger    log.Logger
	sadLogger      log.Logger
	suppressOutput bool
	disableColor   bool
}

func NewShinyLogger(out, err interface{io.Writer}) *ShinyLogger {
	happyLogger := log.New(out, "", 0)
	sadLogger := log.New(err, "", log.Lshortfile)
	var mu sync.Mutex
	return &ShinyLogger{mu, *happyLogger, *sadLogger, false, false}
}

const (
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	reset   = "\x1b[0m"
)

var DefaultLogger *ShinyLogger = NewShinyLogger(os.Stdout, os.Stderr)

func Suppress()                           { DefaultLogger.Suppress() }
func DisableColor()                       { DefaultLogger.DisableColor() }
func Colorized(msg string) (printed bool) { return DefaultLogger.Colorized(msg) }
func Error(err error) bool                { return DefaultLogger.Error(err) }
func FatalError(err error)                { DefaultLogger.FatalError(err) }
func FatalErrorString(msg string)         { DefaultLogger.FatalErrorString(msg) }
func ErrorString(msg string) bool         { return DefaultLogger.ErrorString(msg) }
func Red(msg string) bool                 { return DefaultLogger.Red(msg) }
func Green(msg string) bool               { return DefaultLogger.Green(msg) }
func Yellow(msg string) bool              { return DefaultLogger.Yellow(msg) }
func Blue(msg string) bool                { return DefaultLogger.Blue(msg) }
func Magenta(msg string) bool             { return DefaultLogger.Magenta(msg) }

func (l *ShinyLogger) Suppress() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.suppressOutput = true
}

func (l *ShinyLogger) DisableColor() {
	l.disableColor = true
}

func (l *ShinyLogger) Colorized(msg string) (printed bool) {
	return l.colorized(3, msg, false)
}

func (l *ShinyLogger) FatalErrorString(msg string) {
	l.colorized(3, "{red}"+msg, true)
	os.Exit(1)
}

func (l *ShinyLogger) FatalError(err error) {
	l.colorized(3, "{red}"+err.Error(), true)
	os.Exit(1)
}

func (l *ShinyLogger) Error(err error) bool {
	return l.colorized(3, "{red}"+err.Error(), true)
}

func (l *ShinyLogger) ErrorString(msg string) bool {
	return l.colorized(3, "{red}"+msg, true)
}

func (l *ShinyLogger) Red(msg string) bool {
	return l.colorized(3, "{red}"+msg, false)
}

func (l *ShinyLogger) Green(msg string) bool {
	return l.colorized(3, "{green}"+msg, false)
}

func (l *ShinyLogger) Yellow(msg string) bool {
	return l.colorized(3, "{yellow}"+msg, false)
}

func (l *ShinyLogger) Blue(msg string) bool {
	return l.colorized(3, "{blue}"+msg, false)
}

func (l *ShinyLogger) Magenta(msg string) bool {
	return l.colorized(3, "{magenta}"+msg, false)
}

func (l *ShinyLogger) formatColors(msg string) string {
	if l.disableColor {
		msg = strings.Replace(msg, "{red}", "", -1)
		msg = strings.Replace(msg, "{green}", "", -1)
		msg = strings.Replace(msg, "{yellow}", "", -1)
		msg = strings.Replace(msg, "{blue}", "", -1)
		msg = strings.Replace(msg, "{magenta}", "", -1)
		msg = strings.Replace(msg, "{reset}", "", -1)
	} else {
		msg = strings.Replace(msg, "{red}", red, -1)
		msg = strings.Replace(msg, "{green}", green, -1)
		msg = strings.Replace(msg, "{yellow}", yellow, -1)
		msg = strings.Replace(msg, "{blue}", blue, -1)
		msg = strings.Replace(msg, "{magenta}", magenta, -1)
		msg = strings.Replace(msg, "{reset}", reset, -1)
	}
	return msg
}

func (l *ShinyLogger) colorized(callDepth int, msg string, isError bool) (printed bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.suppressOutput {
		msg = l.formatColors(msg)

		if l == DefaultLogger {
			callDepth += 1 // this was called through a proxy method
		}
		if isError {
			l.sadLogger.Output(callDepth, msg+reset)
		} else {
			l.happyLogger.Output(callDepth, msg+reset)
		}
	}
	return !l.suppressOutput
}
