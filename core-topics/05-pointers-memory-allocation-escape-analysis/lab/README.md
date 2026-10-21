# 🧪 Lab: Escape Analysis Scavenger Hunt

`escapehunt/` contains one program with 10 short functions, each a distinct
stack-vs-heap scenario. Your job: **predict first, verify second.**

## How to run

From the repository root:

```bash
go build -gcflags='-m' ./core-topics/05-pointers-memory-allocation-escape-analysis/lab/escapehunt/
```

(Or `cd` into this directory and run `go build -gcflags='-m' ./lab/escapehunt/`.)

Add a second `-m` for the escape *flow* (why the compiler decided):

```bash
go build -gcflags='-m -m' ./lab/escapehunt/
```

The program also runs normally — `go run ./lab/escapehunt/` — but the output
values are not the point; the compiler diagnostics are.

## Worksheet

For each function in `escapehunt/main.go`, write down your prediction
(stack or heap) for the marked variable **before** scrolling to the answers.

| #  | Function         | Variable        | Your prediction |
|----|------------------|-----------------|-----------------|
| 1  | `sum`            | `total`         |                 |
| 2  | `newCounter`     | `count`         |                 |
| 3  | `report`         | `value`         |                 |
| 4  | `fixedWindow`    | `make([]int, 8)`|                 |
| 5  | `dynamicWindow`  | `make([]int, n)`|                 |
| 6  | `localWindow`    | `make([]int, n)`|                 |
| 7  | `bigBuffer`      | `buf` (1 MiB)   |                 |
| 8  | `counter`        | `n`             |                 |
| 9  | `leak`           | `x`             |                 |
| 10 | `send`           | `v`             |                 |

## Answer key

<details>
<summary>Reveal answers (verified with go1.27.2, linux/amd64)</summary>

| #  | Verdict | Key `-m` output                                   | Why |
|----|---------|---------------------------------------------------|-----|
| 1  | Stack   | `can inline sum` (no escape lines)                | Returned by copy; the value never outlives the frame. |
| 2  | Heap    | `moved to heap: count`                            | A pointer to `count` is returned, so it outlives the frame. |
| 3  | Heap    | `42 escapes to heap`                              | Boxing into the `interface{}` (`any`) parameter of `fmt.Println`. |
| 4  | Stack   | `make([]int, 8) does not escape`                  | Constant, small size; backing array fits the frame. |
| 5  | Heap    | `make([]int, n) escapes to heap`                  | The slice is returned, so its backing array outlives the frame. |
| 6  | Stack*  | `make([]int, n) does not escape`                  | Not returned and never shared — see the note below. |
| 7  | Heap    | `moved to heap: buf`                              | A 1 MiB local exceeds the stack-frame size bounds. |
| 8  | Heap    | `moved to heap: n`, `func literal escapes to heap`| Captured by reference; the closure is returned and outlives the frame. |
| 9  | Heap    | `moved to heap: x`                                | Stored in a package-level variable reachable after return. |
| 10 | Heap    | `moved to heap: v`                                | Sent through a channel; channel operations are opaque to escape analysis, even for a purely local channel. |

\* **Scenario 6 nuance:** the classic rule "unknown size at compile time → heap"
no longer holds by itself in current Go. When a non-constant-size slice does
not escape, the compiler emits a *runtime* size check: small sizes are
allocated on the stack, and only large sizes fall back to
`runtime.makeslice` on the heap. Inspect it yourself with
`go build -gcflags='-S'` and look for the `CALL runtime.makeslice` branch.
The rule *does* still hold whenever the slice also outlives the frame —
which is exactly scenario 5.

</details>

## Reading tips

- Ignore the lines pointing into `main` (high line numbers): passing every
  result to `fmt.Println` boxes them, so `~r0 escapes to heap` there is the
  interface-boxing rule again, not a property of the scenario.
- Look for `can inline` lines: when a call is inlined, the callee's locals are
  analyzed in the caller's frame, which can change escape outcomes. Add a
  `//go:noinline` directive to any scenario and re-run to compare.
- Re-run with `-gcflags='-m -m'` on a scenario you got wrong — the
  `flow:` trace shows the exact assignment chain that forced the escape.
