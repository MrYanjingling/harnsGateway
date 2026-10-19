// Package logging provides a shared structured logging layer for all harnsgateway
// components. It is backed by log/slog and exposes a klog-like facade so that
// call sites can keep using V(level).InfoS(...), InfoS(...), ErrorS(...) etc.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

const DefaultVerbosity = 2

var (
	logger    atomic.Pointer[slog.Logger]
	verbosity atomic.Int32
)

func init() {
	logger.Store(NewLogger(FormatText, os.Stderr))
	verbosity.Store(DefaultVerbosity)
}

// Logger returns the current structured logger.
func Logger() *slog.Logger {
	return logger.Load()
}

// SetLogger replaces the underlying structured logger.
func SetLogger(l *slog.Logger) {
	logger.Store(l)
}

// SetVerbosity sets the maximum V level that will be emitted, mirroring klog -v.
func SetVerbosity(v int) {
	verbosity.Store(int32(v))
}

// Verbosity returns the current V level threshold.
func Verbosity() int {
	return int(verbosity.Load())
}

// Verbose mirrors klog's V(n) helper.
type Verbose struct {
	level int
}

// V returns a Verbose handle for the given verbosity level.
func V(level int) Verbose {
	return Verbose{level: level}
}

// Enabled reports whether this verbosity level is currently enabled.
func (v Verbose) Enabled() bool {
	return v.level <= Verbosity()
}

// Info logs a message when the verbosity level is enabled.
func (v Verbose) Info(args ...interface{}) {
	if !v.Enabled() {
		return
	}
	Logger().Info(fmt.Sprint(args...))
}

// Infof logs a formatted message when the verbosity level is enabled.
func (v Verbose) Infof(format string, args ...interface{}) {
	if !v.Enabled() {
		return
	}
	Logger().Info(fmt.Sprintf(format, args...))
}

// InfoS logs a structured message when the verbosity level is enabled.
func (v Verbose) InfoS(msg string, keysAndValues ...interface{}) {
	if !v.Enabled() {
		return
	}
	Logger().Info(msg, keysAndValues...)
}

// Info logs a message at info level.
func Info(args ...interface{}) {
	Logger().Info(fmt.Sprint(args...))
}

// Infof logs a formatted message at info level.
func Infof(format string, args ...interface{}) {
	Logger().Info(fmt.Sprintf(format, args...))
}

// InfoS logs a structured message at info level.
func InfoS(msg string, keysAndValues ...interface{}) {
	Logger().Info(msg, keysAndValues...)
}

// InfoDepth logs a message at info level. The call depth is accepted for API
// compatibility and ignored.
func InfoDepth(_ int, args ...interface{}) {
	Logger().Info(fmt.Sprint(args...))
}

// Warn logs a message at warn level.
func Warn(args ...interface{}) {
	Logger().Warn(fmt.Sprint(args...))
}

// Warnf logs a formatted message at warn level.
func Warnf(format string, args ...interface{}) {
	Logger().Warn(fmt.Sprintf(format, args...))
}

// Error logs a message at error level.
func Error(args ...interface{}) {
	Logger().Error(fmt.Sprint(args...))
}

// Errorf logs a formatted message at error level.
func Errorf(format string, args ...interface{}) {
	Logger().Error(fmt.Sprintf(format, args...))
}

// ErrorS logs a structured message at error level. A non-nil err is attached
// under the "err" key.
func ErrorS(err error, msg string, keysAndValues ...interface{}) {
	args := keysAndValues
	if err != nil {
		args = append([]interface{}{"err", err}, keysAndValues...)
	}
	Logger().Error(msg, args...)
}

// Fatal logs a message at error level and exits with status 1.
func Fatal(args ...interface{}) {
	Logger().Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf logs a formatted message at error level and exits with status 1.
func Fatalf(format string, args ...interface{}) {
	Logger().Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Flush is a no-op kept for API compatibility with klog.Flush.
func Flush() {}

// Trace logs a method entry at V(3) and returns a function that logs the exit
// with elapsed time when deferred:
//
//	defer logging.Trace("identity.CreateUser", "userName", req.UserName)()
func Trace(method string, kv ...interface{}) func() {
	V(3).InfoS(method+" enter", kv...)
	start := time.Now()
	return func() {
		V(3).InfoS(method+" exit", "cost", time.Since(start))
	}
}

// Log format names accepted by NewLogger.
const (
	FormatText = "text"
	FormatJSON = "json"
)

// NewLogger builds a slog logger writing to w using the requested format.
func NewLogger(format string, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	if format == FormatJSON {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}
