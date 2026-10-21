# Topic 10 Easy Questions

### Q: How do you declare a generic function in Go?
**Frequency**: often
**Why interviewers ask**: It is the entry-level syntax check for generics — if you cannot write this, nothing else about generics matters.
**Answer**: Put type parameters in square brackets between the function name and its parameter list, each with a constraint:
```go
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
```
Inside the body, `T` behaves like a single concrete type that is only known to satisfy its constraint, so you may only use operations the constraint guarantees (here `>`).
**Follow-ups**: How would you declare a generic *type* (e.g. a stack) instead of a function? What happens if you use an operator the constraint does not guarantee?

### Q: What built-in constraint is used for map keys in generics?
**Frequency**: often
**Why interviewers ask**: Map keys are the most common real-world use of generics, and this tests whether you know the predeclared constraint that models map-key eligibility.
**Answer**: `comparable`. It is a predeclared interface (since Go 1.18) satisfied exactly by types that support `==`/`!=`: numbers, strings, booleans, pointers, channels, and structs/arrays built from comparable types. A generic `Set[T comparable]` or `map[T]V` must constrain `T` with `comparable`. Since Go 1.20, interface types also satisfy `comparable`, but comparing them panics at runtime if the dynamic values are not comparable.
**Follow-ups**: Why can't a struct containing a slice field be used with `comparable`? Is `comparable` a normal interface you can use as a variable type?

### Q: What is a type constraint, and what do `any` and the union operator `|` mean in one?
**Frequency**: often
**Why interviewers ask**: Constraints are the core mental model of Go generics — they distinguish "generic over everything" from "generic over operations".
**Answer**: A constraint is an interface used to limit which types may fill a type parameter; it defines the operations available on values of that type parameter. `any` is an alias for `interface{}` and imposes no requirements, so you can only assign, pass, and compare-for-identity such values. A union lists permitted types, e.g.:
```go
type Number interface {
	~int | ~int64 | ~float64
}
```
A union element gives you the operations common to *all* listed types (e.g. `+` and `<` for the numeric union above).
**Follow-ups**: Can a constraint interface with type elements (`|`) be used as an ordinary variable type? How do you call a method on a type parameter's values?

### Q: What is the difference between a generic function and a function taking `interface{}` arguments?
**Frequency**: often
**Why interviewers ask**: It checks whether the candidate understands the actual problem generics solve — static type safety without reflection or boxing.
**Answer**: With `interface{}`, the compiler forgets the concrete type at the call site: you lose static type checking, need type assertions (which can panic), values may be boxed onto the heap, and there is no relationship between parameter and return types. With generics, the compiler type-checks each instantiation: `func Identity[T any](v T) T` guarantees the return type equals the argument type, needs no assertions, and lets the compiler keep values unboxed. Rule of thumb: use generics for type-parametric data structures and algorithms; use `interface{}` (or better, a behavior interface) when types genuinely differ and you only need shared behavior.
**Follow-ups**: When would you still prefer an interface parameter over a type parameter? What does the compiler generate for each instantiation of a generic function?
