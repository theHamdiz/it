// Package cfg  - Because hardcoding values is for people who live dangerously,
// and we're too scared for that
package cfg

import (
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/theHamdiz/it/logger"
	"github.com/theHamdiz/it/retry"
)

// Config holds all the knobs you can tweak until your application breaks
type Config struct {
	LogLevel        logger.LogLevel // How much spam you want in your logs
	LogFile         string          // Where your logs go to die
	ShutdownTimeout time.Duration   // How long before we kill it with fire
	RetryConfig     retry.Config    // For when at first you don't succeed
	EnableColors    bool            // Making logs pretty won't fix your bugs
}

// Our sensible* defaults
// *sensible is a relative term
var defaultConfig = Config{
	LogLevel:        logger.LevelInfo,           // Because DEBUG is too chatty
	ShutdownTimeout: 30 * time.Second,           // Plenty of time to panic
	RetryConfig:     retry.DefaultRetryConfig(), // Hope springs eternal
	EnableColors:    true,                       // Life's too short for monochrome
}

// ConfigOption - Because global variables are evil, but function pointers are fine
type ConfigOption func(*Config)

// Configure applies your questionable configuration choices
// Returns a Config that you'll probably need to change later anyway
func Configure(opts ...ConfigOption) *Config {
	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Let's actually use these settings (what could go wrong?)
	logger.SetLogLevel(cfg.LogLevel)
	if cfg.LogFile != "" {
		if file, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logger.SetLogOutput(file)
		}
	}
	color.NoColor = !cfg.EnableColors
	return &cfg
}

// WithLogLevel - For when you want to see more (or less) of your mistakes
func WithLogLevel(level logger.LogLevel) ConfigOption {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithLogFile - Because printing to stdout is so 1970s
func WithLogFile(file string) ConfigOption {
	return func(c *Config) {
		c.LogFile = file
	}
}

// WithShutdownTimeout - How patient are you really?
func WithShutdownTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.ShutdownTimeout = timeout
	}
}

// WithColors - Because monochrome logs are depressing enough already
func WithColors(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableColors = enable
	}
}

// WithRetryConfig - Configure your retry strategy, as if the first attempt wasn't bad enough
func WithRetryConfig(config retry.Config) ConfigOption {
	return func(c *Config) {
		c.RetryConfig = config
	}
}

// GetLogLevel - In case you forgot what you configured 5 minutes ago
func (c *Config) GetLogLevel() logger.LogLevel {
	return c.LogLevel
}

// GetShutdownTimeout - How long until we give up and kill -9
func (c *Config) GetShutdownTimeout() time.Duration {
	return c.ShutdownTimeout
}

// ColorsEnabled - Are we in fancy mode?
func (c *Config) ColorsEnabled() bool {
	return c.EnableColors
}

// GetRetryConfig - Returns your optimistic retry settings
func (c *Config) GetRetryConfig() retry.Config {
	return c.RetryConfig
}
