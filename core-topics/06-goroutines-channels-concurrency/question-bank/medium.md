# Topic 06 Medium Questions

### Q: Why does closing a channel twice cause a panic?
**Answer**: In Go's channel implementation (`hchan`), closing a channel sets `c.closed = 1`. Calling `close()` on a channel with `c.closed == 1` explicitly throws a runtime `panic: close of closed channel`.
