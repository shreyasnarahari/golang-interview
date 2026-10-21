package solution

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- test helpers -------------------------------------------------------
//
// sendCmd, readResp, and getBulk return errors so they are safe to call from
// client goroutines (t.Fatalf may only run on the test's own goroutine). The
// do/get wrappers below are Fatal-based conveniences for the main goroutine.

// sendCmd writes one command line.
func sendCmd(conn net.Conn, cmd string) error {
	_, err := io.WriteString(conn, cmd+"\r\n")
	return err
}

// readResp reads one \r\n-terminated response line without its terminator.
func readResp(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// getBulk reads a bulk-string reply, reporting ok=false for "$-1".
func getBulk(r *bufio.Reader) (value string, ok bool, err error) {
	header, err := readResp(r)
	if err != nil {
		return "", false, err
	}
	if header == "$-1" {
		return "", false, nil
	}
	var n int
	if _, err := fmt.Sscanf(header, "$%d", &n); err != nil {
		return "", false, fmt.Errorf("malformed bulk header %q: %w", header, err)
	}
	buf := make([]byte, n+2) // value + trailing \r\n
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", false, err
	}
	return string(buf[:n]), true, nil
}

// do sends cmd and reads a single-line response (+OK, :1, -ERR ..., ...).
func do(t *testing.T, conn net.Conn, r *bufio.Reader, cmd string) string {
	t.Helper()
	if err := sendCmd(conn, cmd); err != nil {
		t.Fatalf("write %q: %v", cmd, err)
	}
	resp, err := readResp(r)
	if err != nil {
		t.Fatalf("read response to %q: %v", cmd, err)
	}
	return resp
}

// get sends GET key and decodes the bulk-string reply.
func get(t *testing.T, conn net.Conn, r *bufio.Reader, key string) (string, bool) {
	t.Helper()
	if err := sendCmd(conn, "GET "+key); err != nil {
		t.Fatalf("write GET: %v", err)
	}
	v, ok, err := getBulk(r)
	if err != nil {
		t.Fatalf("read GET %q reply: %v", key, err)
	}
	return v, ok
}

// startServer boots a Server on an ephemeral port and registers cleanup that
// cancels it and asserts ListenAndServe returned nil (graceful shutdown).
func startServer(t *testing.T) (*Server, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	srv := NewServer()

	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe(ctx, "127.0.0.1:0") }()

	waitAddr(t, srv.Addr)
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("ListenAndServe() = %v, want nil on graceful shutdown", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("ListenAndServe did not return within 2s of cancel")
		}
	})
	return srv, cancel
}

