# Critical Review & Rebuild Plan: golang-interview Repository

> **Implementation status** (updated as phases land):
> - ✅ **P0** — Critical fixes: all factual errors E1–E11 corrected, truncated files completed, redundant index removed, `go.mod` bumped to 1.24.
> - ✅ **P1** — Exercise-first conversion: all 15 modules split into `exercises/` (stubs, red by design) + `solutions/` (green); shallow assertions fixed; Makefile, CI workflow, `.golangci.yml`, and `tools/mdcheck` (markdown snippet compile checker) added.
> - ✅ **P2** — Modernization & depth: `high-performance-go.md` dissolved into modules 05/11/12/14; fuzzing + `slog` + `go.work` added to module 15; PGO and full `pprof`/trace workflow added to module 14; Swiss-table maps, modern context APIs, `slices`/`maps`/iterators documented; hands-on `lab/` experiments added to modules 05/11/12/14.
> - ✅ **P3** — Expert tier: four capstone projects under `capstones/` (rate limiter, KV store, job queue, mini-redis — spec + stubs + acceptance tests + race-clean reference solutions) and a `real-interview-q/system-design/` section with three Go-centric design walkthroughs.
> - ✅ **P4** — Question banks: all 15 modules expanded to 15 calibrated questions each (~224 total) in a standardized Anki-parseable format with frequency tags (`make anki` exports `anki.tsv`); company files deduplicated (cross-referenced worker-pool/atomic-counter duplicates), de-hyped (company-flavored framing), and intros corrected to match actual content.

**All four phases complete.** The rebuild roadmap is fully executed.

---

## Verdict in one paragraph

This repository is a **well-organized outline of Go interview knowledge, not a system for building Go mastery**. Its structure (15 numbered topics, 3 phases, company-tagged questions) looks like a curriculum, but the substance is a flashcard deck: every module totals ~100–150 lines across *all* its files, each question-bank file holds exactly **one question**, each `resources.md` holds exactly **two links**, READMEs average ~300 words, and several READMEs are **truncated mid-sentence** (modules 06, 07, 09, 10 end after "Gotcha #1"). The "practice" is pre-solved — tests merely re-assert shipped solutions, so there is nothing for the learner to actually do. The technical accuracy is ~90% good, but there are ~10 concrete factual errors, and the content is stuck at roughly Go 1.19–1.22 while the toolchain era is 1.24–1.27 (Swiss-table maps, PGO, fuzzing, `context.AfterFunc`, `slog`, iterators — all absent). The "expert" tier (modules 11–15) delivers awareness, not expertise — ironically, the supplemental `high-performance-go.md` (186 lines) contains more scheduler/GC depth than the dedicated expert modules it supplements.

---

## 1. Does it provide a clear progression from foundational to expert?

**On paper, yes. In substance, no.**

- The 15-topic ordering is genuinely logical (types → composite internals → control flow → type system → memory → concurrency → context → errors → generics → scheduler → GC → unsafe → profiling → tooling), and the 3-phase weekly plan in the README is sensible scaffolding.
- But **difficulty is flat**: every module — from "01 fundamentals" to "14 profiling" — operates at the same cognitive level: *recall a fact, read a one-line answer*. There is no ramp from recall → application → analysis → creation.
- **There is no applied tier at all.** Mastery of Go comes from building systems (servers, schedulers, stores) and diagnosing real performance problems. The repo contains zero projects, zero system-design material, zero debugging exercises.
- **The expert tier is mislabeled.** Module 11 (GMP scheduler) is 94 lines total; work stealing, handoff, sysmon, and the netpoller get one line each. You cannot practice scheduling, GC, or profiling anywhere in the repo — there is exactly **one benchmark in the entire repository** (`p14_profiling_test.go`) and **zero** `GODEBUG`/`pprof`/`trace` experiments, zero fuzz tests, zero race-hunting exercises.

---

## 2. Detailed findings (evidence)

### 2.1 Factual errors and staleness (must-fix)

