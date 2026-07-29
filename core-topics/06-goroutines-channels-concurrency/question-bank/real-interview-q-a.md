# Topic 06 Real Interview Q&A

### Q: What happens when reading from a closed channel?
**Answer**: If buffered items remain, it returns those items. Once empty, reads immediately return the zero value of the element type and `ok = false`.

### Q: What happens when sending to a closed channel?
**Answer**: It panics immediately with `panic: send on closed channel`.
