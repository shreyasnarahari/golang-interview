# Topic 15 Medium Questions

### Q: What is `TestMain(m *testing.M)` used for?
**Answer**: To execute custom global setup and teardown logic (e.g. starting a test database container, initializing global state) before running any unit tests in a package:
```go
func TestMain(m *testing.M) {
    setup()
    code := m.Run()
    teardown()
    os.Exit(code)
}
```
