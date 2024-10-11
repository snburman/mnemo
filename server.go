package mnemo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var srvMgr = &srvManager{
	servers: make(map[int]*Server),
}

type (
	srvManager struct {
		mu sync.Mutex
		// Servers stored by port
		servers map[int]*Server
	}
	// Server is a websocket server that manages connections and messages.
	Server struct {
		mu    sync.Mutex
		mnemo *Mnemo
		http  *http.Server
		context.Context
		cfg             serverConfig
		msgs            chan []byte
		onNewConnection func(c *Conn)
		connPool        *Pool
	}
	serverConfig struct {
		Port    int
		Pattern string
		Key     string
	}
)

func (s *Server) withNemo(m *Mnemo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mnemo != nil {
		NewError[Server]("mnemo already initialized").WithLogLevel(Fatal).Log()
	}
	s.mnemo = m
}

// WithPort sets the port for the server.
func WithPort(port int) Opt[Server] {
	if port < 1025 || port > 65535 {
		NewError[Server](
			fmt.Sprintf("port must be between 1025 and 65535. Got: %d", port),
		).WithLogLevel(Fatal).Log()
	}
	return func(s *Server) {
		s.cfg.Port = port
		s.http.Addr = ":" + strconv.Itoa(port)
	}
}

// WithPattern sets the pattern for the server's http handler.
func WithPattern(pattern string) Opt[Server] {
	return func(s *Server) {
		s.cfg.Pattern = pattern
	}
}

// WithSilence silences the server's http error log.
func WithSilence() Opt[Server] {
	return func(s *Server) {
		s.http.ErrorLog = log.New(nil, "", 0)
	}
}

// NewServer creates a new server.
//
// The server's key must be unique. If a server with the same key
// already exists, an error is returned.
func NewServer(key string, opts ...Opt[Server]) (*Server, error) {
	mux := http.NewServeMux()
	cfg := serverConfig{
		Key:     key,
		Port:    srvMgr.AssignPort(),
		Pattern: "/" + key,
	}
	srv := &Server{
		http: &http.Server{
			Addr:           ":" + strconv.Itoa(cfg.Port),
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
			ErrorLog:       logger.StandardLog(),
		},
		Context:  context.Background(),
		cfg:      cfg,
		msgs:     make(chan []byte, 16),
		connPool: NewPool(),
	}

	for _, o := range opts {
		o(srv)
	}

	srvMgr.mu.Lock()
	defer srvMgr.mu.Unlock()

	for port, sv := range srvMgr.servers {
		if port == srv.cfg.Port {
			return nil, NewError[Server](fmt.Sprintf("server with port '%d' already exists", srv.cfg.Port))
		}
		if sv.cfg.Key == srv.cfg.Key {
			return nil, NewError[Server](fmt.Sprintf("server with key '%s' already exists", srv.cfg.Key))
		}
	}
	srvMgr.servers[srv.cfg.Port] = srv

	mux.HandleFunc(srv.cfg.Pattern+"/subscribe", srv.HandleSubscribe)

	return srv, nil
}

// ListenAndServe starts the server in a go routine
func (s *Server) ListenAndServe() {
	s.http.RegisterOnShutdown(func() {
		s.connPool.Close()
	})
	// print server info
	logger.Infof("server listening on port %d", s.cfg.Port)
	go s.http.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Remove server from server manager
	srvMgr.mu.Lock()
	delete(srvMgr.servers, s.cfg.Port)
	srvMgr.mu.Unlock()

	s.mu.Lock()
	err := s.http.Shutdown(s.Context)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return nil
}

// HandleSubscribe upgrades the http connection to a websocket connection
// and adds the connection to the connection pool.
func (s *Server) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	conn, err := NewConn(w, r)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := s.connPool.AddConn(conn); err != nil {
		log.Fatal(err)
	}
	// Trigger user defined call back on new connection
	if s.onNewConnection != nil {
		go s.onNewConnection(conn)
	}
	conn.Listen()
}

// SetOnNewConnection sets a user defined call back function
// when a new connection is established.
//
// Mnemo must be initialized with a server before calling this method.
func (s *Server) SetOnNewConnection(fn func(c *Conn)) {
	s.onNewConnection = fn
}

// Publish publishes a message to all connections in the connection pool.
func (s *Server) Publish(msg interface{}) {
	for _, conn := range s.connPool.Conns() {
		select {
		case conn.Messages <- msg:
		default:
			log.Println("closing connection: ", conn.Key)
			conn.Close()
		}
	}
}

// AssignPort assigns a port to the server.
//
// Port is assigned by incrementing the highest port number
// in the server manager's servers map, starting at 8080.
func (s *srvManager) AssignPort() int {
	next := 8080
	for port := range s.servers {
		if next < port {
			next = port + 1
		}
	}
	return next
}