| # | Location | Problem |
|---|---|---|
| E1 | `core-topics/02/README.md:10`, `02/question-bank/hard.md` | Teaches `reflect.SliceHeader`/`StringHeader` without noting **deprecation since Go 1.20** (`unsafe.Slice`/`SliceData`/`String` are the modern APIs) |
| E2 | `core-topics/02/README.md` | `hmap`/`bmap` bucket internals describe the **old map runtime**; Go 1.24 replaced maps with Swiss Tables — unmentioned anywhere in the repo |
| E3 | `core-topics/03/README.md:28`, `03/question-bank/hard.md` | Claims open-coded defers apply to "loops with fixed iterations" — **wrong**: open-coded defers (1.14+) require defers *not* inside loops (≤8 per function) |
| E4 | `core-topics/07/question-bank/hard.md` | `sync.Map` described with `read atomic.Value` — since Go 1.19 it's `atomic.Pointer[readOnly]`; omits the `misses` counter / dirty→read promotion, the actual point of the design |
| E5 | `core-topics/08/README.md` | Omits all Go 1.20/1.21 context APIs: `WithCancelCause`/`Cause`, `WithDeadlineCause`, `WithoutCancel`, `AfterFunc` — now standard interview material |
| E6 | `core-topics/10/README.md:35-38` | Claims Go does "monomorphization for value types" — **wrong**: Go uses gcshape stenciling for all instantiations; `int` and `float64` share a gcshape and one code body. The module's own `hard.md` is more accurate than its README |
| E7 | `core-topics/13/question-bank/real-interview-q-a.md:7` | Says cgo switches "Go **segmented stack** to C OS stack" and involves "disabling GC preemption" — segmented stacks were removed in **Go 1.3**; contradicts the module's own README |
| E8 | `core-topics/15/question-bank/hard.md:4` | Race-detector overhead numbers **swapped**: actual is memory 5–10×, execution time 2–20× |
| E9 | `practice/p10_generics.go:3` | `Stack[T]` commented "thread-safe" with zero synchronization — a false claim in a concurrency-focused repo |
| E10 | `high-performance-go.md:6` | Header advertises `pprof` as a key focus; **no pprof content exists in the file** |
| E11 | `real-interview-q/uber.md:142` | Mutex contention framed as "kernel thread parking / 1–2µs context switch" — Go mutexes use runtime-level `gopark`, not kernel futex, on the contention path |

### 2.2 Missing mandatory modern-Go topics

Fuzzing (1.18, **zero mentions repo-wide**), `go.work` workspaces (1.18), `errors.Join` multi-wrap practice (1.20), PGO (1.21 default), `slog` (1.21), `slices`/`maps` stdlib packages + `min`/`max` builtins (1.21), range-over-func iterators (1.23), `unique`/weak pointers (1.23/1.24), Swiss-table maps (1.24), container-aware `GOMAXPROCS` (1.25), Green Tea GC (1.25 experimental). Root `go.mod` pins `go 1.22` while the installed toolchain is 1.27. For a repo whose brand is "expert in **modern** Golang," this is the largest strategic gap after pedagogical thinness.

### 2.3 The practice isn't practice

- Every exercise ships **complete solutions**; tests re-verify them. Nothing is left for the learner to implement.
- Several tests **don't test the thing being taught**: module 05 teaches stack-vs-heap escape, but its test only checks field values — it would pass even if both functions heap-allocated (should use `testing.AllocsPerRun`). Modules 07 and 10 assert only `len(results)`.
- Named-but-never-built patterns: Fan-Out/Fan-In, Pipeline, Generator, `sync.Once`, `sync.Cond`, `sync.Pool` are listed in READMEs but never exercised.
- `p06_channels.go`'s `NonBlockingReceive` uses `time.After` — it's a *bounded-blocking* receive, not the `select`/`default` non-blocking idiom its own question bank teaches; `p08_context.go` leaks a timer via `time.After` in `select`.
- `p07_sync_patterns.go` fails `gofmt`.

### 2.4 Organization and redundancy

