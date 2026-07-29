# Topic 11 Hard Questions

### Q: What is the `runnext` slot in Processor P?
**Answer**: `runnext` is a high-priority 1-element slot in **P**'s structure. When goroutine $G_1$ spawns $G_2$, $G_2$ is placed in `runnext` to execute next immediately, optimizing temporal locality and CPU cache affinity for pipeline communication.
