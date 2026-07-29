# Topic 04 Medium Questions

### Q: Can a struct value satisfy an interface if its method has a pointer receiver?
**Answer**: No. Method sets state that values of type `T` only possess value receiver methods. Only pointers `*T` possess both pointer and value receiver methods.