- **Depth inversion**: `high-performance-go.md` §1–2 contain scheduler/GC detail (61-tick GRQ check, runnext, mcache/mcentral/mheap) that modules 11/12 *lack*, while also being near-verbatim redundant with them. Two sources of truth for the same material.
- `real-interview-q/interview-questions.md` (21 lines) is a pure index duplicating `real-interview-q/README.md`; the root README still tells readers it contains questions. One index entry (amazon "AWS SDK context propagation") points at content that doesn't exist.
- Company-file intros **over-promise**: amazon.md advertises errgroup/S3 uploaders, twitch.md advertises sliding-window limiters, meta.md promises cgo content — none exist. `meta.md` ends abruptly mid-topic.
- Near-duplicate questions across company files: bytedance.md Q1 ≈ uber.md Q3 (worker pools), uber.md Q2 ≈ datadog.md Q1 (atomic counters). Uneven depth: Google gets 4 questions/292 lines; Meta, Amazon, Microsoft get 1 each.
- `practice-go/` is an orphan nested module (10-line banner `main.go`) invisible to the root `go test ./...`.

### 2.5 Infrastructure

No CI (`.github/` absent), no lint config, no Makefile. The ~25 Go code blocks embedded in markdown are **never compiled or tested** — errors like E11 above go unchecked. Good news: the root module is clean — `go test ./...` from root genuinely covers all 15 practice suites (verified, all green), and `.gitignore` is appropriate.

### 2.6 What's genuinely good (keep)

- Uniform module layout (README / practice / question-bank / resources) is a solid skeleton.
- Module 04 (structs/interfaces) shows the target standard: receiver matrix, `iface`/`eface` internals, and README+code+test reinforcing the nil-interface trap coherently.
- `real-interview-q/` company files have **complete, compilable, well-explained answers** — the strongest content in the repo.
- `high-performance-go.md`'s unique content (allocator internals, escape-analysis rules, atomic-vs-mutex table) is accurate and worth keeping.
- All tests pass and are race-clean; links are canonical (go.dev, Russ Cox, Ardan Labs).

---

## 3. Recommendations (what / why / how / outcome)

### R1. Convert all practice into exercise-first, test-driven stubs
1. **What**: Replace pre-solved practice files with skeleton files (`// TODO: implement`) whose tests fail until implemented; move current solutions to a `solutions/` subpackage. Fix assertion shallowness (test values, not lengths; use `testing.AllocsPerRun` in module 05; add edge cases: nil map, empty string, nil receiver).
2. **Why**: Retrieval and implementation under test pressure is how interview skill is actually built; reading solutions produces recognition, not recall. Interviews are write-code-under-time-pressure events.
3. **How**: Per module: `exercises/ex.go` (stubs + doc comments specifying contracts), `exercises/ex_test.go` (table-driven, currently failing), `solutions/` (current code, moved). Learner workflow: `go test` → red → implement → green. Add a root `Makefile` target `make exercise T=06` that runs only that module's tests.
4. **Outcome**: Learner can implement every core idiom from scratch, verified by tests — e.g., write a worker pool, a bounded channel fan-in, and an error-wrapping validator with zero reference material.

### R2. Add hands-on runtime labs for the "unpracticable" topics
1. **What**: New `lab/` component in modules 05, 11, 12, 14 with guided experiments, each with explicit "what to observe" checkpoints.
2. **Why**: Scheduler, GC, and profiling expertise is observational — you must *see* `schedtrace`, `gctrace`, escape-analysis output, and pprof flame graphs to internalize the models. Currently these topics are pure trivia.
3. **How**:
   - `05/lab`: `-gcflags='-m -m'` scavenger hunt — 10 snippets, predict escape before checking.
   - `11/lab`: CPU-bound loop with/without async preemption; `GODEBUG=schedtrace=1000` reading guide; `GOMAXPROCS` scaling experiment.
   - `12/lab`: allocation-churn program; observe `NumGC`/heap under `GOGC=off/50/100` and `GOMEMLIMIT` via `GODEBUG=gctrace=1`; `debug.SetGCPercent` exercise.
   - `14/lab`: intentionally slow HTTP server with `net/http/pprof`; capture CPU + heap profiles; `go tool pprof -top/-peek` worksheet; `runtime/trace` capture; `-bench` + `benchstat` before/after an optimization; one deliberate data race to find with `-race`.
4. **Outcome**: Learner can diagnose a real latency/memory regression: capture profiles, read a flame graph, identify allocation hotspots, tune GOGC/GOMEMLIMIT, and verify with benchmarks.

