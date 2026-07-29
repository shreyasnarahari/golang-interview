# Topic 15 Hard Questions

### Q: What performance penalty does compiling with `-race` introduce?
**Answer**: `-race` increases CPU execution time by $\approx 2\times - 10\times$ and memory footprint by $\approx 5\times - 20\times$. Therefore, `-race` builds should never be used in production binaries.
