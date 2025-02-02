package cfg_test

import (
	"os"
	"testing"
	"time"

	"github.com/theHamdiz/it/cfg"
	"github.com/theHamdiz/it/logger"
)

func TestConfigure_WithLogLevel(t *testing.T) {
	newLogLevel := logger.LevelError
	cfg_ := cfg.Configure(cfg.WithLogLevel(newLogLevel))

	if cfg_.GetLogLevel() != newLogLevel {
		t.Errorf("Expected log level to be %v, got %v", newLogLevel, cfg_.GetLogLevel())
	}
}

func TestConfigure_WithLogFile(t *testing.T) {
	tempFile := "testlog.txt"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(tempFile)

	_ = cfg.Configure(cfg.WithLogFile(tempFile))

	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s to be created, but it was not", tempFile)
	}
}

func TestConfigure_WithShutdownTimeout(t *testing.T) {
	newTimeout := 45 * time.Second
	cfg_ := cfg.Configure(cfg.WithShutdownTimeout(newTimeout))

	if cfg_.GetShutdownTimeout() != newTimeout {
		t.Errorf("Expected shutdown timeout to be %v, got %v", newTimeout, cfg_.GetShutdownTimeout())
	}
}

func TestConfigure_WithColors(t *testing.T) {
	cfg_ := cfg.Configure(cfg.WithColors(false))
	if cfg_.ColorsEnabled() {
		t.Errorf("Expected colors to be disabled, but they are enabled")
	}

	cfg_ = cfg.Configure(cfg.WithColors(true))
	if !cfg_.ColorsEnabled() {
		t.Errorf("Expected colors to be enabled, but they are disabled")
	}
}

func TestConfigure_WithMultipleOptions(t *testing.T) {
	newLogLevel := logger.LevelDebug
	newLogFile := "testlog2.txt"
	newTimeout := 60 * time.Second
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Failed to remove temp file %s: %v", name, err)
		}
	}(newLogFile)

	cfg_ := cfg.Configure(
		cfg.WithLogLevel(newLogLevel),
		cfg.WithLogFile(newLogFile),
		cfg.WithShutdownTimeout(newTimeout),
		cfg.WithColors(false),
	)

	if cfg_.GetLogLevel() != newLogLevel {
		t.Errorf("Expected log level to be %v, got %v", newLogLevel, cfg_.GetLogLevel())
	}

	if _, err := os.Stat(newLogFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s to be created, but it was not", newLogFile)
	}

	if cfg_.GetShutdownTimeout() != newTimeout {
		t.Errorf("Expected shutdown timeout to be %v, got %v", newTimeout, cfg_.GetShutdownTimeout())
	}

	if cfg_.ColorsEnabled() {
		t.Errorf("Expected colors to be disabled, but they are enabled")
	}
}
