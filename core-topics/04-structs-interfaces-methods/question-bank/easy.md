# Topic 04 Easy Questions

### Q: What is the empty interface `any`?
**Frequency**: often
**Why interviewers ask**: It is the foundation of Go's pre-generics polymorphism and still appears constantly in real code.
**Answer**: `any` (a Go 1.18 alias for `interface{}`) declares an interface with zero methods. Since every type has at least zero methods, every type satisfies it, so it can hold a value of any type. Use it for truly heterogeneous data (e.g. `json.Unmarshal` targets, `slog` key-value pairs); prefer generics when you need to preserve static type information.
**Follow-ups**: How do you get the concrete type back out of an `any` value? When would you choose a generic type parameter over `any`?

### Q: How does a type implement an interface in Go?
**Frequency**: often
**Why interviewers ask**: Implicit interface satisfaction is the biggest difference from Java/C++ and shapes Go's decoupled package design.
**Answer**: Implicitly — there is no `implements` keyword. A type satisfies an interface simply by defining all of its methods. This means interfaces can be declared in the consumer's package (the "accept interfaces" idiom) and retroactively satisfied by types in packages that never heard of them. The check happens at compile time wherever the concrete type is assigned to the interface.
**Follow-ups**: How would you assert at compile time that a type implements an interface? Can a type satisfy an interface from a package it doesn't import?

### Q: What is the difference between a value receiver and a pointer receiver method?
**Frequency**: often
**Why interviewers ask**: Receiver choice is the first decision in every method declaration and reveals whether the candidate understands Go's copy semantics.
**Answer**: A value receiver `func (t T) M()` operates on a copy of the value; mutations inside the method are lost. A pointer receiver `func (t *T) M()` operates on the original value and can mutate it. Rules of thumb: use pointer receivers when the method mutates state, when the struct is large (avoid copying), or when the struct contains a no-copy field like `sync.Mutex`; value receivers are fine for small immutable types like `time.Time`. Be consistent: if any method needs a pointer receiver, usually all should use one.
**Follow-ups**: What happens if you mix receiver types and then assign the value to an interface? Why does copying a `sync.Mutex` cause bugs?

### Q: What does embedding a struct field do, and how are its fields and methods accessed?
**Frequency**: sometimes
**Why interviewers ask**: Embedding is Go's composition mechanism and candidates must know it is not inheritance.
**Answer**: Declaring a field without a name (`type Server struct { log.Logger }` style: `type A struct { B }`) embeds the type; its fields and methods are *promoted*, so `a.X` works as shorthand for `a.B.X`. Promotion is pure syntactic sugar over delegation — there is no subtyping, no virtual dispatch, and the outer type does not "become" the inner type (though promoted methods do count toward the outer type's interface satisfaction).
**Follow-ups**: What happens when the outer and embedded struct both define the same method? Does an embedded value's method promote to the outer value if it has a pointer receiver?
