---
trigger: always_on
---

# Go Project Rules
---

## Go 1.25+ Features to Use

### testing/synctest Package (Stable in 1.25)
Test concurrent code with virtual time:
```go
import "testing/synctest"

func TestTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        
        // Time advances instantly when goroutines block
        select {
        case <-ctx.Done():
            // Happens immediately in test
        }
    })
}
```

### encoding/json/v2 (Experimental)
Enable with `GOEXPERIMENT=jsonv2`:
```go
// Streaming without intermediate encoders
jsonv2.MarshalWrite(w, data)
jsonv2.UnmarshalRead(r, &result)

// Better defaults: nil slices → [], faster decode
```

### Experimental GC (greenteagc)
Enable with `GOEXPERIMENT=greenteagc` for 10-40% reduction in GC overhead.

### FlightRecorder (1.25)
Lightweight execution trace capture:
```go
import "runtime/trace"

rec, _ := trace.NewFlightRecorder(trace.FlightRecorderConfig{})
defer rec.Close()

// On significant event, snapshot recent trace
f, _ := os.Create("trace.out")
rec.WriteTo(f, "")
```

### Container-Aware GOMAXPROCS (1.25)
GOMAXPROCS now dynamically adjusts to CPU limits in containers.

### WaitGroup.Go (1.25)
Simplified goroutine launching:
```go
var wg sync.WaitGroup
wg.Go(func() {
    // do work
})
wg.Wait()
```

### T.Attr for Test Metadata (1.25)
```go
func TestFeature(t *testing.T) {
    t.Attr("issue", "PROJ-1234")
    t.Attr("description", "Tests edge case handling")
}
```

### os.Root with ReadLinkFS (1.25)
Filesystem isolation with symlink support:
```go
root, _ := os.OpenRoot("/safe/directory")
defer root.Close()

fsys := root.FS().(fs.ReadLinkFS)
target, _ := fsys.ReadLink("symlink.txt")
```

### Core Types Removed from Generics (1.25)
Operations on type sets no longer require core types:
```go
func Slice[S ~[]byte | ~string](s S, i, j int) S {
    return s[i:j] // Now works!
}
```

### Generic Type Aliases (1.24)
Type aliases can now be parameterized:
```go
type ComparableVector[T comparable] = Vector[T]
type IntNode = Node[int]
```

### Tool Directives (1.24)
Track executable dependencies in `go.mod`:
```
tool golang.org/x/tools/cmd/stringer
```
Run with `go tool stringer`.

### JSON omitzero (1.24)
Use `omitzero` tag for cleaner JSON output:
```go
type Config struct {
    Name    string `json:"name"`
    Timeout int    `json:"timeout,omitzero"` // omits if zero value
}
```

### Weak Pointers and Cleanup (1.24)
Use `weak.Make` for cache-friendly weak references:
```go
import "weak"

w := weak.Make(&obj)
if v := w.Value(); v != nil {
    // use v
}
```

Use `runtime.AddCleanup` over `runtime.SetFinalizer`:
```go
runtime.AddCleanup(&obj, func(ptr *int) {
    // cleanup logic
}, &associated)
```

### Range Over Functions (1.23+)
Use iterator functions for custom iteration:
```go
func Backward[E any](s []E) func(func(int, E) bool) {
    return func(yield func(int, E) bool) {
        for i := len(s) - 1; i >= 0; i-- {
            if !yield(i, s[i]) {
                return
            }
        }
    }
}

for i, v := range Backward(slice) {
    // iterates in reverse
}
```

### Benchmark Looping (1.24)
Use `b.Loop()` for more predictable benchmarks:
```go
func BenchmarkFoo(b *testing.B) {
    for b.Loop() {
        // benchmarked code
    }
}
```

---

## Project Structure

