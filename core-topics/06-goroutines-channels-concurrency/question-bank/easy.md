# Topic 06 Easy Questions

### Q: How do you perform a non-blocking channel read?
**Answer**: Using a `select` statement with a `default` case:
```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("Channel not ready")
}
```
