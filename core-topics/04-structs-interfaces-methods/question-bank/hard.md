# Topic 04 Hard Questions

### Q: Explain the runtime structure of `itab` in `iface`.
**Answer**: An `itab` stores:
1. `inter *interfacetype`: Description of the interface type.
2. `_type *_type`: Description of the concrete underlying type.
3. `hash uint32`: Copy of `_type.hash` for fast type assertion comparisons.
4. `fun [1]uintptr`: Array of pointers to the concrete methods implementing the interface methods.
