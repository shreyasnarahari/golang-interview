# Topic 08 Medium Questions

### Q: Why should context keys be custom unexported types instead of strings?
**Answer**: Using plain strings (e.g. `"userID"`) creates collision risks between packages sharing the same context. A custom type prevents key collisions across package boundaries:
```go
type keyType int
const userIDKey keyType = 1
```