```
myproject/
├── cmd/
│   └── myapp/
│       └── main.go           # Entry point, thin logic
├── internal/
│   ├── domain/               # Business logic
│   ├── repository/           # Data access
│   └── service/              # Application services
├── pkg/                      # Public library code (if any)
├── testdata/                 # Test fixtures
├── go.mod
├── go.sum
└── Makefile
```

### Key Principles
- `cmd/` contains only entry points with minimal logic
- `internal/` for private application code (enforced by Go)
- `pkg/` only if code is meant for external consumption
- One package per directory
- Package names: short, lowercase, no underscores

---

## Error Handling

### Always Handle Errors
```go
// WRONG
result, _ := doSomething()

// RIGHT
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}
```

### Error Wrapping
Use `%w` for error chains:
```go
if err != nil {
    return fmt.Errorf("failed to process %s: %w", name, err)
}
```

### Custom Error Types
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Check with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    // handle validation error
}
```

### Sentinel Errors
```go
var ErrNotFound = errors.New("not found")

// Check with errors.Is
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

---

## Code Patterns

### Context Usage
Always pass context as first parameter:
```go
func ProcessData(ctx context.Context, data []byte) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // process...
}
```

### Struct Initialization
Prefer named fields:
```go
// WRONG
user := User{"John", 30, true}

// RIGHT
user := User{
    Name:   "John",
    Age:    30,
    Active: true,
}
```

### Interface Design
Keep interfaces small and focused:
```go
// Good: single method interface
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Accept interfaces, return structs
func ProcessReader(r io.Reader) (*Result, error) {
    // ...
}
```

### Defer for Cleanup
```go
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    return io.ReadAll(f)
}
```

---

## Concurrency

### Goroutine Lifecycle
Always ensure goroutines can exit:
```go
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            process(job)
        }
    }
}
```

### Channel Patterns
```go
// Buffered channel for known workload
results := make(chan Result, len(items))

// Close channels from sender side
go func() {
    defer close(results)
    for _, item := range items {
        results <- process(item)
    }
}()
```

### sync.WaitGroup
```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()
```

### sync.Once for Lazy Init
```go
var (
    instance *Client
    once     sync.Once
)

func GetClient() *Client {
    once.Do(func() {
        instance = &Client{}
    })
    return instance
}
```

---

## Testing

### Table-Driven Tests
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -1, -2},
        {"zero", 0, 0, 0},
    }
    
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := Add(tc.a, tc.b)
            if result != tc.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", 
                    tc.a, tc.b, result, tc.expected)
            }
        })
    }
}
```

### Subtests with Cleanup
```go
func TestDatabase(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()
    })
    
    t.Run("Insert", func(t *testing.T) {
        // test insert
    })
    
    t.Run("Query", func(t *testing.T) {
        // test query
    })
}
```

### Testing with Context
```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(
        context.Background(), 
        5*time.Second,
    )
    defer cancel()
    
    result, err := SlowOperation(ctx)
    // assertions...
}
```

---

## Performance

### Preallocate Slices
```go
// WRONG
var results []string
for _, item := range items {
    results = append(results, process(item))
}

// RIGHT
results := make([]string, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

### strings.Builder for Concatenation
```go
var b strings.Builder
for _, s := range parts {
    b.WriteString(s)
}
result := b.String()
```

### sync.Pool for Frequent Allocations
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func Process() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    // use buf...
}
```

---

## Tooling Commands

```bash
# Format
go fmt ./...

# Lint
golangci-lint run

# Test with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build
go build -o bin/myapp ./cmd/myapp

# Update dependencies
go get -u ./...
go mod tidy

# Run tool dependencies
go tool <toolname>
```

---

## Common Mistakes to Avoid

1. **Naked returns in long functions** — only use in short functions
2. **init() abuse** — prefer explicit initialization
3. **Ignoring context cancellation** — always check `ctx.Done()`
4. **Unbounded goroutines** — use worker pools
5. **Data races** — use `-race` flag in tests
6. **Copying sync primitives** — pass by pointer
7. **Logging in libraries** — return errors, let caller log
