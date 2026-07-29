# Topic 06 Hard Questions

### Q: Explain the internal `hchan` structure in the Go runtime.
**Answer**: `hchan` contains:
1. `qcount`: Current number of items in buffer.
2. `dataqsiz`: Circular buffer capacity.
3. `buf`: Pointer to circular array of elements.
4. `sendx` & `recvx`: Index offsets for circular buffer.
5. `recvq` & `sendq`: Wait queues (`waitq`) of blocked goroutines (`sudog`).
6. `lock`: Mutex protecting all operations on `hchan`.
