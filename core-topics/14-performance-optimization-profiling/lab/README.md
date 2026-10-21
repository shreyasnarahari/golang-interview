# Lab 14: Live Profiling with `net/http/pprof`

Profile a deliberately slow HTTP server and identify its hot functions with `go tool pprof`.

## The Program

`pprofserver/` serves three kinds of endpoints on `:8080`:

- `GET /cpu` — burns CPU hashing in a tight loop (`burnCPU`)
- `GET /alloc` — allocates hundreds of thousands of small buffers (`churnMemory`)
- `/debug/pprof/*` — registered automatically by the blank import `_ "net/http/pprof"`

## Step 1 — Start the server

From the module directory (`core-topics/14-performance-optimization-profiling`):

```bash
go run ./lab/pprofserver
```

## Step 2 — Generate CPU load

In a second terminal, keep the CPU endpoint busy:

```bash
while true; do curl -s localhost:8080/cpu > /dev/null; done
```

## Step 3 — Capture a CPU profile

In a third terminal, capture 30 seconds of samples:

```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

Then, inside the interactive pprof shell:

```
(pprof) top
(pprof) list burnCPU
```

`top` shows `main.burnCPU` (and its `crypto/sha256` callees) dominating the flat samples; `list burnCPU` annotates the hashing loop line by line. Add `-http=:8081` to the capture command to browse a flame graph instead.

## Step 4 — Capture a heap profile

Stop the CPU load loop, then hammer the allocation endpoint:

```bash
while true; do curl -s localhost:8080/alloc > /dev/null; done
```

And capture:

```bash
go tool pprof -alloc_space http://localhost:8080/debug/pprof/heap
```

`top` points at `main.churnMemory` as the source of allocation churn (`-alloc_space` shows everything allocated; the default `-inuse_space` shows what is still live).

## Step 5 (optional) — Execution trace

```bash
curl -o trace.out "http://localhost:8080/debug/pprof/trace?seconds=5"
go tool trace trace.out
```

Stop the server and the load loops when done (`Ctrl-C`).
