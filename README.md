# it 🎭

Because we kinda need this shit daily and we know it!

[![GoDoc](https://godoc.org/github.com/theHamdiz/it?status.svg)](https://pkg.go.dev/github.com/theHamdiz/it)
[![Go Report Card](https://goreportcard.com/badge/github.com/theHamdiz/it)](https://goreportcard.com/report/github.com/theHamdiz/it)
[![License: The Whatever, Just Take It License](https://img.shields.io/badge/License-MIT-yellow.svg)](TRUELICENSE)
![Tests](https://img.shields.io/badge/tests-mostly%20passing-green)
![Panic Rate](https://img.shields.io/badge/panic%20rate-just%20right-blue)

## Installation 🚀

```bash
go get github.com/theHamdiz/it
```

## Core Features

### Error Handling (Because panic() is a lifestyle)

```go
import "github.com/theHamdiz/it"

// The "I believe in miracles" approach
config := it.Must(LoadConfig())

// The "let's log and pray" method
user := it.Should(GetUser())

// Maybe we'll connect to the database... eventually
maybeConnect := it.Could(func() (*sql.DB, error) {
    return sql.Open("postgres", "postgres://procrastinator:later@localhost")
})

// Sometime later, when we feel ready...
db := maybeConnect() // First call actually does the work
alsoDb := maybeConnect() // Returns same result, wasn't worth trying again anyway

if value, ok := it.Might(maybeThisWorks); ok {
    // Nice, we got something
} else {
    // No biggie, we weren't counting on it anyway
}

// The "cover your tracks" strategy
err := it.WrapError(dbErr, "database had an existential crisis",
    map[string]any{"attempt": 42, "mood": "gothic"})
```

### SafeGo - Panic-Proof Goroutines

Because letting goroutines die alone in the dark is just cruel.

```go
import "github.com/theHamdiz/it"

// Fire and forget (but not really)
it.SafeGo(func() {
    DoSomethingDangerous()  // We'll catch you if you fall
})

// For the context-aware crowd
it.SafeGoWithContext(ctx, func(ctx context.Context) {
    ResponsiblyDangerous(ctx)  // Safety first, but make it contextual
})
```

Includes automatic panic recovery and proper context propagation. Perfect for when you want to live dangerously but with a safety net.

Now go forth and spawn goroutines without fear of them taking your program down with them.


### Logging (Now with proper prefixes)

```go
import "github.com/theHamdiz/it"

it.Trace("Like println() but fancier")
it.Debug("For when you're feeling verbose")
it.Info("FYI: Something happened")
it.Warn("Houston, we have a potential problem")
it.Error("Everything is fine 🔥")
it.Audit("For when legal is watching")

// Structured logging (because JSON makes everything enterpriseTM)
it.StructuredInfo("API Call", map[string]any{
    "status": 200,
    "response_time": "too long",
    "excuse": "network congestion"
})
```

## Sub-Packages (For the Control Freaks)

### Pool - Object Recycling Center

```go
import "github.com/theHamdiz/it/pool"

pool_ := pool.NewPool(func() *ExpensiveObject {
    // Save the environment, reuse your objects
    return &ExpensiveObject{}
})
obj := pool_.Get()
// Return your shopping cart
defer pool_.Put(obj)
```

### Debouncer - Function Anger Management

```go
import "github.com/theHamdiz/it/debouncer"

calm := debouncer.NewDebouncer(100 * time.Millisecond)
relaxedFunc := calm.Debounce(func() {
    // Now with less spam
    NotifyEveryone("Updates!")
})
```

### Load Balancer - Work Distribution Committee

```go
import "github.com/theHamdiz/it/lb"

// Democratic work distribution
lb_ := lb.NewLoadBalancer(10)
err := lb_.Execute(ctx, func() error {
    // Share the pain
    return HeavyLifting()
})
```

### Benchmarker - The Performance Theater

Because measuring performance makes you feel better about your terrible code.

```go
import "github.com/theHamdiz/it/bm"

// Run your function until the numbers look good
result := bm.Benchmark("definitely-fast", 1000, func() {
    SuperOptimizedCode() // yeah right
})

// Or collect your own timings when you don't trust us
timings := []time.Duration{...}
result := bm.AnalyzeBenchmark("probably-fast", timings)

// Slack-friendly output for bragging rights
fmt.Println(result) // Screenshots or it didn't happen
```

Gives you min, max, average, median, and that standard deviation thing nobody really understands. Perfect for making your code look faster than it is in production.

JSON support included because apparently everything needs to be in a dashboard these days.

Now go forth and benchmark the hell out of that O(n2) algorithm you're trying to justify.

### Exponential Backoff - Professional Failure Management

Because at some point, everything fails. Might as well be ready for it.

```go
import "github.com/theHamdiz/it/retry"

// For when you're feeling optimistic
cfg := retry.DefaultRetryConfig()

// For when you know it's going to be rough
cfg := retry.Config{
    Attempts:     5,
    InitialDelay: time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,
    RandomFactor: 0.1,
}

// Actually doing the thing
result, err := retry.WithBackoff(ctx, cfg, func(ctx context.Context) (string, error) {
    return CallThatFlakyService()
})
```

Implements exponential backoff with jitter because hammering a failing service is both rude and stupid. Comes with sane defaults for when you're too lazy to think.

Default config gives you 3 attempts with increasing delays, starting at 100ms and capping at 10s. Add some randomness to avoid the thundering herd, because your systems have enough problems already.

Now go forth and embrace failure like a professional.

### Shutdown Manager - Graceful Program Retirement

Because even software needs a dignified exit strategy.

```go
import "github.com/theHamdiz/it/sm"

// Set up the end times
sm_ := sm.NewShutdownManager()

// Add some last wishes
sm_.AddAction(
    "save-data",
    func(ctx context.Context) error {
        return db.Close()
    },
    5 * time.Second,
    true, // This one matters
)

// Start watching for the end
sm_.Start()

// Wait for the inevitable
if err := sm_.Wait(); err != nil {
    log.Fatal("Failed to die gracefully:", err)
}
```

```go
// Or just use the global shortcut (for the pragmatists)
import "github.com/theHamdiz/it"

it.GracefulShutdown(ctx, server, 30*time.Second, done, func() {
    // Your last words here
})

// Need a phoenix-like rebirth?
it.GracefulRestart(ctx, server, 30*time.Second, done, func() {
    // Rise from the ashes
})
```

Handles shutdown signals (SIGINT, SIGTERM by default), manages cleanup tasks with timeouts, and ensures your program dies with dignity instead of just crashing.

Critical actions fail the whole shutdown if they fail, non-critical ones just log and continue, because some things aren't worth dying twice over.

Perfect for when you need your program to clean up after itself instead of leaving a mess for the OS to deal with.

### Rate Limiter - Traffic Control for Functions

```go
// The proper way (for when you need full control)
import "github.com/theHamdiz/it/rl"

limiter := rl.NewRateLimiter(time.Second, 10)
defer limiter.Close()

err := limiter.Execute(ctx, func() error {
    return DoSomethingFast()  // But not too fast
})

// Or better yet, with actual returns
result, err := rl.ExecuteRateLimited(limiter, ctx, func() (string, error) {
    return GetSomethingQuickly()  // Responsibly quick
})
```

```go
// The shortcut (for the rest of us)
import "github.com/theHamdiz/it"

// Turn any function into a well-behaved citizen
chillVersion := it.RateLimiter(time.Second, func() string {
    return IWouldSpamThisIfICould()
}).(func() string)

// Usage:
result := chillVersion()  // Now properly paced

// Waiting (but not forever)
success := it.WaitFor(5*time.Second, func() bool {
    return systemIsReady()  // Are we there yet?
})

if !success {
    // Time to give up and go home
}
```

Perfect for when your functions need to learn some self-control, without all the ceremony.

Now go forth and rate limit responsibly. Your servers will thank you.

### Time Keeper - Time Tracking for the Obsessed

Because if you're not measuring it, you're just guessing.

```go
import "github.com/theHamdiz/it/tk"

// Basic timing
tk_ := tk.NewTimeKeeper("database-query").Start()
defer tk_.Stop()

// With a callback for the micromanagers
tk_ := tk.NewTimeKeeper("expensive-operation",
    tk.WithCallback(func(d time.Duration) {
        metrics.Record("too-slow", d)
    }),
).Start()

// Quick one-liner for the lazy
result := tk.TimeFn("important-stuff", func() string {
    return DoTheWork()
})

// Async timing for the parallel obsessed
atk := tk.NewAsyncTimeKeeper("batch-process")
for task := range tasks {
    atk.Track(func() { process(task) })
}
durations := atk.Wait()
```

```go
// The dirty shortcut approach!!
import "github.com/theHamdiz/it"

result := it.TimeFunction("critical-operation", func() string {
    return DoSomethingWorthTiming()
})

// Block timing with defer
defer it.TimeBlock("expensive-stuff")()

// Custom timing callback
it.TimeFunctionWithCallback("important-task",
    func() string {
        return ExpensiveOperation()
    },
    func(d time.Duration) {
        if d > time.Second {
            alertSomeone()  // Someone's being slow
    },
)

// Parallel timing (for the impatient)
durations := it.TimeParallel("batch-process",
    func() { task1() },
    func() { task2() },
    func() { task3() },
)
// Now you know which one to blame
```

Times your operations, logs the results, and lets you obsess over performance with minimal effort. Supports both synchronous and async operations, because sometimes you need to time multiple failures simultaneously.

Perfect for when you need to prove that it's not your code that's slow, it's everyone else's.

### Circuit Breaker - Your Code's Emotional Support System

Because sometimes your dependencies need a time-out, just like your ex.

```go
import "github.com/theHamdiz/it/cb"

// Create a breaker that gives up after 3 failures
// and needs 30 seconds of alone time
breaker := cb.NewCircuitBreaker(3, 30*time.Second)

// Wrap your flaky calls with trust issues
err := breaker.Execute(func() error {
    return ThatThingThatAlwaysBreaks()
})

// Check if we're still in therapy
if breaker.IsOpen() {
    // Time to implement plan B
}
```

Includes state tracking, configurable thresholds, and automatic recovery. Perfect for when your microservices are having a midlife crisis.

Now go forth and fail gracefully, because that's what mature code does.

### Result - Because nil Checks Are So 1970s

```go
import "github.com/theHamdiz/it/result"

// For the optimists
res := result.Ok("success")
if res.IsOk() {
    value := res.UnwrapOr("plan B")
}

// For the realists
res := result.Try(func() string {
    return maybeExplode()
}).OrElse(func() Result[string] {
    return Ok("at least we tried")
})

// For the functional programming enthusiasts
result.Ok(42).
    Filter(isNotTooLarge).
    Map(addOne).
    Match(
        func(n int) { fmt.Println("Yay:", n) },
        func(err error) { /* pretend this never happened */ },
    )

// For those who like to live dangerously
value := result.Some(42).UnwrapOrPanic() // YOLO

// For the overachievers
pairs := result.Zip(
    result.Ok("hello"),
    result.Ok(42),
) // Because one value isn't complicated enough

// For when you need to fail collectively
results := result.Collect([]Result[int]{
    Ok(1), Ok(2), Err[int](errors.New("oops"))
}) // Misery loves company
```

Because if you're going to handle errors, you might as well do it with style.
Now with 100% more Option types, because nil was getting lonely.

### Math - For The Algorithmically Gifted

```go
import "github.com/theHamdiz/it/math"

// O(1) summation that would make Gauss proud
sum := math.Sum(1000000)

// Want to sum a specific range? We've got you covered
rangeSum := math.SumRange(42, 100)

// Need overflow protection? We're responsible adults here
safeSum, err := math.SumWithOverflowCheck(1000000)

// Living dangerously? MustSum will panic if things go wrong
yoloSum := math.MustSum(1000000)

// Sum of squares in O(1) because why not?
squares := math.SumOfSquares(100)

// Sum of cubes, because squares are so last century
cubes := math.SumOfCubes(50)

// Fourth powers, for when you really want to show off
fourthPowers := math.SumOfFourthPowers(25)

// Arithmetic series for the classically inclined
arithmetic := math.ArithmeticSeries(1, 2, 100) // 1, 3, 5, ...

// Geometric series for the exponentially minded
geometric := math.GeometricSeries(1, 2, 10) // 1, 2, 4, 8, ...

// Fast exponentiation because we're not savages
power := math.Pow(2, 10)

// Factorial without the stack overflow drama
factorial, err := math.Factorial(10)

// Need an approximation? Stirling's got your back
approxFact := math.FactorialStirlingApprox(100)

// Binomial coefficients without the tears
choose, err := math.Binomial(20, 10)

// Fibonacci that won't make your CPU cry
fib := math.Fibonacci(42)
```

Ultimate Security Suite – When “Good Enough” Isn’t Good Enough

```go
import "github.com/theHamdiz/it"

// When you need a secret that's totally random*:
secret := it.GenerateSecret(32)
// *Usually uses crypto/rand, but if that fails... well, we get creative with time.
// It's like using your birthday as a password, but with nanoseconds. Security through obscurity!

// When your password is too lazy to protect itself:
hashed, err := it.HashPassword("mySuperSecret", 12)
// Your password is sent to a rigorous bootcamp (bcrypt rounds), emerging as a hardened hash with its own unique salt.
// If the bootcamp fails, you'll get a polite error message.

// Think your password can waltz past the velvet rope?
err = it.VerifyPassword(hashed, "mySuperSecret")
// If err is nil, congratulations—your password made the cut.
// Otherwise, it's like a bouncer telling you, "Not on the list, buddy."
```

Now go forth and generate cryptographically convincing secrets, hash those passwords like they’re training for a marathon, and verify them with the confidence of a seasoned doorman. Enjoy your Ultimate Security Suite—because sometimes, even security needs a little swagger.

### Config - Because Hardcoding is a Crime

For when you need to make your application configurable, but still predictably unreliable.

```go
import (
	"github.com/theHamdiz/it/logger"
	"github.com/theHamdiz/it/cfg"
)

// You could use global shortcuts
// Redirect logs to your favorite black hole
file, _ := os.OpenFile("void.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
it.SetLogOutput(file)

// Set log level from "meh" to "everything's on fire"
it.SetLogLevel(logger.LevelDebug)

// Let environment variables do the heavy lifting
// Optimist mode
os.Setenv("LOG_LEVEL", "PANIC")
// The ultimate backup strategy
os.Setenv("LOG_FILE", "/dev/null")
it.InitFromEnv()

// Or you could create a config with sensible* defaults
// *sensible is a relative term
cfg_ := cfg.Configure(
    cfg.WithLogLevel(logger.LevelDebug),    // Maximum verbosity
    cfg.WithLogFile("regrets.log"),         // For posterity
    cfg.WithShutdownTimeout(5*time.Second), // Ain't nobody got time for that
    cfg.WithColors(true),                   // Pretty errors are still errors
)

// Check what you did to yourself
if cfg_.ColorsEnabled() {
    // Congratulations, your logs are now fabulous
}
```

Includes functional options, reasonable defaults, and just enough flexibility to be dangerous. Perfect for when you need to explain why production is different from your laptop.

Now go forth and configure responsibly, or don't. We're not your parents.


### Version - Software Identity Management

Because every program needs to know who it is and who to blame.

```go
// Get all the details
info := version.Get()
fmt.Println(info)
// Output: v1.2.3 built at 2023-12-25T12:00:00Z from main@abc1234 (go1.21) running on linux/amd64 in production

// For your JSON APIs
jsonData, _ := json.Marshal(info.ToMap())

// Build with:
go build -ldflags="
    -X 'package/version.version=1.2.3'
    -X 'package/version.buildTime=$(date -u +"%Y-%m-%dT%H:%M:%SZ")'
    -X 'package/version.gitCommit=$(git rev-parse HEAD)'
    -X 'package/version.gitBranch=$(git rev-parse --abbrev-ref HEAD)'
    -X 'package/version.environment=production'"
```

Tracks version, build time, git info, and runtime details. Perfect for logging, debugging, and finding out which commit broke production.

Values are injected at build time via ldflags, because hardcoding versions is for amateurs.

## Performance Notes

- Everything is O(1)*
- Memory usage is optimal**
- CPU friendly***

\* Except when it isn't.

\**Compared to loading the entire Wikipedia into RAM.

\*** Your definition of friendly may vary LOL!


## Known Features 🐛

- Sometimes panics exactly when you expect
- Occasionally logs the right thing
- Works perfectly in production*

\* Results may vary

## FAQ 🤔

**Q: Is this production ready?**
A: Define "production" and "ready"

**Q: Why should I use this?**
A: Because writing your own boilerplate is so 2020

**Q: Is it fast?**
A: Faster than your last deployment rollback

## Contributing 🤝

1. Fork (the repo, not your codebase)
2. Create (bugs)
3. Submit PR (with tests maybe?)
4. Wait (like your HTTP requests)

## License

[The Whatever, Just Take It License](TRUELICENSE) - Because even chaos needs a license.

> P.S. Don't believe the `MIT LICENSE` File, it's just here for `go.dev`

---

## Actual Serious Note

This package provides robust utilities for:
- Error handling and recovery.
- Structured and leveled logging.
- Some benchmarking functionalities.
- Rate limiting, load balancing & a debouncer.
- Some time keeping & measuring functionality.
- Object pooling and resource management.
- Mathematical optimizations - Some O(1) goodies.
- Graceful shutdowns & restarts.
- Some version management/reporting stuff.
- Oh yeah & a circuit breaker, whatever that might be.

Documentation: [GoDoc](https://pkg.go.dev/github.com/theHamdiz/it)

---

Now go write some code that might actually work, or not!

> *No functions were harmed in the making of this package*
