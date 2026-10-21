# Topic 01 Real Interview Questions & Answers

### Q: "Walk me through Go's parameter-passing model — is anything ever passed by reference?" (asked at a cloud infrastructure company)
**Frequency**: often
**Why interviewers ask**: Infra companies care about allocation and aliasing behavior in hot paths; this one question exposes your whole mental model of Go values.
**Answer**: Go is strictly pass-by-value — every argument is copied — but several types are *descriptor* types whose copied header contains a pointer to shared data, so they behave reference-like. What the interviewer wants is a precise breakdown: (1) strings copy a 16-byte {ptr, len} header, never the bytes; (2) slices copy a 24-byte {ptr, len, cap} header — the callee can mutate shared elements but can't change the caller's length; (3) maps, channels, and func values copy a pointer to a runtime structure (`hmap`, `hchan`), so writes through them are visible to the caller; (4) arrays and structs are copied in full — a 1 MB array argument really copies 1 MB. Practical implications they'd chain into: pass structs by pointer when large or when mutation must be visible; never copy a struct containing a `sync.Mutex` after first use (`go vet` copylocks catches this); and returning a pointer to a local is fine because escape analysis moves it to the heap automatically.
**Follow-ups**: If everything is copied, why is it still idiomatic to return `[]T` from a function instead of taking an output parameter? When would passing a small struct by pointer actually be *slower*? (Pointer forces heap escape/GC pressure; a value copy can stay on the stack or in registers.)

### Q: "Our payment API receives names in any language. Where do Go's string semantics bite us?" (asked at a fintech handling international users)
**Frequency**: sometimes
**Why interviewers ask**: Fintechs deal with real-world Unicode (names, addresses, CJK, emoji) and want to know you won't ship byte-level bugs into user-facing text handling.
**Answer**: A Go string is an immutable sequence of UTF-8 **bytes**, so anything byte-based mis-handles multi-byte text: `len(name)` counts bytes, not characters (`len("你好")` is 6, not 2); `name[i]` yields a `byte`, not a character; `name[:10]` as a "truncate for display" can slice through the middle of a rune, producing invalid UTF-8 that renders as U+FFFD — and in a payments context can break downstream systems that validate input encoding. Correct tools: `for range` over the string to decode runes, `utf8.RuneCountInString` for character counts, `[]rune` conversion when you need random access or reversal, and `golang.org/x/text` (cases, collation, normalization) for locale-correct comparisons — two visually identical names can differ byte-wise due to Unicode normalization (NFC vs NFD), so naive `==` on strings can reject a legitimate user. Validation should also reject invalid UTF-8 explicitly (`utf8.ValidString`) before it enters the database.
**Follow-ups**: A teammate proposes storing a `maxLen` check as `len(s) > 280` — what's wrong? (Bytes vs characters; a 280-"character" limit should count runes, and arguably grapheme clusters, not bytes.) How would you compare two names case-insensitively across locales? (`strings.EqualFold` handles simple Unicode case folding; full locale-aware collation needs `x/text/collate`.)

### Q: "Design a type-safe permission/flags system for our CLI tool. Show me how you'd do the constants." (asked at a developer-tools company)
**Frequency**: sometimes
**Why interviewers ask**: It tests whether you can produce the idiomatic Go enum pattern — named type, `iota`, bit flags, and a `String()` method — in one shot.
**Answer**: Define a named type plus an `iota` block so zero is meaningful and values are exhaustive:

```go
type Perm uint8

const (
	PermRead Perm = 1 << iota // 1
	PermWrite                  // 2
	PermExec                   // 4
)

// Combine and test flags with bitwise ops.
var p = PermRead | PermWrite
if p&PermRead != 0 { /* ... */ }
```

Points that earn depth credit: use a *named type* (`Perm`, not bare `int`) so function signatures reject raw integers — Go has no implicit conversion between named types; make the zero value deliberate (here "no permissions" — often you burn `iota` on an `Unknown`/`None` member with `_` if zero should be invalid); implement `String()` (or run `stringer -type=Perm` from `golang.org/x/tools`) so logs print names instead of numbers; and pick `uint8`/`uint32` deliberately since bit flags have a hard width limit. Mention that Go deliberately lacks sum types, so the "enum" is a convention, not a compiler-checked exhaustiveness guarantee — for closed sets you add a `switch` with a `default` panic/error.
**Follow-ups**: A teammate adds a flag by writing `PermFoo = 5` mid-block — what breaks? (It silently collides with `PermRead|PermExec`; bit-flag values must stay powers of two — this is why the `1 << iota` pattern matters.) How would you serialize `Perm` to JSON as names rather than numbers? (Custom `MarshalJSON`/`UnmarshalJSON`, or a lookup table — `encoding/json` has no enum support.)
