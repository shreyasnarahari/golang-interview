# Topic 15 Real Interview Q&A

### Q: What is the Go Race Detector and how do you run it?
**Answer**: A built-in compiler tool powered by ThreadSanitizer that detects data races at runtime. Run using `go test -race ./...` or `go run -race main.go`.

### Q: How do you generate code coverage reports in Go?
**Answer**: `go test -coverprofile=coverage.out ./...` followed by `go tool cover -html=coverage.out`.
