# Lab: GC Churn — GOGC / GOMEMLIMIT Trade-off

`gcchurn/` is a small program that allocates ~100 MiB of short-lived garbage per
round (100,000 × 1 KiB buffers, only the last one stays reachable) and prints
`NumGC`, `HeapAlloc`, and `HeapSys` after each of 5 rounds.

Run it under `GODEBUG=gctrace=1` with different pacing settings and compare.

## Runs

```sh
# 1. Default pacing (GOGC=100)
GODEBUG=gctrace=1 go run ./lab/gcchurn

# 2. More aggressive collection
GOGC=50 GODEBUG=gctrace=1 go run ./lab/gcchurn

# 3. GC pacing disabled, no memory limit — heap grows unbounded
GOGC=off go run ./lab/gcchurn

# 4. GC pacing disabled, soft memory cap set — GC only runs near the limit
GOGC=off GOMEMLIMIT=64MiB GODEBUG=gctrace=1 go run ./lab/gcchurn
```

Typical results on an 8-core machine (your numbers will vary):

| Setting | NumGC after 5 rounds | HeapSys | GC CPU |
|---|---|---|---|
| `GOGC=100` (default) | ~158 | ~11 MiB | ~13% |
| `GOGC=50` | ~304 | ~7 MiB | ~13% |
| `GOGC=off`, no limit | 0 | grows ~97 MiB/round | 0% |
| `GOGC=off` + `GOMEMLIMIT=64MiB` | ~9 | plateaus ~59 MiB | ~1% |

## Reading a `gctrace` line

```
gc 158 @0.160s 13%: 0.045+0.35+0.058 ms clock, 0.36+0.011/0.093/0.006+0.46 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 8 P
```

| Field | Meaning |
|---|---|
| `gc 158` | GC cycle number since program start |
| `@0.160s` | Time since program start when the cycle ran |
| `13%` | Share of total CPU consumed by GC so far |
| `a+b+c ms clock` | Wall-clock time: `a` = STW sweep termination, `b` = concurrent mark & scan, `c` = STW mark termination |
| `a+b/c/d+e ms cpu` | Same phases in CPU time; the concurrent phase is split into mutator-assist / background GC / idle GC time |
| `3->4->1 MB` | Heap size at GC start → heap size at GC end → live heap after GC |
| `4 MB goal` | Target heap size the GC pacer is aiming for this cycle |
| `0 MB stacks, 0 MB globals` | Memory scanned in goroutine stacks and global variables |
| `8 P` | Number of Ps (logical processors) in use |

Only the two STW phases stop the world — here they are tens of microseconds,
which is how Go keeps pauses below 1 ms while marking runs concurrently.

## What to conclude

1. **`GOGC` trades memory for CPU.** Lowering it (100 → 50) shrinks the heap
   goal and peak `HeapSys`, but roughly doubles `NumGC` and GC CPU overhead.
   Raising it does the opposite.
2. **`GOGC=off` disables pacing entirely.** With no memory limit the heap grows
   without bound (~97 MiB per round of garbage here) — a real service would be
   OOM-killed. Never ship `GOGC=off` alone.
3. **`GOMEMLIMIT` is a safety net, not a pacing knob.** With `GOGC=off` +
   `GOMEMLIMIT=64MiB`, GC stayed silent until the heap neared the cap, then ran
   just enough cycles to hold `HeapSys` under ~59 MiB — fewest cycles and lowest
   GC CPU of all runs. The limit is *soft*: if the **live** heap approaches it,
   the runtime GCs continuously (thrashing) to stay under it, so the limit must
   comfortably exceed the live set.

## Programmatic equivalents

The same knobs are available in-process via `runtime/debug`:

- `debug.SetGCPercent(n)` — sets `GOGC`; `debug.SetGCPercent(-1)` disables GC
  like `GOGC=off`. Returns the previous value.
- `debug.SetMemoryLimit(bytes)` — sets `GOMEMLIMIT`. Returns the previous value.
