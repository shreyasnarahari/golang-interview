# Topic 09 Medium Questions

### Q: What feature was introduced in Go 1.20 for joining multiple errors?
**Answer**: `errors.Join(err1, err2, ...)`. It combines multiple non-nil errors into a single composite error whose `Unwrap()` method returns a slice `[]error`.
