# it

Utility Functions for Error Handling and Logging in Go

[![GoDoc](https://pkg.go.dev/badge/github.com/theHamdiz/it)](https://pkg.go.dev/github.com/theHamdiz/it)
[![Go Report Card](https://goreportcard.com/badge/github.com/theHamdiz/it)](https://goreportcard.com/report/github.com/theHamdiz/it)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
    - [Error Handling Functions](#error-handling-functions)
    - [Logging Functions](#logging-functions)
    - [Structured Logging](#structured-logging)
    - [Advanced Usage](#advanced-usage)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Overview

**it** is a Go package that provides utility functions for error handling and logging, simplifying common patterns while adhering to Go's best practices. It offers a collection of functions to manage errors by panicking on unrecoverable errors, logging errors while continuing execution, and other utilities for robust error handling.

## Features

- Simplified error handling with `Must` and `Should`
- Logging functions for different log levels: `Trace`, `Debug`, `Info`, `Warn`, `Error`, and `Fatal`
- Structured logging in JSON format for easy parsing and analysis
- Configurable log levels and outputs
- Color-coded console output for enhanced readability
- Utilities for error wrapping, panic recovery, and timing functions

## Installation

To install the `it` package, use `go get`:

```bash
go get github.com/theHamdiz/it
```

Import the package in your Go code:

```go
import "github.com/theHamdiz/it"
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/theHamdiz/it"
)

func main() {
    // Use Must for critical operations
    hardResult := it.Must(SomeCriticalFunction())

    // Use Should for operations where you can proceed on error
    softResult := it.Should(SomeNonCriticalFunction())

    fmt.Println("Hard Result:", hardResult)
    fmt.Println("Soft Result:", softResult)
}

func SomeCriticalFunction() (string, error) {
    // Simulate a critical function that may fail
    return "", fmt.Errorf("critical error occurred")
}

func SomeNonCriticalFunction() (string, error) {
    // Simulate a non-critical function that may fail
    return "default value", fmt.Errorf("non-critical error occurred")
}
```

## Usage Examples

### Error Handling Functions

#### `Must`

Use `Must` when an error is unrecoverable and should halt the program execution.

```go
result := it.Must(SomeFunction())
```

#### `Should`

Use `Should` when you want to log an error but continue execution.

```go
result := it.Should(SomeFunction())
```

### Logging Functions

#### Basic Logging

```go
it.Info("Application started")
it.Warn("Low disk space")
it.Error("Failed to connect to database")
```

#### Formatted Logging

```go
it.Infof("Server started on port %d", port)
it.Warnf("Disk space low: %d%% remaining", diskSpace)
it.Errorf("Error %d: %s", errorCode, errorMessage)
```

#### Debug and Trace Logging

Set the log level to include debug and trace messages:

```go
it.SetLogLevel(it.LevelDebug) // or it.LevelTrace
```

Log messages:

```go
it.Debug("Cache initialized")
it.Trace("Entered function X")
```

### Structured Logging

#### Structured Info

Log messages in JSON format with additional data:

```go
userData := map[string]string{
    "username": "johndoe",
    "ip":       "192.168.1.1",
}
it.StructuredInfo("User logged in", userData)
```

**Output:**

```json
{"level":"INFO","message":"User logged in","data":{"username":"johndoe","ip":"192.168.1.1"}}
```

### Advanced Usage

#### Setting Log Output

Redirect logs to a file:

```go
file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
    it.Fatalf("Failed to open log file: %v", err)
}
defer file.Close()
it.SetLogOutput(file)
```

#### Error Wrapping

Add context to errors:

```go
func readFile(filename string) error {
    _, err := os.Open(filename)
    if err != nil {
        return it.WrapErrorf(err, "failed to open file %s", filename)
    }
    return nil
}
```

#### Panic Recovery

Gracefully handle panics:

```go
func main() {
    defer it.RecoverPanic()
    // Code that may panic
    panic("unexpected error")
}
```

#### Timing Functions

Measure function execution time:

```go
it.TimeFunction("compute", compute)
```

Measure code block execution time:

```go
defer it.TimeBlock("main")()
// Code block to measure
```

## Configuration

### Setting Log Level

Control the verbosity of logs:

```go
it.SetLogLevel(it.LevelInfo) // Default level
it.SetLogLevel(it.LevelDebug)
```

Available levels:

- `it.LevelTrace`
- `it.LevelDebug`
- `it.LevelInfo`
- `it.LevelWarn`
- `it.LevelError`
- `it.LevelFatal`

### Environment Variables

Initialize logger settings from environment variables:

```go
os.Setenv("LOG_LEVEL", "DEBUG")
os.Setenv("LOG_FILE", "app.log")
it.InitFromEnv()
```

Supported variables:

- `LOG_LEVEL`: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`
- `LOG_FILE`: Path to a file for log output

## Documentation

For detailed documentation of all functions, visit the [GoDoc page](https://pkg.go.dev/github.com/theHamdiz/it).

## Contributing

Contributions are welcome! Please submit issues and pull requests for bug fixes, enhancements, or new features.

To contribute:

1. Fork the repository
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a pull request

Please ensure your code adheres to Go conventions and includes tests where appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Thank you for using **it**! If you have any questions or feedback, feel free to open an issue or submit a pull request.