# Topic 12 Medium Questions

### Q: What are the STW (Stop-The-World) phases in Go's concurrent GC?
**Answer**:
1. **Sweep Termination**: Brief STW pause to turn on write barrier and start mark phase.
2. **Mark Termination**: Brief STW pause to turn off write barrier and clear mark state.
Both pauses combined total less than 1 millisecond.
