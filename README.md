## it

Utility Functions for Error Handling, Logging, and Retry Logic in Go

[![GoDoc](https://pkg.go.dev/badge/github.com/theHamdiz/it)](https://pkg.go.dev/github.com/theHamdiz/it)
[![Go Report Card](https://goreportcard.com/badge/github.com/theHamdiz/it)](https://goreportcard.com/report/github.com/theHamdiz/it)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

### Table of Contents

 • Overview  
 • Features  
 • Installation  
 • Quick Start  
 • Usage Examples  
 • Error Handling  
 • Logging  
 • Structured Logging  
 • Retry and Exponential Backoff  
 • Graceful Actions  
 • Rate Limiters  
 • Utility Functions  
 • Configuration  
 • Documentation  
 • Contributing  
 • License  

#### Overview

it is a Go package providing utility functions for error handling, logging, and execution retries, simplifying common patterns while adhering to Go best practices. It offers a collection of functions to manage errors, structured logging, retries with exponential backoff, and other robust utilities like graceful shutdown & restart of any server; with done notification as well as action functions to perform.

#### Features

 • Simplified error handling with Must and Should  
 • Logging functions for different log levels: Trace, Debug, Info, Warn, Error, and Fatal  
 • Structured logging in JSON format for easy parsing and analysis  
 • Configurable log levels, outputs, and color-coded console output  
 • Utility functions for retry mechanisms, exponential backoff, and context-based retries  
 • Enhanced error wrapping, panic recovery, and timing functions  

### Installation

To install the it package, use go get:

`go get github.com/theHamdiz/it`

Import the package in your Go code:

```go
import "github.com/theHamdiz/it"
```

### Quick Start

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
    return "", fmt.Errorf("critical error occurred")
}

func SomeNonCriticalFunction() (string, error) {
    return "default value", fmt.Errorf("non-critical error occurred")
}
```

### Usage Examples

#### Error Handling

#### Must

 Use Must when an error is unrecoverable and should halt the program execution.

```go
result := it.Must(SomeFunction())
```

#### Should

 Use Should when you want to log an error but continue execution.

```go
 result := it.Should(SomeFunction())
 ```

#### Ensure

 Panics if err is not nil. Use it when a critical error cannot be recovered.

```go
it.Ensure(SomeCriticalFunction())
```

#### Attempt

 Logs the error but continues execution. Use it for non-critical errors.

```go
it.Attempt(SomeFunction())
```

#### WrapError

 Adds contextual information to an error message, making it easier to track errors.

```go
err := it.WrapError(err, "processing file", map[string]string{"file": filename})
```

#### Logging

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

`Debug` and `Trace` Logging  

 Set the log level to include debug and trace messages:  

```go
it.SetLogLevel(it.LevelDebug)
```  

#### Log messages

```go
it.Debug("Cache initialized")
 
it.Trace("Entered function X")  
```

---

### LogStackTrace

Logs the current stack trace, which is helpful for debugging complex issues by displaying the call stack.  

```go
func LogStackTrace()
```

**Example Usage:**

```go
it.LogStackTrace()
```

---

### LogErrorWithStack

Logs an error along with the current stack trace. This provides more detailed information to aid in debugging by capturing the error context and call stack.  

```go
func LogErrorWithStack(err error)
```

- **`err`**: The error to log, along with its stack trace.

**Example Usage:**

```go
it.LogErrorWithStack(err)
```

#### LogOnce

 Logs a message only once, avoiding repetitive log entries in loops.  

```go
it.LogOnce("This message will only be logged once")
```  

#### Audit

 Logs an audit-specific message for tracking important actions.  

```go
 it.Audit("User login attempt recorded")
```  

#### Structured Logging

#### Structured Info

 Logs messages in JSON format with additional data.  

```go
userData := map[string]string{"username": "johndoe", "ip": "192.168.1.1"}
it.StructuredInfo("User logged in", userData)
```  

---

### StructuredDebug

Logs a structured debug-level message in JSON format, useful for detailed logging with additional contextual data.  

```go
func StructuredDebug(message string, data any)
```

- **`message`**: The debug message to log.  
- **`data`**: Additional contextual data in key-value format.  

**Example Usage:**

```go
it.StructuredDebug("Cache hit", map[string]string{"key": "user:1234"})
```

---

### StructuredWarning

Logs a structured warning message in JSON format with additional data. Useful for tracking non-critical issues in a structured way.  

```go
func StructuredWarning(message string, data any)
```

- **`message`**: The warning message to log.  
- **`data`**: Additional key-value data to provide context.  

**Example Usage:**

```go
it.StructuredWarning("High memory usage detected", map[string]interface{}{"usage": 95})
```

---

### StructuredError  

Logs an error message in JSON format with additional contextual data, useful for error tracking with structured logs.  

```go
func StructuredError(message string, data any)  
```

- **`message`**: The error message to log.  
- **`data`**: Additional context in a key-value format.  

**Example Usage:**  

```go
it.StructuredError("File not found", map[string]string{"filename": "config.yaml"})  
```

#### BufferedLogger  

 Logs to any specified writer with buffering, supporting os.Stdout, file, or custom io.Writer.  

```go
file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
logger := it.NewBufferedLogger(file)
logger.Log("Buffered log message")
logger.Flush()
```  

### Retry and Exponential Backoff  

#### Retry  

 Retries a function with a fixed delay. Useful for handling transient errors.  

`err := it.Retry(3, time.Second, SomeFunction)`  

#### RetryExponential  

 Retries a function with exponential backoff, doubling the delay after each attempt.  

`err := it.RetryExponential(5, time.Second, SomeFunction)`  

#### RetryWithContext  

 Retries a function with a fixed delay, stopping if the context is canceled.  

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := it.RetryWithContext(ctx, 3, time.Second, SomeFunction)
```

#### RetryExponentialWithContext  

 Retries a function with exponential backoff, doubling the delay after each attempt, but stops if the context is canceled.  

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)  
defer cancel()  
err := it.RetryExponentialWithContext(ctx, 5, time.Second, SomeFunction)  
```

---

### Graceful Actions

#### GracefulShutdown

`GracefulShutdown` listens for an interrupt signal (e.g., `SIGINT` or `SIGTERM`) and attempts to gracefully shut down the given server within the specified timeout. If an action function is provided, it will execute this function after shutdown completes. If a done channel is provided, it will signal completion on the channel after the shutdown and executing the action.

`func GracefulShutdown(ctx context.Context, server interface{ Shutdown(context.Context) error }, timeout time.Duration, done chan<- bool, action func())`

 • `ctx`: The base context for shutdown, which can be `context.Background()` or another `context`.  
 • `server`: The server object to shut down, which must have a Shutdown method that takes a `context.Context`.  
 • `timeout`: The maximum time to wait for the server to shut down gracefully before forcing termination.  
 • `done`: An optional channel to signal completion once the shutdown and action are complete. If done is nil, no notification is sent.  
 • `action`: An optional function to execute after the server has shut down. This can be used for cleanup or other post-shutdown tasks. If action is nil, no action is performed.

Example Usage:  

 1. Without a done channel or action:  

```go
it.GracefulShutdown(context.Background(), server, 5*time.Second, nil, nil)
```  

 2. With a done channel and action:  

```go
done := make(chan bool)
cleanupAction := func() {
    log.Println("Performing post-shutdown cleanup...")
    // Additional cleanup code here
}
go it.GracefulShutdown(context.Background(), server, 5*time.Second, done, cleanupAction)
<-done // Wait for the shutdown process to complete
```

---

#### GracefulRestart

`GracefulRestart` listens for a signal to restart the server gracefully. It attempts to shut down the given server within the specified timeout and then optionally performs an action before signaling completion on the done channel, if provided.

`func GracefulRestart(ctx context.Context, server interface{ Shutdown(context.Context) error }, timeout time.Duration, done chan<- bool, action func())`

 • `ctx`: The context for shutdown, typically `context.Background()` or similar.  
 • `server`: The server instance to restart, which must implement a Shutdown method.  
 • `timeout`: The maximum time allowed for the graceful shutdown before initiating a restart.  
 • `done`: An optional channel to signal completion once the shutdown, action, and restart are complete. If done is nil, no notification is sent.  
 • `action`: An optional function to execute after the server has shut down. This can be used to reinitialize services, reload configurations, or perform any other custom restart logic. If action is nil, no action is performed.  

Example Usage:  

 1. Without a done channel or action:  

```go
it.GracefulRestart(context.Background(), server, 5*time.Second, nil, nil)
```  

 2. With a done channel and an action:  

```go
done := make(chan bool)
restartAction := func() {
    log.Println("Performing custom restart actions...")
    // Additional initialization or setup code here
}
go it.GracefulRestart(context.Background(), server, 5*time.Second, done, restartAction)
<-done // Wait for the restart process to complete
```

#### RateLimiter

`RateLimiter` returns a rate-limited version of any `handler` function, allowing it to execute only once per specified rate interval. This function is designed to work with handlers of any type, including HTTP route handlers and custom functions with any signature.

`func RateLimiter(rate time.Duration, fn interface{}) interface{}`  

 • rate: The interval at which fn is allowed to execute. Only one call to fn will be allowed per rate period.  
 • fn: The original handler function to be rate-limited. It can have any number and type of input arguments and return any type.  

#### Returns  

A rate-limited function that has the same signature as fn. The returned function will wait for the specified rate interval before each execution.

#### Example Usage  

With an HTTP Handler  

```go
func myHTTPHandler(w http.ResponseWriter, r *http.Request) {
 w.Write([]byte("Hello, World!"))
}

limitedHandler := it.RateLimiter(1*time.Second, myHTTPHandler).(func(http.ResponseWriter, *http.Request))
http.HandleFunc("/limited", limitedHandler)
```

In this example, `myHTTPHandler` is wrapped with `RateLimiter`, and it will execute at most once per second when called.

With a Custom Function

```go
func customHandler(name string, age int) string {
 return fmt.Sprintf("Name: %s, Age: %d", name, age)
}

limitedCustomHandler := it.RateLimiter(2*time.Second, customHandler).(func(string, int) string)
result := limitedCustomHandler("Alice", 30)
fmt.Println(result) // Output: "Name: Alice, Age: 30"
```

Here, customHandler is rate-limited to execute at most once every two seconds, regardless of how often limitedCustomHandler is called.  

Notes  

 • Generic Handler Support: RateLimiter is designed to work with any function signature, making it compatible with various handler types.  
 • Type Assertion: Since RateLimiter returns an interface{}, you will need to use type assertion to call the wrapped function with the correct parameter and return types.  
 • Rate Control: Calls to the returned function are delayed based on the specified rate, ensuring that fn is only invoked once per interval.  

#### Utility Functions  

#### WaitFor  

 Waits until a specified condition is met or times out.  

```go
it.WaitFor(time.Second*10, func() bool { return someCondition() })
```  

#### DeferWithLog  

 Creates a deferred function that logs a message upon completion, helpful for complex defer chains.  

```go
defer it.DeferWithLog("Cleanup complete")()
```  

#### TimeFunction  

 Measures and logs the execution time of a function.  

```go
it.TimeFunction("compute", compute)
```  

#### TimeBlock  

 Starts a timer and logs the execution time of a code block.  

```go
defer it.TimeBlock("main")()
```  

#### Configuration  

#### Setting Log Output  

Redirect logs to a file or other output destination.  

```go
file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
it.SetLogOutput(file)
```  

#### Setting Log Level  

Control verbosity of logs:  

```go
it.SetLogLevel(it.LevelInfo)
```  

#### Available levels  

 • it.LevelTrace  
 • it.LevelDebug  
 • it.LevelInfo  
 • it.LevelWarn  
 • it.LevelError  
 • it.LevelFatal  
 • it.LevelAudit  

#### Environment Variables  

Initialize logger settings from environment variables:  

```go
os.Setenv("LOG_LEVEL", "DEBUG")
os.Setenv("LOG_FILE", "app.log")
it.InitFromEnv()
```

#### Supported variables  

 • LOG_LEVEL: TRACE, DEBUG, INFO, WARN, ERROR, FATAL  
 • LOG_FILE: Path to a file for log output  

#### Documentation  

For detailed documentation of all functions, visit the GoDoc page.  

#### Contributing  

Contributions are welcome! Please submit issues and pull requests for bug fixes, enhancements, or new features.  

 1. Fork the repository.  
 2. Create a new branch (git checkout -b feature/your-feature).  
 3. Commit your changes (git commit -am 'Add new feature').  
 4. Push to the branch (git push origin feature/your-feature).  
 5. Open a pull request.  

Please ensure your code adheres to Go conventions and includes tests where appropriate.  

#### License  

This project is licensed under the MIT License - see the LICENSE file for details.  

Thank you for using it! If you have any questions or feedback, feel free to open an issue or submit a pull request.  
