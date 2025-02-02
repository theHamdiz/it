package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/theHamdiz/it/logger"
)

type testWriter struct {
	bytes.Buffer
}

// Helper function to create a new test logger
func newTestLogger() (*logger.Logger, *testWriter) {
	buf := &testWriter{}
	logger_ := logger.NewLoggerWithLevelAndOutput(logger.LevelTrace, buf)
	return logger_, buf
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    logger.LogLevel
		expected string
	}{
		{logger.LevelTrace, "TRACE"},
		{logger.LevelDebug, "DEBUG"},
		{logger.LevelInfo, "INFO"},
		{logger.LevelWarning, "WARNING"},
		{logger.LevelError, "ERROR"},
		{logger.LevelFatal, "FATAL"},
		{logger.LevelAudit, "AUDIT"},
		{logger.LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if got := test.level.String(); got != test.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestLogger_BasicLogging(t *testing.T) {
	logger_, buf := newTestLogger()

	tests := []struct {
		name     string
		logFunc  func(string)
		message  string
		contains string
	}{
		{"Trace", logger_.Trace, "trace message", "TRACE"},
		{"Debug", logger_.Debug, "debug message", "DEBUG"},
		{"Info", logger_.Info, "info message", "INFO"},
		{"Warn", logger_.Warn, "warn message", "WARN"},
		{"Error", logger_.Error, "error message", "ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			test.logFunc(test.message)
			output := buf.String()
			if !strings.Contains(output, test.contains) {
				t.Errorf("Expected output to contain %q, got %q", test.contains, output)
			}
			if !strings.Contains(output, test.message) {
				t.Errorf("Expected output to contain %q, got %q", test.message, output)
			}
		})
	}
}

func TestLogger_FormattedLogging(t *testing.T) {
	logger_, buf := newTestLogger()

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		format   string
		args     []interface{}
		contains string
	}{
		{"Tracef", logger_.Tracef, "trace %s", []interface{}{"formatted"}, "TRACE"},
		{"Debugf", logger_.Debugf, "debug %s", []interface{}{"formatted"}, "DEBUG"},
		{"Infof", logger_.Infof, "info %s", []interface{}{"formatted"}, "INFO"},
		{"Warnf", logger_.Warnf, "warn %s", []interface{}{"formatted"}, "WARN"},
		{"Errorf", logger_.Errorf, "error %s", []interface{}{"formatted"}, "ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			test.logFunc(test.format, test.args...)
			output := buf.String()
			if !strings.Contains(output, test.contains) {
				t.Errorf("Expected output to contain %q, got %q", test.contains, output)
			}
			expectedMessage := fmt.Sprintf(test.format, test.args...)
			if !strings.Contains(output, expectedMessage) {
				t.Errorf("Expected output to contain %q, got %q", expectedMessage, output)
			}
		})
	}
}

func TestLogger_StructuredLogging(t *testing.T) {
	logger_, buf := newTestLogger()

	testData := map[string]interface{}{
		"user_id": 123,
		"action":  "test",
	}

	tests := []struct {
		name    string
		logFunc func(string, map[string]interface{})
		level   string
	}{
		{"StructuredInfo", logger_.StructuredInfo, "INFO"},
		{"StructuredDebug", logger_.StructuredDebug, "DEBUG"},
		{"StructuredError", logger_.StructuredError, "ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			test.logFunc("test message", testData)

			var entry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if entry["level"] != test.level {
				t.Errorf("Expected level %q, got %q", test.level, entry["level"])
			}
			if entry["message"] != "test message" {
				t.Errorf("Expected message %q, got %q", "test message", entry["message"])
			}

			data, ok := entry["data"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected data field to be a map")
			}
			if data["user_id"].(float64) != float64(testData["user_id"].(int)) {
				t.Errorf("Expected user_id %v, got %v", testData["user_id"], data["user_id"])
			}
			if data["action"] != testData["action"] {
				t.Errorf("Expected action %v, got %v", testData["action"], data["action"])
			}
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	buf := &testWriter{}

	tests := []struct {
		setLevel  logger.LogLevel
		logLevel  logger.LogLevel
		shouldLog bool
	}{
		{logger.LevelError, logger.LevelTrace, false},
		{logger.LevelError, logger.LevelDebug, false},
		{logger.LevelError, logger.LevelInfo, false},
		{logger.LevelError, logger.LevelWarning, false},
		{logger.LevelError, logger.LevelError, true},
		{logger.LevelError, logger.LevelFatal, true},
		{logger.LevelTrace, logger.LevelTrace, true},
		{logger.LevelTrace, logger.LevelDebug, true},
		{logger.LevelTrace, logger.LevelInfo, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("SetLevel_%s_LogLevel_%s", test.setLevel, test.logLevel), func(t *testing.T) {
			buf.Reset()
			logger_ := logger.NewLoggerWithLevelAndOutput(test.setLevel, buf)
			logger_.StructuredLog(test.logLevel, "test message", nil)

			hasOutput := buf.Len() > 0
			if hasOutput != test.shouldLog {
				t.Errorf("Expected shouldLog=%v, got output=%v", test.shouldLog, hasOutput)
			}
		})
	}
}

// SyncWriter is a thread-safe writer for testing
type SyncWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sw *SyncWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.buf.Write(p)
}

func (sw *SyncWriter) String() string {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.buf.String()
}

func (sw *SyncWriter) Reset() {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.buf.Reset()
}

func TestLogger_Concurrent(t *testing.T) {
	syncBuf := &SyncWriter{}
	logger_ := logger.NewLoggerWithLevelAndOutput(logger.LevelTrace, syncBuf)

	const goroutines = 100
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	messagesChan := make(chan string, goroutines*messagesPerGoroutine)

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := fmt.Sprintf("Message from goroutine %d: %d", id, j)
				logger_.Info(msg)
				messagesChan <- msg
			}
		}(i)
	}

	wg.Wait()
	close(messagesChan)

	sentMessages := make(map[string]bool)
	for msg := range messagesChan {
		sentMessages[msg] = true
	}

	output := syncBuf.String()
	receivedCount := 0
	for msg := range sentMessages {
		if strings.Count(output, msg) > 0 {
			receivedCount++
		}
	}

	expectedMessages := goroutines * messagesPerGoroutine
	if receivedCount != expectedMessages {
		t.Errorf("Expected %d unique messages, got %d", expectedMessages, receivedCount)
		t.Errorf("Test completed in %v", time.Since(start))
	} else {
		t.Logf("Successfully processed %d messages in %v", expectedMessages, time.Since(start))
	}
}

func TestDefaultLogger(t *testing.T) {
	logger_ := logger.DefaultLogger()
	if logger_ == nil {
		t.Error("DefaultLogger() returned nil")
	}
}

func TestNewLoggerWithLevelAndOutput(t *testing.T) {
	buf := &testWriter{}
	logger_ := logger.NewLoggerWithLevelAndOutput(logger.LevelDebug, buf)

	if logger_ == nil {
		t.Error("NewLoggerWithLevelAndOutput() returned nil")
	}

	logger_.Debug("test message")
	if buf.String() == "" {
		t.Error("Logger failed to write to buffer")
	}
}
