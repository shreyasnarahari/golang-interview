# Topic 02 Real Interview Q&A

### Q: What is the difference between an Array and a Slice in Go?
**Answer**: An array has a fixed length defined at compile time (`[5]int`), and its size is part of its type. A slice (`[]int`) is a dynamic reference wrapper around an array consisting of a 24-byte `SliceHeader` (Data pointer, Len, Cap).

### Q: How does Go handle map collisions?
**Answer**: Each bucket (`bmap`) holds 8 key-value pairs. When a bucket fills up, Go attaches overflow buckets via pointer. When the load factor exceeds 6.5, Go triggers incremental map expansion.
