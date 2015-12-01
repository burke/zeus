package shinylog

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
)

type ShinyLogger struct {
	mu             sync.Mutex
	happyLogger    log.Logger
	sadLogger      log.Logger
	suppressOutput bool
	disableColor   bool
}

func NewShinyLogger(out, err interface {
	io.Writer
}) *ShinyLogger {
	happyLogger := log.New(out, "", 0)
	sadLogger := log.New(err, "", log.Lshortfile)
	var mu sync.Mutex
	return &ShinyLogger{mu, *happyLogger, *sadLogger, false, false}
}

func NewTraceLogger(out interface {
	io.Writer
}) *log.Logger {
	return log.New(out, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

const (
	red         = "\x1b[31m"
	green       = "\x1b[32m"
	brightgreen = "\x1b[1;32m"
	yellow      = "\x1b[33m"
	blue        = "\x1b[34m"
	magenta     = "\x1b[35m"
	reset       = "\x1b[0m"
)

var dlm sync.RWMutex
var defaultLogger *ShinyLogger = NewShinyLogger(os.Stdout, os.Stderr)
var traceLogger *log.Logger = nil

func DefaultLogger() *ShinyLogger {
	dlm.RLock()
	defer dlm.RUnlock()
	return defaultLogger
}

func SetDefaultLogger(sl *ShinyLogger) {
	dlm.Lock()
	defaultLogger = sl
	dlm.Unlock()
}

func TraceLogger() *log.Logger {
	dlm.RLock()
	defer dlm.RUnlock()
	return traceLogger
}

func SetTraceLogger(sl *log.Logger) {
	dlm.Lock()
	traceLogger = sl
	dlm.Unlock()
}

func Suppress()                           { DefaultLogger().Suppress() }
func DisableColor()                       { DefaultLogger().DisableColor() }
func Colorized(msg string) (printed bool) { return DefaultLogger().Colorized(msg) }
func Error(err error) bool                { return DefaultLogger().Error(err) }
func FatalError(err error)                { DefaultLogger().FatalError(err) }
func FatalErrorString(msg string)         { DefaultLogger().FatalErrorString(msg) }
func ErrorString(msg string) bool         { return DefaultLogger().ErrorString(msg) }
func Red(msg string) bool                 { return DefaultLogger().Red(msg) }
func Green(msg string) bool               { return DefaultLogger().Green(msg) }
func Brightgreen(msg string) bool         { return DefaultLogger().Brightgreen(msg) }
func Yellow(msg string) bool              { return DefaultLogger().Yellow(msg) }
func Blue(msg string) bool                { return DefaultLogger().Blue(msg) }
func Magenta(msg string) bool             { return DefaultLogger().Magenta(msg) }

func TraceEnabled() bool {
	return TraceLogger() != nil
}

func Trace(format string, v ...interface{}) bool {
	if TraceEnabled() {
		TraceLogger().Printf(format, v...)
		return true
	}
	return false
}

func (l *ShinyLogger) Suppress() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.suppressOutput = true
}

func (l *ShinyLogger) DisableColor() {
	l.disableColor = true
}

func (l *ShinyLogger) Colorized(msg string) (printed bool) {
	return l.colorized(3, msg, false, true)
}

func (l *ShinyLogger) ColorizedSansNl(msg string) (printed bool) {
	return l.colorized(3, msg, false, false)
}

// If we send SIGTERM rather than explicitly exiting,
// the signal can be handled and the master can clean up.
// This is a workaround for Go not having `atexit` :(.
func terminate() {
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.SIGTERM)
}

func (l *ShinyLogger) FatalErrorString(msg string) {
	l.colorized(3, "{red}"+msg, true, true)
	terminate()
}

func (l *ShinyLogger) FatalError(err error) {
	l.colorized(3, "{red}"+err.Error(), true, true)
	terminate()
}

func (l *ShinyLogger) Error(err error) bool {
	return l.colorized(3, "{red}"+err.Error(), true, true)
}

func (l *ShinyLogger) ErrorString(msg string) bool {
	return l.colorized(3, "{red}"+msg, true, true)
}

func (l *ShinyLogger) Red(msg string) bool {
	return l.colorized(3, "{red}"+msg, false, true)
}

func (l *ShinyLogger) Green(msg string) bool {
	return l.colorized(3, "{green}"+msg, false, true)
}

func (l *ShinyLogger) Brightgreen(msg string) bool {
	return l.colorized(3, "{brightgreen}"+msg, false, true)
}

func (l *ShinyLogger) Yellow(msg string) bool {
	return l.colorized(3, "{yellow}"+msg, false, true)
}

func (l *ShinyLogger) Blue(msg string) bool {
	return l.colorized(3, "{blue}"+msg, false, true)
}

func (l *ShinyLogger) Magenta(msg string) bool {
	return l.colorized(3, "{magenta}"+msg, false, true)
}

func (l *ShinyLogger) formatColors(msg string) string {
	if l.disableColor {
		msg = strings.Replace(msg, "{red}", "", -1)
		msg = strings.Replace(msg, "{green}", "", -1)
		msg = strings.Replace(msg, "{brightgreen}", "", -1)
		msg = strings.Replace(msg, "{yellow}", "", -1)
		msg = strings.Replace(msg, "{blue}", "", -1)
		msg = strings.Replace(msg, "{magenta}", "", -1)
		msg = strings.Replace(msg, "{reset}", "", -1)
	} else {
		msg = strings.Replace(msg, "{red}", red, -1)
		msg = strings.Replace(msg, "{green}", green, -1)
		msg = strings.Replace(msg, "{brightgreen}", brightgreen, -1)
		msg = strings.Replace(msg, "{yellow}", yellow, -1)
		msg = strings.Replace(msg, "{blue}", blue, -1)
		msg = strings.Replace(msg, "{magenta}", magenta, -1)
		msg = strings.Replace(msg, "{reset}", reset, -1)
	}
	return msg
}

func (l *ShinyLogger) colorized(callDepth int, msg string, isError bool, printNewline bool) (printed bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.suppressOutput {
		msg = l.formatColors(msg)

		if l == DefaultLogger() {
			callDepth += 1 // this was called through a proxy method
		}
		if isError {
			l.sadLogger.Output(callDepth, msg+reset)
		} else {
			if printNewline {
				fmt.Println(msg + reset)
			} else {
				fmt.Print(msg + reset)
			}
			//l.happyLogger.Output(callDepth, msg+reset)
		}
	}
	return !l.suppressOutput
}
