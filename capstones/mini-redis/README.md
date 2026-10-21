# Capstone: mini-redis

Build a miniature Redis-like TCP server. This capstone exercises everything
from the concurrency tier in one place: `net` listeners, goroutine-per-
connection serving, protocol parsing, shared state behind a mutex, graceful
shutdown driven by `context`, and production-grade observability
(`net/http/pprof` + atomic counters).

This is a classic "real interview" shape: it is small enough to finish in
45–60 minutes but deep enough that every design choice has a follow-up
question attached to it.

## Wire protocol

A RESP-*inspired* subset. RESP proper uses `$<len>` bulk strings and `*<n>`
arrays for requests; we deliberately simplify requests to one text command
per line so the parser is a `bufio.Reader` loop, while keeping RESP-style
*responses* so the wire format still teaches type-tagged replies.

Request: one command per line, terminated by `\n` or `\r\n`. Tokens are
space-separated, so keys and values may not contain spaces (a documented
limitation — see Trade-offs).

```
PING              -> "+PONG\r\n"
SET key value     -> "+OK\r\n"
GET key           -> "$<len>\r\n<value>\r\n" or "$-1\r\n" when missing
DEL key [key ...] -> ":<n>\r\n" (count of keys that actually existed)
anything else     -> "-ERR <message>\r\n"
```

Rules:

- Command names are case-insensitive (`ping` works).
- Wrong argument counts return `-ERR ...`; an empty line returns `-ERR`.
- Errors never close the connection; the client may keep sending commands.
- One connection may carry unlimited sequential commands (keep-alive).

## API contract

Top-level package `miniredis`:

```go
srv := miniredis.NewServer()
go srv.ListenAndServe(ctx, "127.0.0.1:0")   // blocks until ctx is canceled
addr := srv.Addr()                           // nil until bound
stats := srv.Stats()                         // atomic-counter snapshot
go srv.ListenAndServePprof(ctx, "127.0.0.1:0") // pprof on a separate port
```

- `NewServer() *Server` — construct a server backed by an empty `Store`.
- `(*Server).ListenAndServe(ctx, addr) error` — bind, accept, serve. On
  `ctx` cancellation: close the listener, close all in-flight connections
  (unblocking their reads), wait for every handler goroutine to exit, then
  return `nil`. A listener failure unrelated to shutdown is returned.
- `(*Server).Addr() net.Addr` — the bound address (supports `:0` ephemeral
  ports, which the tests rely on).
- `(*Server).ListenAndServePprof(ctx, addr) error` — serve the
  `net/http/pprof` handlers (already registered on `http.DefaultServeMux`
  by importing the package) on a *separate* port; `nil` on graceful
  shutdown.
- `(*Server).Stats() Stats` — snapshot of `CommandsProcessed`,
  `ConnectionsTotal`, `ConnectionsActive`.

Backing store (build it here; do **not** import another capstone):

```go
s := miniredis.NewStore()
s.Set("k", "v")
v, ok := s.Get("k")
n := s.Del("k", "missing") // returns how many keys actually existed
```

## Requirements checklist

1. **Goroutine-per-connection**: `Accept` loop spawns one goroutine per
   client; each connection has its own `bufio.Reader` and read/execute/
   write loop.
2. **Thread-safe store**: a plain `map[string]string` guarded by
   `sync.RWMutex`. Reads (GET) take the read lock; writes take the write
   lock.
3. **Graceful shutdown**: `ctx` cancel → close listener → close tracked
   connections → `sync.WaitGroup` waits for all handlers → return `nil`.
   After shutdown, `Stats().ConnectionsActive == 0` and new dials are
   refused.
4. **Observability**: `sync/atomic` counters (commands processed,
   connections total/active) and a pprof HTTP server on a separate port.
5. **Race-clean**: `go test -race ./capstones/mini-redis/solution/` passes.

## Design trade-offs (be ready to defend these)

- **Why a text request line instead of full RESP?** A `*3\r\n$3\r\nSET...`
  parser needs length-prefixed reading and state; the interview value is in
  the server architecture, not the tokenizer. Say this out loud, and name
  the limitation: no spaces or binary data in keys/values. The bulk-string
  *response* format shows you know the real protocol.
- **Why one mutex for the whole map?** Correct and simple. Follow-up:
  "how would you scale it?" → sharded maps (`N` stripe locks keyed by
  `hash(key) % N`), which is what Redis-adjacent systems do before going
  single-threaded-with-event-loop.
- **Why track open connections in a map instead of just closing the
  listener?** Closing the listener stops *new* connections but blocked
  `ReadString` calls on existing connections would hang forever; the
  `WaitGroup.Wait()` would never return. You must actively close in-flight
  conns to unblock their reads. This is the single most-missed piece.
- **Why pprof on a separate port?** Profiling endpoints are unauthenticated
  and can expose internals (goroutine stacks, heap). Never share a port —
  or a mux — with the data path.
- **Why `atomic.Int64` counters instead of expvar?** Both are acceptable;
  atomics are typed and cheap, expvar gives you a free HTTP endpoint.
  Mention expvar as the alternative.

## What the interviewer is listening for

- You state the **shutdown ordering** before writing code: listener, then
  conns, then wait — and you can explain *why* each step is needed.
- You know `Accept` returning an error after `Close` is expected, and you
  distinguish "closed because ctx canceled" (return `nil`) from a real
  failure (`errors.Is(err, net.ErrClosed)` / `ctx.Err()`).
- Your per-connection goroutine has a `defer` that unregisters the conn,
  decrements the active gauge, closes the conn, and calls `wg.Done()` — so
  counters can't leak even on panic-adjacent paths.
- `t.Cleanup`-style reasoning about goroutine leaks: after cancel, prove
  every goroutine exits (the `WaitGroup` is the proof; pprof's
  `/debug/pprof/goroutine?debug=1` is how you'd verify in production).
- You mention race-detector validation unprompted, and you know which
  fields are guarded by the mutex vs. which are atomic (and why mixing
  them is fine as long as each field has exactly one discipline).
- Bonus: you can talk about what real Redis does differently —
  single-threaded event loop, no mutex at all, expiry via active + passive
  eviction, `UNLINK` vs `DEL`.

## Stretch goals

- **Idle timeout**: `conn.SetReadDeadline` refreshed per command; closes
  slowloris-style idle clients. Question: what happens to your shutdown
  drain when deadlines are already unblocking reads?
- **Full RESP request parsing**: accept `*<n>\r\n$<len>\r\n...` arrays so
  the real `redis-cli` can talk to your server.
- **Key expiration**: `SET key value EX seconds` backed by a min-heap of
  deadlines (see the kv-store capstone).
- **Sharding**: split the store into `k` striped maps and measure the
  contention difference under `-race` + `go test -bench`.
- **Metrics endpoint**: serve `/debug/vars` (expvar) or a JSON `/stats`
  next to pprof.
- **Backpressure / slow clients**: per-conn write deadlines; what do you
  do when a client never reads? (Redis has `client-output-buffer-limit`.)

## Files

- `server.go` — exercise stubs (`panic("TODO: implement ...")`). Make the
  tests green.
- `server_test.go` — acceptance tests: PING/PONG round-trip, SET/GET/DEL,
  unknown-command errors, 16 concurrent clients, graceful shutdown, stats,
  pprof endpoint. Red until you implement.
- `solution/` — reference implementation, green including `-race`.

## Running

```sh
go test ./capstones/mini-redis/            # your implementation (red until done)
go test -race ./capstones/mini-redis/solution/   # reference, green
```
