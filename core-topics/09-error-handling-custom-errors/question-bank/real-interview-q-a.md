# Topic 09 Real Interview Q&A

### Q: (Cloud infra company) "A caller of your storage client gets an error. Walk me through how they distinguish 'file not found' from 'permission denied' — and how you, as the library author, make that possible."
**Frequency**: often
**Why interviewers ask**: Cloud SDKs live or die by whether callers can react programmatically to errors; this tests end-to-end API design, not just syntax.
**Answer**: Model answer covers both sides:
- *Caller side*: `errors.Is(err, os.ErrNotExist)` for equivalence against a sentinel, `errors.As(err, &pathErr)` with `var pathErr *os.PathError` to extract structured detail (`Op`, `Path`, underlying `Err`). `Is` compares across the wrap chain; `As` extracts a concrete type from it.
- *Author side*: wrap, don't replace — return `fmt.Errorf("storage: get %q: %w", key, err)` so the caller's `errors.Is(err, os.ErrNotExist)` still works through my layers. Publish a small documented taxonomy: exported sentinels (`ErrKeyNotFound`) for conditions callers handle, exported types only when callers need fields. Never make callers string-match `Error()` output — that freezes my message wording as API.
Mention `os.ErrPermission`/`os.ErrExist` as the portable sentinels the stdlib provides over platform-specific `syscall.Errno` values, and that permission-denied surfaces as `errors.Is(err, fs.ErrPermission)`.
**Follow-ups**: What changes if the storage backend is remote and errors cross an RPC boundary? Why is string-matching error messages an anti-pattern?

### Q: (Payments platform) "Design the error model for a payment-processing SDK. How do callers tell retryable from non-retryable failures?"
**Frequency**: often
**Why interviewers ask**: Payments systems must classify failures (retry vs. fail-fast vs. reconcile), so error design directly affects money correctness.
**Answer**: A layered design:
```go
var ErrInsufficientFunds = errors.New("insufficient funds") // sentinel: business outcome

type Error struct {
    Code      Code // machine-readable enum: CodeRateLimited, CodeCardDeclined, ...
    Retryable bool
    Op        string
    Err       error // wrapped cause
}
func (e *Error) Error() string { return e.Op + ": " + e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }
```
- Callers use `errors.As(err, &sdkErr)` to read `Retryable`/`Code` for retry decisions, and `errors.Is(err, ErrInsufficientFunds)` for business outcomes they must surface.
- Sentinels for stable business conditions; the typed `Error` for operational metadata. Keep the exported surface tiny — every exported code is a compatibility promise.
- Optionally an `Is(error) bool` override so `errors.Is(err, &Error{Code: CodeCardDeclined})` matches by code regardless of other fields.
- For batch charges, aggregate per-item failures with `errors.Join` so callers can `As`-extract each item's `*Error`.
Emphasize: never encode retryability only in message text, and document which codes are stable.
**Follow-ups**: How would you version this error model as the SDK evolves? Where does `context.DeadlineExceeded` fit in the retry decision?

### Q: (Database company) "Errors from your distributed database cross RPC boundaries and arrive at the client as plain strings. How do you preserve matchability, and how do you implement unwrapping on your custom error types?"
**Frequency**: sometimes
**Why interviewers ask**: Distributed systems destroy error identity at serialization boundaries; designing around that is a real senior-level concern.
**Answer**: Two parts:
- *Unwrapping locally*: implement `Unwrap() error` on the custom struct so `errors.Is`/`errors.As` traverse it:
```go
type RetryableError struct{ Cause error }
func (e *RetryableError) Error() string { return "retryable: " + e.Cause.Error() }
func (e *RetryableError) Unwrap() error { return e.Cause }
```
- *Across the wire*: `Unwrap` cannot survive serialization — the concrete types don't exist on the client. So the wire format carries *stable identifiers*, not types: an error code enum (like gRPC `status.Code` or `cockroachdb/errors`-style markers), a message, and causal metadata. The client reconstructs typed errors from the code and re-attaches matchability via sentinels (`errors.Is(err, ErrTransactionRetry)`) or a code-matching `Is(error) bool` override. Principle: inside one process, use wrap chains and `Is`/`As`; across processes, use codes, and translate codes back into the local sentinel/typed taxonomy at the boundary.
**Follow-ups**: Why can't wrapped Go errors just be serialized with gob/JSON? How does gRPC's `status.FromError` fit this pattern?
