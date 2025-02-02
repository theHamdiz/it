# it 🎭

Because we kinda need this shit daily and we know it!

[![GoDoc](https://godoc.org/github.com/theHamdiz/it?status.svg)](https://pkg.go.dev/github.com/theHamdiz/it)
[![Go Report Card](https://goreportcard.com/badge/github.com/theHamdiz/it)](https://goreportcard.com/report/github.com/theHamdiz/it)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Tests](https://img.shields.io/badge/tests-mostly%20passing-green)
![Panic Rate](https://img.shields.io/badge/panic%20rate-just%20right-blue)

## Installation 🚀

```bash
go get github.com/theHamdiz/it
```

## Core Features

### Error Handling (Because panic() is a lifestyle)

```go
// The "I believe in miracles" approach
config := it.Must(LoadConfig())

// The "let's log and pray" method
user := it.Should(GetUser())

// The "cover your tracks" strategy
err := it.WrapError(dbErr, "database had an existential crisis",
    map[string]any{"attempt": 42, "mood": "gothic"})
```

### Logging (Now with proper prefixes)

```go
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

## Sub-packages (For the Control Freaks)

### Pool - Object Recycling Center

```go
pool_ := pool_.NewPool(func() *ExpensiveObject {
    // Save the environment, reuse your objects
    return &ExpensiveObject{}
})
obj := pool_.Get()
// Return your shopping cart
defer pool_.Put(obj)
```

### Debouncer - Function Anger Management

```go
calm := debouncer.NewDebouncer(100 * time.Millisecond)
relaxedFunc := calm.Debounce(func() {
    // Now with less spam
    NotifyEveryone("Updates!")
})
```

### Load Balancer - Work Distribution Committee

```go
// Democratic work distribution
lb_ := lb.NewLoadBalancer(10)
err := lb_.Execute(ctx, func() error {
    // Share the pain
    return HeavyLifting()
})
```

### Result - Because null Checks Are So 1970s

```go
res := result_.Ok("success")
if res.IsOk() {
    value := res.UnwrapOr("plan B")
}
```

### Math - For The Algorithmically Gifted

```go
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

## Configuration 🔧

```go
// Redirect logs to your favorite black hole
file, _ := os.OpenFile("void.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
it.SetLogOutput(file)

// Set log level from "meh" to "everything's on fire"
it.SetLogLevel(it.LevelDebug)

// Let environment variables do the heavy lifting
// Optimist mode
os.Setenv("LOG_LEVEL", "PANIC")
// The ultimate backup strategy
os.Setenv("LOG_FILE", "/dev/null")
it.InitFromEnv()
```

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

[The Whatever, Just Take It License](LICENSE) - Because even chaos needs a license.

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