// waitAddr polls get until the server reports a non-nil bound address.
func waitAddr(t *testing.T, get func() net.Addr) net.Addr {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := get(); addr != nil {
			return addr
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("server did not bind a listener within 2s")
	return nil
}

// dial opens a client connection to the server's bound address.
func dial(t *testing.T, srv *Server) (net.Conn, *bufio.Reader) {
	t.Helper()
	conn, err := net.Dial("tcp", srv.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn, bufio.NewReader(conn)
}

// --- acceptance tests ---------------------------------------------------

func TestPingPong(t *testing.T) {
	srv, _ := startServer(t)
	conn, r := dial(t, srv)

	if got := do(t, conn, r, "PING"); got != "+PONG" {
		t.Fatalf("PING = %q, want %q", got, "+PONG")
	}
	// Lowercase commands must also work; keep-alive across many round trips.
	for i := 0; i < 10; i++ {
		if got := do(t, conn, r, "ping"); got != "+PONG" {
			t.Fatalf("ping #%d = %q, want %q", i, got, "+PONG")
		}
	}
}

func TestSetGetDel(t *testing.T) {
	srv, _ := startServer(t)
	conn, r := dial(t, srv)

	if got := do(t, conn, r, "SET name gopher"); got != "+OK" {
		t.Fatalf("SET = %q, want +OK", got)
	}
	if v, ok := get(t, conn, r, "name"); !ok || v != "gopher" {
		t.Fatalf("GET name = %q, %v; want %q, true", v, ok, "gopher")
	}
	if _, ok := get(t, conn, r, "missing"); ok {
		t.Fatal("GET missing: want nil bulk reply ($-1), got a value")
	}
	if got := do(t, conn, r, "DEL name"); got != ":1" {
		t.Fatalf("DEL name = %q, want :1", got)
	}
	if got := do(t, conn, r, "DEL name"); got != ":0" {
		t.Fatalf("DEL name (again) = %q, want :0", got)
	}
	if _, ok := get(t, conn, r, "name"); ok {
		t.Fatal("GET name after DEL: want $-1, got a value")
	}
	// Multi-key DEL counts only keys that existed.
	if got := do(t, conn, r, "SET a 1"); got != "+OK" {
		t.Fatalf("SET a = %q, want +OK", got)
	}
	if got := do(t, conn, r, "SET b 2"); got != "+OK" {
		t.Fatalf("SET b = %q, want +OK", got)
	}
	if got := do(t, conn, r, "DEL a b nope"); got != ":2" {
		t.Fatalf("DEL a b nope = %q, want :2", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	srv, _ := startServer(t)
	conn, r := dial(t, srv)

	if got := do(t, conn, r, "FOOBAR"); !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("FOOBAR = %q, want -ERR ...", got)
	}
	if got := do(t, conn, r, "SET onlykey"); !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("SET with one arg = %q, want -ERR ...", got)
	}
	if got := do(t, conn, r, ""); !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("empty line = %q, want -ERR ...", got)
	}
	// The connection must stay usable after errors.
	if got := do(t, conn, r, "PING"); got != "+PONG" {
		t.Fatalf("PING after errors = %q, want +PONG", got)
	}
}

func TestConcurrentClients(t *testing.T) {
	srv, _ := startServer(t)

	const clients = 16
	const opsPerClient = 25

	var wg sync.WaitGroup
	for c := 0; c < clients; c++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", srv.Addr().String())
			if err != nil {
				t.Errorf("client %d dial: %v", c, err)
				return
			}
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)

			key := fmt.Sprintf("key-%d", c)
			for i := 0; i < opsPerClient; i++ {
				want := fmt.Sprintf("value-%d-%d", c, i)
				if err := sendCmd(conn, "SET "+key+" "+want); err != nil {
					t.Errorf("client %d SET write: %v", c, err)
					return
				}
				if resp, err := readResp(r); err != nil || resp != "+OK" {
					t.Errorf("client %d SET = %q, %v; want +OK", c, resp, err)
					return
				}
				// All clients also hammer one shared key to stress the store mutex.
				if err := sendCmd(conn, fmt.Sprintf("SET shared %d", i)); err != nil {
					t.Errorf("client %d SET shared write: %v", c, err)
					return
				}
				if resp, err := readResp(r); err != nil || resp != "+OK" {
					t.Errorf("client %d SET shared = %q, %v; want +OK", c, resp, err)
					return
				}
				if err := sendCmd(conn, "GET "+key); err != nil {
					t.Errorf("client %d GET write: %v", c, err)
					return
				}
				got, ok, err := getBulk(r)
				if err != nil || !ok || got != want {
					t.Errorf("client %d GET %s = %q, %v, %v; want %q", c, key, got, ok, err, want)
					return
				}
			}
			if err := sendCmd(conn, "DEL "+key); err != nil {
				t.Errorf("client %d DEL write: %v", c, err)
				return
			}
			if resp, err := readResp(r); err != nil || resp != ":1" {
				t.Errorf("client %d DEL %s = %q, %v; want :1", c, key, resp, err)
			}
		}(c)
	}
	wg.Wait()

	stats := srv.Stats()
	if stats.ConnectionsTotal < clients {
		t.Errorf("ConnectionsTotal = %d, want >= %d", stats.ConnectionsTotal, clients)
	}
	if stats.CommandsProcessed < clients*opsPerClient*3 {
		t.Errorf("CommandsProcessed = %d, want >= %d", stats.CommandsProcessed, clients*opsPerClient*3)
	}
}

func TestGracefulShutdown(t *testing.T) {
	srv, cancel := startServer(t)
	conn, r := dial(t, srv)

	// A round trip first guarantees the server has accepted and tracked this
	// connection before we cancel.
	if got := do(t, conn, r, "PING"); got != "+PONG" {
		t.Fatalf("PING = %q, want +PONG", got)
	}

	cancel()

	// The open connection must be closed by the server during drain.
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := r.ReadString('\n'); err == nil {
		t.Error("expected read on drained connection to fail after shutdown")
	}

	// The listener must be closed: new dials fail.
	addr := srv.Addr().String()
	deadline := time.Now().Add(2 * time.Second)
	for {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			break // connection refused: listener is gone
		}
		_ = c.Close()
		if time.Now().After(deadline) {
			t.Fatal("listener still accepting connections after shutdown")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := srv.Stats().ConnectionsActive; got != 0 {
		t.Errorf("ConnectionsActive = %d after shutdown, want 0", got)
	}
}

func TestStats(t *testing.T) {
	srv, _ := startServer(t)

	c1, r1 := dial(t, srv)
	c2, r2 := dial(t, srv)

	do(t, c1, r1, "PING")
	do(t, c2, r2, "PING")
	do(t, c2, r2, "SET k v")

	stats := srv.Stats()
	if stats.CommandsProcessed != 3 {
		t.Errorf("CommandsProcessed = %d, want 3", stats.CommandsProcessed)
	}
	if stats.ConnectionsTotal != 2 {
		t.Errorf("ConnectionsTotal = %d, want 2", stats.ConnectionsTotal)
	}
	if stats.ConnectionsActive != 2 {
		t.Errorf("ConnectionsActive = %d, want 2", stats.ConnectionsActive)
	}
}

func TestPprofEndpoint(t *testing.T) {
	srv, _ := startServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServePprof(ctx, "127.0.0.1:0") }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("ListenAndServePprof() = %v, want nil", err)
		}
	})

	addr := waitAddr(t, srv.PprofAddr)
	resp, err := http.Get("http://" + addr.String() + "/debug/pprof/cmdline")
	if err != nil {
		t.Fatalf("GET /debug/pprof/cmdline: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /debug/pprof/cmdline = %s, want 200 OK", resp.Status)
	}
}
