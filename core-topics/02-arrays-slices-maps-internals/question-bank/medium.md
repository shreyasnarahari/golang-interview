# Topic 02 Medium Questions

### Q: Why does deleting a key from a map not immediately free memory to the OS?
**Answer**: Deleting keys clears the bucket entries, but the allocated buckets (`hmap.buckets`) remain allocated to prevent re-allocation overhead when future keys are added. To reclaim memory, construct a new map and copy active key-value pairs.