### R3. Fix the 11 factual errors and modernize to the current Go era
1. **What**: Correct E1–E11 (table in §2.1); add the missing modern topics from §2.2; bump `go.mod` to the current stable Go.
2. **Why**: An interview-prep repo with wrong answers is worse than no repo — learners will repeat the open-coded-defer and gcshape errors in interviews. Swiss-table maps and PGO are now *the* expected answers for map-internals and optimization questions.
3. **How**: Line-edit the listed files; restructure module 02 to teach old `hmap` *and* Swiss tables as "know the evolution" (interviewers still ask classic buckets); add `context.AfterFunc`/`WithoutCancel`/`WithCancelCause` to module 08; add a fuzzing section + `FuzzXxx` exercise to module 15; add PGO + Swiss maps + `slog`/`slices`/`maps`/iterators coverage; CI pins a recent toolchain.
4. **Outcome**: Every claim in the repo is correct against a current Go release; learner's answers match what 2026-era interviewers expect.

### R4. Re-tier the curriculum into an explicit 4-level progression with capstones
1. **What**: Reorganize 15 flat modules into 4 levels with explicit gate criteria, and add a new **Level 4: Applied** of capstone projects. Keep the existing numbering as paths within levels to preserve deep links.
2. **Why**: The current flat list hides the skill ramp and ends at "testing tooling" — there is nowhere to *integrate* knowledge. Mastery requires synthesis: building systems that force concurrency + errors + context + performance together.
3. **How**:
   - L1 Foundations (modules 01–05): gate = exercises green without reference.
   - L2 Concurrency & Idioms (06–10): gate = implement worker pool, pipeline, fan-in from scratch; pass a hidden race-detector suite.
   - L3 Runtime Internals (11–15 + merged high-performance content): gate = complete all labs and explain observations in writing.
   - L4 Applied (new `capstones/`): build (a) a rate limiter library (token bucket + sliding window, `sync` + atomics), (b) an in-memory KV store with TTL + snapshots (maps internals, GC pressure, benchmarks), (c) a worker-pool/job-queue library with graceful shutdown (context, errgroup-style), (d) a mini Redis-like TCP server (net, RESP parsing, goroutine-per-conn, pprof-instrumented). Each capstone: spec README, stubbed package, hidden acceptance tests, benchmark targets.
4. **Outcome**: Learner has four portfolio-grade artifacts and can discuss design trade-offs (lock granularity, backpressure, allocation strategy) — exactly what senior-level interviews probe.

### R5. Merge `high-performance-go.md` into modules 11/12/14 and eliminate duplicate sources of truth
1. **What**: Dissolve the standalone file: allocator internals (mcache/mcentral/mheap) and escape-analysis rules → modules 12/05; the 61-tick/runnext/preemption detail → module 11; atomic-vs-mutex table → module 07/14; write the missing pprof section it falsely advertises → module 14.
2. **Why**: Two overlapping texts drift apart (already happened — the modules are *weaker* than the supplement). One canonical, deepest version per topic.
3. **How**: Move unique sections into the module READMEs; delete the file or leave a pointer stub; cross-link related modules (02↔05 slice headers, 06↔11 hchan/GMP).
4. **Outcome**: A learner following the numbered path encounters each topic once, at full depth.

### R6. Rebuild question banks for volume, format, and spaced repetition
1. **What**: Expand from ~1 question/file (~5/module) to 10–20/module; standardize format: question, why-interviewers-ask, model answer, follow-ups, frequency tag; consolidate into one `question-bank.md` per module (difficulty sections inside) or keep files but fill them.
2. **Why**: ~80 total questions cannot cover the interview surface; one-question files are scaffolding artifacts. Consistent format enables self-quizzing and Anki export.
3. **How**: Seed from existing content + syllabus.txt (which already lists the full topic map); tag each with frequency (asked-often/asked-sometimes/rare-but-senior); generate an `anki.tsv` export target in the Makefile.
4. **Outcome**: Learner can run structured mock interviews against themselves with 200+ calibrated questions.

### R7. Rehabilitate `real-interview-q/`: dedupe, complete, de-hype, add system design
1. **What**: Merge duplicate questions (worker pools, atomic counters); complete or trim over-promising intros and the truncated meta.md; delete `interview-questions.md` (fold index into README); relabel questions as "company-*flavored*" rather than "actually asked"; add a `system-design/` section (design a rate limiter, an LRU cache service, a job scheduler — with Go-specific deep dives).
2. **Why**: The coded answers are the repo's best asset, but unverifiable attribution and duplicates erode trust; senior Go interviews include system design, currently 0% covered.
3. **How**: Cross-reference each company question to the module that teaches its foundation; move duplicates to a shared `patterns/` file referenced by both companies; write 3–5 system-design walkthroughs.
4. **Outcome**: Learner gets honest, non-redundant company prep plus the missing system-design dimension.

