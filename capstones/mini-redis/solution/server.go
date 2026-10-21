// Package solution is the reference implementation of the mini-redis
// capstone: a miniature Redis-like TCP server speaking a line-based,
// RESP-inspired protocol.
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
package solution

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
	"strings"
	"sync"
	"sync/atomic"
)

const (
	respOK   = "+OK\r\n"
	respPong = "+PONG\r\n"
	respNil  = "$-1\r\n"
)

// Store is a thread-safe in-memory string key-value store.
type Store struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewStore returns an empty Store ready for use.
func NewStore() *Store {
	return &Store{m: make(map[string]string)}
}

// Get returns the value for key and whether it was present.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

// Set stores value under key, overwriting any previous value.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

// Del removes keys and returns how many of them actually existed.
func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for _, k := range keys {
		if _, ok := s.m[k]; ok {
			delete(s.m, k)
			removed++
		}
	}
	return removed
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
	store *Store

	commands    atomic.Int64
	connsTotal  atomic.Int64
	connsActive atomic.Int64

	mu    sync.Mutex // guards addr, pprofAddr, and conns
	addr  net.Addr
	paddr net.Addr
	conns map[net.Conn]struct{}
	wg    sync.WaitGroup
}

// NewServer returns a Server backed by an empty Store.
func NewServer() *Server {
	return &Server{store: NewStore()}
}

// ListenAndServe binds a TCP listener on addr and serves it, one goroutine
// per connection, until ctx is canceled or the listener fails.
//
// On ctx cancellation the listener is closed, in-flight connections are
// closed (unblocking their reads), and ListenAndServe waits for every
// connection handler to exit before returning nil. Any other listener
// failure is returned to the caller.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.addr = ln.Addr()
	s.conns = make(map[net.Conn]struct{})
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.closeAndDrain()
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.connsTotal.Add(1)
		s.connsActive.Add(1)
		s.trackConn(conn, true)
		s.wg.Add(1)
		go s.serveConn(conn)
	}
}

// Addr reports the address the main listener is bound to. It returns nil
// until ListenAndServe has bound its listener.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// ListenAndServePprof serves the net/http/pprof handlers (registered on
// http.DefaultServeMux) on a separate address, so profiling endpoints never
// share a port with the data path. It shuts down when ctx is canceled and
// returns nil on graceful shutdown.
func (s *Server) ListenAndServePprof(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.paddr = ln.Addr()
	s.mu.Unlock()

	srv := &http.Server{Handler: http.DefaultServeMux}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// PprofAddr reports the address the pprof listener is bound to. It returns
// nil until ListenAndServePprof has bound its listener.
func (s *Server) PprofAddr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paddr
}

// Stats returns a snapshot of the server's atomic counters.
func (s *Server) Stats() Stats {
	return Stats{
		CommandsProcessed: s.commands.Load(),
		ConnectionsTotal:  s.connsTotal.Load(),
		ConnectionsActive: s.connsActive.Load(),
	}
}

// trackConn adds or removes conn from the active-connection set.
func (s *Server) trackConn(conn net.Conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		s.conns[conn] = struct{}{}
	} else {
		delete(s.conns, conn)
	}
}

// closeAndDrain closes every tracked connection so blocked reads return,
// then waits for all connection handlers to finish.
func (s *Server) closeAndDrain() {
	s.mu.Lock()
	for conn := range s.conns {
		_ = conn.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
}

// serveConn handles one client connection until the peer disconnects, the
// connection errors, or the server shuts down.
func (s *Server) serveConn(conn net.Conn) {
	defer func() {
		s.trackConn(conn, false)
		s.connsActive.Add(-1)
		_ = conn.Close()
		s.wg.Done()
	}()

	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		if _, err := io.WriteString(conn, s.execute(line)); err != nil {
			return
		}
	}
}

// execute parses one command line and returns the encoded response.
func (s *Server) execute(line string) string {
	s.commands.Add(1)

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "-ERR empty command\r\n"
	}
	cmd, args := strings.ToUpper(fields[0]), fields[1:]

	switch cmd {
	case "PING":
		if len(args) != 0 {
			return "-ERR wrong number of arguments for 'ping'\r\n"
		}
		return respPong
	case "SET":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'set' (usage: SET key value)\r\n"
		}
		s.store.Set(args[0], args[1])
		return respOK
	case "GET":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'get'\r\n"
		}
		v, ok := s.store.Get(args[0])
		if !ok {
			return respNil
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	case "DEL":
		if len(args) == 0 {
			return "-ERR wrong number of arguments for 'del'\r\n"
		}
		return fmt.Sprintf(":%d\r\n", s.store.Del(args...))
	default:
		return fmt.Sprintf("-ERR unknown command %q\r\n", cmd)
	}
}
