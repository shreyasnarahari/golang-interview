# Topic 01: Golang Fundamentals & Data Types

## 📌 Overview
Go is a statically typed, compiled programming language designed at Google. Understanding Go's basic types, zero values, string immutability, UTF-8 encoding, and pass-by-value semantics is essential for building a solid foundation.

---

## 🎯 Key Concepts

### 1. Variables, Constants & Zero Values
- Go initializes every variable to its **zero value** if an explicit initial value is not provided.
- **Zero Value Table**:
  - `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `float32`, `float64`: `0`
  - `bool`: `false`
  - `string`: `""` (empty string)
  - Pointers (`*T`), Slices (`[]T`), Maps (`map[K]V`), Channels (`chan T`), Interfaces (`any`): `nil`
  - Structs: A struct with all fields initialized to their respective zero values.

```go
var a int           // 0
var b string        // ""
var c []int         // nil
var d map[string]int// nil
```

### 2. Strings, Runes, and Bytes
- In Go, a `string` is an immutable sequence of arbitrary bytes (typically UTF-8 encoded text).
- `byte` is an alias for `uint8`.
- `rune` is an alias for `int32` and represents a Unicode code point.
- **Key Difference**:
  - `len(str)` returns the **number of bytes**, NOT characters/runes.
  - To count Unicode code points, use `utf8.RuneCountInString(str)`.

```go
s := "Go 🐹"
fmt.Println(len(s))                    // 7 bytes (🐹 takes 4 bytes in UTF-8)
fmt.Println(utf8.RuneCountInString(s)) // 4 runes ('G', 'o', ' ', '🐹')
```

### 3. Strict Pass-By-Value Semantics
- **CRITICAL GO PRINCIPLE**: Everything in Go is strictly **Pass-By-Value**.
- When you pass a variable to a function, Go creates a copy of the value.
- When passing a pointer `*T`, Go passes a copy of the pointer (memory address value).
- When passing a slice or map, Go passes a copy of the slice header or map pointer descriptor.

---

## ⚠️ Common Gotchas
1. **`len(string)` vs `utf8.RuneCountInString`**: Indexing a string `s[i]` yields a `byte` (`uint8`), not a character or `rune`.
2. **String Immutability**: Strings cannot be modified in place (`s[0] = 'a'` fails to compile). Convert to `[]rune` or `[]byte` to modify, then convert back.
