// Package miniredis is the exercise stub for the mini-redis capstone: a
// miniature Redis-like TCP server speaking a line-based, RESP-inspired
// protocol. Implement every function below; see README.md for the spec.
//
// Wire protocol: one command per line, terminated by \n or \r\n.
// Commands are space-separated tokens (keys and values may not contain
// spaces):
//
//	PING              -> "+PONG\r\n"
//	SET key value     -> "+OK\r\n"
//	GET key           -> "$<len>\r\n<value>\r\n" or "$-1\r\n" when missing
//	DEL key [key ...] -> ":<n>\r\n" (number of keys actually removed)
//	anything else     -> "-ERR <message>\r\n"
package miniredis

import (
	"context"
	"net"
)

// Store is a thread-safe in-memory string key-value store.
type Store struct {
	// TODO: add the fields your implementation needs (map, mutex, ...).
}

// NewStore returns an empty Store ready for use.
func NewStore() *Store {
	panic("TODO: implement NewStore")
}

// Get returns the value for key and whether it was present.
func (s *Store) Get(key string) (string, bool) {
	panic("TODO: implement Get")
}

// Set stores value under key, overwriting any previous value.
func (s *Store) Set(key, value string) {
	panic("TODO: implement Set")
}

// Del removes keys and returns how many of them actually existed.
func (s *Store) Del(keys ...string) int {
	panic("TODO: implement Del")
}

// Stats is a point-in-time snapshot of the server's counters.
type Stats struct {
	// CommandsProcessed counts every command line executed.
	CommandsProcessed int64
	// ConnectionsTotal counts every accepted connection.
	ConnectionsTotal int64
	// ConnectionsActive counts connections currently being served.
	ConnectionsActive int64
}

// Server is a mini Redis-like TCP server. The zero value is not usable;
// construct one with NewServer.
type Server struct {
	// TODO: add the fields your implementation needs (store, listener
	// state, connection tracking, WaitGroup, atomic counters, ...).
}

// NewServer returns a Server backed by an empty Store.
func NewServer() *Server {
	panic("TODO: implement NewServer")
}

// ListenAndServe binds a TCP listener on addr and serves it, one goroutine
// per connection, until ctx is canceled or the listener fails.
//
// On ctx cancellation the listener is closed, in-flight connections are
// closed (unblocking their reads), and ListenAndServe waits for every
// connection handler to exit before returning nil. Any other listener
// failure is returned to the caller.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	panic("TODO: implement ListenAndServe")
}

// Addr reports the address the main listener is bound to. It returns nil
// until ListenAndServe has bound its listener.
func (s *Server) Addr() net.Addr {
	panic("TODO: implement Addr")
}

// ListenAndServePprof serves the net/http/pprof handlers (registered on
// http.DefaultServeMux) on a separate address, so profiling endpoints never
// share a port with the data path. It shuts down when ctx is canceled and
// returns nil on graceful shutdown.
func (s *Server) ListenAndServePprof(ctx context.Context, addr string) error {
	panic("TODO: implement ListenAndServePprof")
}

// PprofAddr reports the address the pprof listener is bound to. It returns
// nil until ListenAndServePprof has bound its listener.
func (s *Server) PprofAddr() net.Addr {
	panic("TODO: implement PprofAddr")
}

// Stats returns a snapshot of the server's atomic counters.
func (s *Server) Stats() Stats {
	panic("TODO: implement Stats")
}
