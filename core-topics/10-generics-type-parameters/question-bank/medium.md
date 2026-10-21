# Topic 10 Medium Questions

### Q: Why does this fail to compile, and how do you fix it? `func Sum[T any](s []T) T { var total T; for _, v := range s { total += v }; return total }`
**Frequency**: often
**Why interviewers ask**: It is the classic behavior-prediction question that separates candidates who have written generic code from those who have only read about it.
**Answer**: `any` guarantees no operations, so `+=` is rejected: the compiler cannot know `T` supports addition. Fix it with a constraint whose type set only contains addable types, e.g. a union:
```go
type Addable interface {
	~int | ~int64 | ~float64 | ~string
}
```
or use `cmp.Ordered` if you also need comparisons. Every operator used on a type-parameter value must be valid for *every* type in the constraint's type set.
**Follow-ups**: Why does the union need `~int` instead of `int` if callers use `type Celsius int`? Could you use a method-based constraint here instead of a union?

### Q: A struct has a slice field, e.g. `type Row struct { Cells []string }`. Why can't it be used as `T` in `func Index[T comparable](...)` — and what changed in Go 1.20 about interface types satisfying `comparable`?
**Frequency**: sometimes
**Why interviewers ask**: It tests precise knowledge of comparability rules and a subtle spec relaxation that bites people in production with runtime panics.
**Answer**: Slices, maps, and funcs are not comparable with `==` (only to `nil`), and a struct/array is comparable only if all its fields/elements are, so `Row` is not comparable and cannot instantiate `T comparable`. Since Go 1.20 the spec was relaxed: *interface* types (like `any`) now satisfy `comparable`, so `Equal[any](a, b)` compiles — but it panics at runtime if both dynamic values are identical and non-comparable (e.g. `[]int`). The compiler accepts it because interfaces are "not strictly comparable": comparability depends on the dynamic type.
**Follow-ups**: How would you make `Index` work for `Row` anyway (hint: a method-based constraint or an `equal func(a, b T) bool` parameter)? What happens when you compare two interface values whose dynamic types differ?

### Q: Can you write a type switch on a value of type parameter type `T`? If so, how?
**Frequency**: sometimes
**Why interviewers ask**: It probes understanding of the boundary between compile-time generic typing and runtime type inspection.
**Answer**: Not directly — a value of type parameter type has no interface methods to switch on unless the constraint provides them. But you can convert to `any` first, which preserves the dynamic type at runtime:
```go
func Describe[T any](v T) string {
	switch x := any(v).(type) {
	case int:
		return fmt.Sprintf("int %d", x)
	case string:
		return x
	default:
		return "unknown"
	}
}
```
This works but defeats most of the point of generics (per-case runtime dispatch); prefer constraining `T` or accepting a behavior interface when you need case analysis.
**Follow-ups**: Why is a type switch inside a generic function usually a design smell? What is the difference between converting `v` to `any` and calling a method declared on the constraint?

### Q: Given `type Celsius int`, why does `Max[Celsius]` fail if `Max` is declared as `func Max[T int | float64](a, b T) T`, and how do you fix it?
**Frequency**: often
**Why interviewers ask**: The `~` underlying-type rule is the most commonly hit surprise when applying generics to real domain types.
**Answer**: A plain type term `int` in a constraint matches *only* `int` itself, not defined types whose underlying type is `int`, so `Celsius` is excluded. Prefixing with `~` matches by underlying type:
```go
func Max[T ~int | ~float64](a, b T) T
```
Now `Celsius` is accepted, and inside the function `T`'s values still support `>` because every type in the type set does. Note the result keeps the caller's type — `Max(Celsius(3), Celsius(5))` returns `Celsius`, not `int`.
**Follow-ups**: Does `~int` also grant access to methods declared on `Celsius` inside the generic function? (No — only the constraint's methods are visible.) When would you deliberately *not* use `~`?
