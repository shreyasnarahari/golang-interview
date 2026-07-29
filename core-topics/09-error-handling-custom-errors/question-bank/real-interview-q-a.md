# Topic 09 Real Interview Q&A

### Q: What is the difference between `errors.Is` and `errors.As`?
**Answer**:
- `errors.Is(err, target)` compares errors for equivalence across a wrapped chain.
- `errors.As(err, &targetStruct)` extracts a specific custom error type from the wrapped chain.

### Q: How do you implement a custom error type that supports unwrapping?
**Answer**: Implement an `Unwrap() error` method on the custom error struct.
