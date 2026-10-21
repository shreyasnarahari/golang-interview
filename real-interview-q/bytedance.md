# ByteDance / TikTok — Company-Flavored Go Interview Questions & Deep Dives

> **Target Roles**: High-Concurrency Backend Engineer, RPC & Framework Developer, Video Pipeline Infrastructure Engineer.
>
> *These are company-flavored questions inspired by ByteDance's public Go stack (Kitex, Netpoll) and candidate reports — not verbatim interview transcripts.*

---

## 🏢 Why ByteDance Asks Go
ByteDance operates one of the largest Go microservice architectures in the world (powering TikTok/Douyin) and open-sourced **Kitex** (high-performance RPC framework) and **Netpoll** (epoll-based Go network library). This file covers **map bucket internals and `hmap` evacuation**, **interface (`eface`) boxing costs**, and **allocation elimination**; the goroutine worker-pool question commonly reported in ByteDance loops is covered in [uber.md](uber.md).

---

## 📌 Question 1: High-Performance Fixed-Memory Goroutine Pool

> This question is covered in the Uber question bank — see [uber.md, Question 3: Worker Pool Pattern with Graceful Shutdown](uber.md). The ByteDance-flavored variant emphasizes a **fixed memory footprint under extreme traffic spikes** (e.g. 500,000 req/s): bounded task queues and worker reuse instead of uncapped `go func()` spawning, so GC object counts stay flat. Use the Uber implementation as the base, and be ready to justify **backpressure** — rejecting or blocking submissions when the queue is full rather than letting goroutine counts grow unboundedly.

---

## 📌 Question 2: Custom Map vs `hmap` Bucket Evacuation Internals

### Question
Explain why ByteDance created custom map implementations in high-performance RPC frameworks like Kitex instead of default `map[string]any`. How does Go's runtime `hmap` handle key-value buckets during resize evacuation?

### Detailed Solution
1. **Default Map Resize**:
   - **Go 1.24+**: the runtime map switched to a **Swiss-table** design (`internal/runtime/maps`): 8-slot groups with a per-group 64-bit control word, SIMD-friendly probing, and incremental directory growth instead of classic bucket doubling.
   - **Go ≤ 1.23 (legacy `hmap`, still commonly asked)**: `hmap` maintains an array of buckets (`bmap`); when the load factor exceeds 6.5, Go doubles the bucket count ($2 \times B$), and re-allocation is incremental (evacuating 1–2 buckets per write to avoid latency spikes). Interviews framed around "bucket evacuation" refer to this legacy design — name both and note the version.
2. **Interface Boxing Cost**:
   - `map[string]any` stores values in empty interface `any` (`eface`), which forces heap allocations for scalar types.
3. **ByteDance Optimization**:
   - Uses typed generic maps or static memory buffers to avoid `eface` boxing allocations entirely.
