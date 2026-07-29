# Topic 11 Real Interview Q&A

### Q: What is GOMAXPROCS and what does it default to?
**Answer**: `GOMAXPROCS` sets the number of logical **P** (Processor) contexts in the Go runtime. Defaults to the number of logical CPU cores on the host machine.

### Q: What is work stealing in the Go runtime?
**Answer**: When a Processor **P** empties its local run queue, it steals half the goroutines from another randomly selected **P**'s queue.
