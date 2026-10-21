# Topic 04 Medium Questions

### Q: Can a struct value satisfy an interface if its method has a pointer receiver?
**Frequency**: often
**Why interviewers ask**: Method sets are the most common source of "does not implement" compile errors for developers new to Go.
**Answer**: No. The method set of `T` contains only value-receiver methods; the method set of `*T` contains both pointer- and value-receiver methods. So `var _ I = T{}` fails to compile when `I`'s method is declared on `*T`, while `var _ I = &T{}` works. The compiler lets you *call* a pointer method on an addressable value `t.M()` by implicitly rewriting it to `(&t).M()`, but interface assignment requires the method to be in the static method set — no implicit address-taking happens there.
**Follow-ups**: Why can you call a pointer method on an addressable value but not store that value in an interface? What is the method set of an interface type itself?

### Q: Why does `m["k"].Inc()` fail to compile when `Inc` has a pointer receiver and `m` is a `map[string]Counter`?
**Frequency**: sometimes
**Why interviewers ask**: It tests whether the candidate understands addressability, not just method sets.
**Answer**: Map index expressions are not addressable — the runtime may relocate or copy bucket entries, so taking `&m["k"]` is forbidden. Calling a pointer-receiver method requires an addressable receiver, so the compiler rejects it (modern toolchains say `cannot call pointer method Inc on Counter`). Fixes: make the map store pointers (`map[string]*Counter`), or copy the value out, mutate, and write it back:
```go
c := m["k"]
c.Inc()
m["k"] = c
```
**Follow-ups**: Is a struct field reachable through a pointer addressable (e.g. `p.Field.M()`)? Why are slice elements addressable when map values are not?

### Q: What happens when you assign a typed nil pointer to an interface and compare it to `nil`?
**Frequency**: often
**Why interviewers ask**: The nil-interface trap causes real production bugs, especially in error handling.
**Answer**: The comparison is `false`. An interface value is a (type, value) pair; it is only `nil` when *both* are nil. Assigning `var p *T = nil` to an interface `i` stores a non-nil type descriptor with a nil value:
```go
var p *MyErr
var err error = p
fmt.Println(err == nil) // false — calling err.Error() may panic
```
The classic failure is a function returning `error` that does `return p` where `p` is a nil `*MyErr`. Fix: return the literal `nil`, or declare the function's return type as the interface and only assign non-nil concrete values.
**Follow-ups**: How does `errors.Is`/reflection-based `reflect.ValueOf(x).IsNil()` interact with this? How would you defensively check for a typed-nil inside a function that receives an `error`?

### Q: If an outer struct and its embedded struct both define a method `Name()`, which one wins, and what if two embedded structs both define it?
**Frequency**: sometimes
**Why interviewers ask**: Shadowing rules decide whether embedding-based designs compile and behave as intended.
**Answer**: The shallowest method wins: the outer struct's own `Name()` shadows the promoted one (which remains reachable explicitly via `a.B.Name()`). With two embedded structs at the *same depth* both providing `Name()`, the selector `a.Name()` is ambiguous and is a **compile error** — you must disambiguate with `a.B1.Name()` or `a.B2.Name()`. This differs from C++ multiple inheritance's runtime surprises: Go resolves everything statically at selection time.
**Follow-ups**: If the ambiguous method is part of an interface the outer struct must satisfy, does it still satisfy the interface? (No — ambiguity means no promotion, so satisfaction fails.)
