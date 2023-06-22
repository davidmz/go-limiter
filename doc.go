// Package limiter provides a resource limiter that restricts access to a shared resource.
//
// A Limiter ensures that access to the resource is granted no more frequently than once
// within a specified interval. Multiple clients (goroutines) can concurrently request
// access to the resource by invoking the Take method.
//
// The Take method waits until the resource becomes available and returns true to one of
// the clients. Other clients either continue to wait or receive false if the waiting time
// exceeds the specified timeout. If the resource is available at the time Take is called,
// one client immediately receives true, while the others wait for the next available time slot.
//
// Example usage:
//
//	limiter := NewLimiter(1 * time.Second)
//
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		if !limiter.Take(5 * time.Second) {
//			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
//			return
//		}
//
//		// Handling request normally
//		fmt.Fprintf(w, "Hello, World!")
//	})
package limiter
