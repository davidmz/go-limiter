# Limiter

The `limiter` package provides a resource limiter for controlling access to a shared resource within a specified interval.

## Usage

See https://pkg.go.dev/github.com/davidmz/go-limiter for more information about package API and usage.

```go
import "github.com/davidmz/go-limiter"

// Create a new limiter instance with the desired interval:
limiter := limiter.New(1 * time.Second)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // The Take method returns true if access is granted within the specified
    // timeout, and false otherwise
    if !limiter.Take(3 * time.Second) {
        http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
        return
    }

    // Perform the resource operation
    fmt.Fprintf(w, "Hello, World!")
})

http.ListenAndServe(":8080", nil)

```
