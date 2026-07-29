# Topic 10 Medium Questions

### Q: Why do slice and map elements not satisfy `comparable` if they contain non-comparable fields?
**Answer**: In Go, slices, maps, and functions cannot be compared with `==`. Any struct or type containing a slice/map field is non-comparable and cannot be used as a map key or `comparable` generic parameter.