### R8. Add CI and repo hygiene so quality can't silently rot
1. **What**: GitHub Actions workflow: `go build/test/race/vet/gofmt` all packages **plus compile-check every Go code block embedded in markdown**; `.golangci.yml`; bump `go.mod`; delete or repurpose the orphan `practice-go/`.
2. **Why**: ~25 markdown snippets are currently untested — that's how errors like E9/E11 shipped. Every future edit stays green automatically.
3. **How**: Extract markdown code blocks via a small script (or move snippet code into `examples/` packages with tests and embed via include); workflow matrix on the two latest Go versions; `make ci` runs the same checks locally.
4. **Outcome**: Every code sample in the repo is guaranteed compilable and tested on every commit.

---

## 4. Proposed repository structure

```
golang-interview/
├── README.md                  # Roadmap: 4 levels, gates, how to study, progress checklist
├── syllabus.txt               # Keep as topic map (refresh to current Go)
├── Makefile                   # test, race, exercise T=NN, lab L=NN, anki export, ci
├── go.mod                     # current stable Go
├── .golangci.yml
├── .github/workflows/ci.yml   # build, test -race, vet, fmt, markdown-snippet compile check
│
├── level-1-foundations/       # modules 01–05 (moved, structure below)
├── level-2-concurrency/       # modules 06–10
├── level-3-runtime/           # modules 11–15 + absorbed high-performance-go.md
│   └── NN-topic/
│       ├── README.md          # deep, accurate, cross-linked
│       ├── exercises/         # stubs; failing until implemented
│       ├── solutions/         # reference implementations
│       ├── lab/               # (levels 1&3) guided observation experiments
│       ├── question-bank.md   # 10–20 Qs: why-asked, answer, follow-ups, freq tag
│       └── resources.md       # curated, annotated links
│
├── level-4-applied/           # NEW
│   ├── rate-limiter/
│   ├── kv-store/
│   ├── job-queue/
│   └── mini-redis/            # spec + stubs + hidden acceptance tests + benchmarks
│
└── interview-bank/            # was real-interview-q/
    ├── README.md              # single index
    ├── by-topic/              # dedup'd canonical pattern questions
    ├── by-company/            # existing company files, cleaned & cross-linked
    └── system-design/         # NEW: 3–5 Go-centric design walkthroughs
```

---

## 5. Prioritized roadmap

- **P0 — Stop the bleeding (small, high value)**: Fix factual errors E1–E11; complete truncated READMEs (06/07/09/10 gotcha lists, meta.md); delete redundant `interview-questions.md`; fix gofmt violation; fix misleading `Stack[T]` comment and `NonBlockingReceive`; bump `go.mod`.
- **P1 — Make practice real**: R1 exercise-first conversion for all 15 modules; R8 CI incl. markdown-snippet compilation.
- **P2 — Modernize + deepen content**: R3 modern topics (context 1.21 APIs, Swiss maps, fuzzing, PGO, slog/iterators); R5 merge high-performance-go.md; R2 labs for 05/11/12/14.
- **P3 — Build the expert tier**: R4 capstones (rate limiter first, then KV store, job queue, mini-redis); R7 system-design section.
- **P4 — Scale the banks**: R6 question-bank expansion + Anki export; R7 company-file dedup/completion.

---

## 6. Execution approaches (choose at approval)

The review above is the complete deliverable of the analysis. On approval, choose how much to execute now:

1. **Review only** — save this document into the repo (e.g., `REVIEW.md`) and stop; implementation is yours to schedule.
2. **P0 critical-fix pass** — additionally execute the full P0 list (~15 targeted edits: factual corrections, truncation completion, redundancy removal, go.mod bump). No restructuring.
3. **P0 + P1** — critical fixes plus the exercise-first conversion of all 15 modules and CI setup. (Larger effort; restructuring into levels/capstones/labs deferred to a follow-up session.)
