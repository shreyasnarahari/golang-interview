# Topic 11 Labs: Scheduler Experiments

Hands-on experiments for the GMP scheduler. Build before running so the Go
toolchain's own scheduler traces don't mix into `GODEBUG` output.

---

## Lab 1: schedtrace — Watch the Run Queues Fill Up

`schedtrace/` spawns 4×`NumCPU` busy goroutines — more runnable goroutines
than there are **P**s — so excess goroutines pile up in local run queues and
the Global Run Queue.

**Run:**

```bash
go build -o /tmp/schedtrace ./lab/schedtrace
GODEBUG=schedtrace=1000 /tmp/schedtrace
```

(`schedtrace=1000` prints a scheduler snapshot every 1000 ms.)

**What to observe:**

```
SCHED 1002ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 needspinning=1 idlethreads=0 runqueue=17 [ 2 0 0 1 0 2 1 1 ] schedticks=[ 45 47 44 46 47 42 50 46 ]
```

| Column | Meaning |
|---|---|
| `SCHED 1002ms` | Time since program start. |
| `gomaxprocs=8` | Number of **P**s (logical processors). |
| `idleprocs=0` | **P**s with no work — 0 here because every **P** is saturated. |
| `threads=9` | Total **M**s (OS threads) the runtime has created. |
| `spinningthreads` | **M**s actively spinning, hunting for work to steal. |
| `idlethreads` | Parked **M**s with no **P** bound. |
| `runqueue=17` | Goroutines waiting in the **Global Run Queue**. |
| `[ 2 0 0 1 ... ]` | Per-**P** local `runq` lengths. |
| `schedticks=[ ... ]` | Per-**P** scheduler tick counts; every 61st tick forces a GRQ check. |

**Takeaway:** With more runnable goroutines than **P**s, `runqueue` and the
per-**P** queues stay non-zero and `idleprocs` drops to 0. This is the queue
backlog you saw in the architecture diagram, observed live.

---

## Lab 2: gomaxprocs — Scaling CPU-Bound Work

`gomaxprocs/` runs the same CPU-bound workload across `NumCPU` worker
goroutines twice: once with `runtime.GOMAXPROCS(1)` (one **P**, one core) and
once with `runtime.GOMAXPROCS(NumCPU)` (all cores).

**Run:**

```bash
go run ./lab/gomaxprocs
```

**What to observe:**

```
NumCPU=8, 8 workers x 200000000 iterations each
GOMAXPROCS=1:   1.410761201s
GOMAXPROCS=8: 329.435202ms
speedup: 4.28x on 8 cores
```

**Takeaway:** With one **P** all goroutines timeshare a single OS thread
context; with `NumCPU` **P**s they run truly in parallel. Speedup is below
the core count because of scheduling overhead and shared hardware (SMT
siblings, memory bandwidth). Try `GOMAXPROCS=2` and `GOMAXPROCS=4` by editing
`measure(...)` to see the scaling curve.

---

## Lab 3: preemption — Asynchronous Preemption Since Go 1.14

`preemption/` pins `GOMAXPROCS` to 1 and launches a goroutine running a tight
`for {}` loop with no function calls — the worst case for the old cooperative
scheduler.

**Run:**

```bash
go run ./lab/preemption
```

**What to observe:**

```
main resumed after 56ms despite a tight loop hogging the only P
```

The program terminates normally. In Go < 1.14 this exact program would hang
forever: the loop has no `morestack` preemption point, so with a single **P**
`main` could never resume after `time.Sleep`.

**Takeaway:** Since Go 1.14 the runtime uses `SIGURG`-based asynchronous
preemption to pause goroutines at (almost) any instruction, so even
function-call-free loops cannot starve the scheduler or the GC.
