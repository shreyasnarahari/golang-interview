# Topic 10 Real Interview Q&A

### Q: What is the difference between `int` and `~int` in a type constraint?
**Answer**: `int` strictly matches `int`. `~int` matches `int` as well as any custom defined type whose underlying type is `int` (e.g. `type UserID int`).

### Q: Can a struct method introduce a new type parameter?
**Answer**: No. Method declarations cannot declare type parameters that were not declared on the receiving receiver type.
