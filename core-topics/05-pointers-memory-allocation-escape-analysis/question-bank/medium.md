# Topic 05 Medium Questions

### Q: Why does passing a local struct to `fmt.Println(x)` cause it to escape to the heap?
**Answer**: Because `fmt.Println(a ...any)` accepts variadic `any` (`interface{}`). Converting concrete types to interface values requires runtime dynamic type descriptors (`eface`), which forces heap allocations.
