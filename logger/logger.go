package logger

// ===================================================
// Imports Area
// ===================================================

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

// ===================================================
// Definitions Area
// ===================================================

// LogWriter wraps an io.Writer to maintain type consistency with atomic.Value
type LogWriter struct {
	w io.Writer
}
type LogLevel int32

type ILogger interface {
	Trace(msg string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// Logger handles all logging operations
type Logger struct {
	level atomic.Int32
	// holds LogWriter
	output atomic.Value
	// maps message to sync.Once
	onceMessages sync.Map
	colors       struct {
		magenta *color.Color
		blue    *color.Color
		red     *color.Color
		cyan    *color.Color
		yellow  *color.Color
	}
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func NewLoggerWithLevelAndOutput(level LogLevel, w io.Writer) *Logger {
	l := newDefaultLogger()
	l.level.Store(int32(level))
	l.output.Store(LogWriter{w})
	return l
}

type StructuredLogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data,omitempty"`
	Caller    string         `json:"caller,omitempty"`
}

// ===================================================
// Declarations Area
// ===================================================

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
	LevelAudit
)

// Global logger instance
var defaultLogger = newDefaultLogger()
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

var structuredLogPool = sync.Pool{
	New: func() interface{} {
		return &StructuredLogEntry{
			Data: make(map[string]any),
		}
	},
}

// ===================================================
// Public Functions Area
// ===================================================

func SetLogLevel(level LogLevel) {
	defaultLogger.level.Store(int32(level))
}

func SetLogOutput(w io.Writer) {
	defaultLogger.output.Store(LogWriter{w})
}

func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelAudit:
		return "AUDIT"
	default:
		return "UNKNOWN"
	}
}

func (lw LogWriter) Write(p []byte) (n int, err error) {
	return lw.w.Write(p)
}

func (l *Logger) Trace(msg string) {
	l.log(LevelTrace, msg)
}

func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

func (l *Logger) Warn(msg string) {
	l.log(LevelWarning, msg)
}

func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

func (l *Logger) Fatal(msg string) {
	l.log(LevelFatal, msg)
	// Because sometimes you just need to rage-quit
	os.Exit(1)
}

// Tracef is a Formatted versions of our logging methods
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logf(LevelTrace, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(LevelWarning, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(LevelFatal, format, args...)
	// Still rage-quitting, just with more style
	os.Exit(1)
}

func (l *Logger) StructuredLog(level LogLevel, msg string, data map[string]any) {
	if !l.shouldLog(level) {
		return
	}

	entry := structuredLogPool.Get().(*StructuredLogEntry)
	defer structuredLogPool.Put(entry)

	// Reset the entry
	entry.Timestamp = time.Now()
	entry.Level = level.String()
	entry.Message = msg
	for k, v := range data {
		entry.Data[k] = v
	}

	// Add caller info in debug level
	if level <= LevelDebug {
		if _, file, line, ok := runtime.Caller(2); ok {
			entry.Caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(entry); err != nil {
		l.Error(fmt.Sprintf("Failed to encode structured log: %v", err))
		return
	}

	writer := l.getWriter()
	if _, err := writer.Write(buf.Bytes()); err != nil {
		l.Error(fmt.Sprintf("Failed to write structured log: %v", err))
	}
}

// StructuredInfo is a Convenience methods for structured logging
func (l *Logger) StructuredInfo(msg string, data map[string]any) {
	l.StructuredLog(LevelInfo, msg, data)
}

func (l *Logger) StructuredDebug(msg string, data map[string]any) {
	l.StructuredLog(LevelDebug, msg, data)
}

func (l *Logger) StructuredError(msg string, data map[string]any) {
	l.StructuredLog(LevelError, msg, data)
}

// LogOnce Add a message to the log only once.
func (l *Logger) LogOnce(msg string) {
	// Get or create a sync.Once instance for this message
	once, _ := l.onceMessages.LoadOrStore(msg, &sync.Once{})

	// Use the sync.Once to ensure the message is logged only once
	once.(*sync.Once).Do(func() {
		l.Info(msg)
	})
}

// LogOncef Add a formatted version of the thing.
func (l *Logger) LogOncef(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.LogOnce(msg)
}

// ===================================================
// Private Functions Area
// ===================================================

func (l *Logger) writeColored(level LogLevel, msg string) {
	writer := l.getWriter()
	var err error

	switch level {
	case LevelTrace:
		_, err = l.colors.magenta.Fprint(writer, msg)
	case LevelDebug:
		_, err = l.colors.blue.Fprint(writer, msg)
	case LevelInfo:
		_, err = l.colors.cyan.Fprint(writer, msg)
	case LevelWarning:
		_, err = l.colors.yellow.Fprint(writer, msg)
	case LevelError, LevelFatal:
		_, err = l.colors.red.Fprint(writer, msg)
	default:
		_, err = fmt.Fprint(writer, msg)
	}

	if err != nil {
		// If we can't log, we're probably in trouble, but let's try one last time
		_, err := fmt.Fprintf(os.Stderr, "Logging error: %v\n", err)
		if err != nil {
			return
		}
	}
}

func newDefaultLogger() *Logger {
	l := &Logger{}
	l.level.Store(int32(LevelInfo))
	l.output.Store(LogWriter{os.Stdout})
	l.onceMessages = sync.Map{}
	l.colors.magenta = color.New(color.FgMagenta)
	l.colors.blue = color.New(color.FgBlue)
	l.colors.red = color.New(color.FgRed)
	l.colors.cyan = color.New(color.FgCyan)
	l.colors.yellow = color.New(color.FgYellow)
	return l
}

func (l *Logger) log(level LogLevel, msg string) {
	if !l.shouldLog(level) {
		return
	}

	prefix := l.getLevelPrefix(level)
	// Still keeping our emoji-based logging because we're not monsters
	l.writeColored(level, fmt.Sprintf("%s %s\n", prefix, msg))
}

func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}
	l.log(level, fmt.Sprintf(format, args...))
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return LogLevel(l.level.Load()) <= level
}

func (l *Logger) getWriter() LogWriter {
	return l.output.Load().(LogWriter)
}

func (l *Logger) getLevelPrefix(level LogLevel) string {
	switch level {
	case LevelTrace:
		return "[🛤️ TRACE]"
	case LevelDebug:
		return "[🐛 DEBUG]"
	case LevelInfo:
		return "[✅️ INFO]"
	case LevelWarning:
		return "[🚧 WARN]"
	case LevelError:
		return "[🚨 ERROR]"
	case LevelFatal:
		return "[💀 FATAL]"
	case LevelAudit:
		return "[🏋️‍♂️ AUDIT]"
	default:
		return "[🤷 UNKNOWN]"
	}
}
