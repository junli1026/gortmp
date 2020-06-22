package rtmp

import (
	"errors"
	"net"
	"sync"
	"time"

	l "github.com/junli1026/gortmp/logging"
)

type serverImpl interface {
	newContext(conn net.Conn) interface{}
	read(data []byte, context interface{}) (int, []byte, error)
}

type connHandler struct {
	conn    net.Conn
	readbuf []byte
	s       *baseServer
	context interface{}
}

func newHandler(conn net.Conn, s *baseServer) *connHandler {
	handler := &connHandler{
		conn:    conn,
		readbuf: make([]byte, 0),
		s:       s,
		context: s.impl.newContext(conn),
	}
	return handler
}

func (h *connHandler) read() error {
	var buf = make([]byte, 1024*10)
	l.Logger.Info("connection handler starts working\n")

	// if not data for 60 seconds, we close conn with error timeout
	if err := h.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		l.Logger.Error(err)
		return err
	}
	length, err := h.conn.Read(buf)
	if err != nil {
		l.Logger.Error(err)
		return err
	}
	h.readbuf = append(h.readbuf, buf[:length]...)
	return nil
}

func (h *connHandler) writeAll(data []byte) error {
	for data != nil && len(data) > 0 {
		length, err := h.conn.Write(data)
		if err != nil {
			l.Logger.Error(err)
			return err
		}
		data = data[length:]
	}
	return nil
}

func (h *connHandler) run() {
	h.s.wg.Add(1)
	defer h.s.wg.Done()
	for {
		if err := h.read(); err != nil {
			return
		}

		for {
			length, reply, err := h.s.impl.read(h.readbuf, h.context)
			if err != nil {
				l.Logger.Error(err)
				return
			}

			if err = h.writeAll(reply); err != nil {
				return
			}

			if length != 0 {
				h.readbuf = h.readbuf[length:]
			} else {
				break
			}
		}
	}
}

type serverState int

const (
	running  serverState = 0
	stopping serverState = 1
	stopped  serverState = 2
)

type baseServer struct {
	addr     string
	listener *net.TCPListener
	mux      sync.Mutex
	state    serverState
	stopch   chan int
	impl     serverImpl
	wg       sync.WaitGroup
}

func newBaseServer(addr string, impl serverImpl) *baseServer {
	return &baseServer{
		addr:     addr,
		listener: nil,
		state:    stopped,
		wg:       sync.WaitGroup{},
		impl:     impl,
	}
}

func (s *baseServer) run() error {
	s.mux.Lock()
	if s.state != stopped {
		defer s.mux.Unlock()
		return errors.New("server is not in 'stopped' state")
	}
	s.state = running
	s.mux.Unlock()

	address, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		l.Logger.Fatal(err)
	}

	s.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
		l.Logger.Fatal(err)
	}
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mux.Lock()
			if s.state == stopping {
				s.mux.Unlock()
				break
			}
			s.mux.Unlock()
			l.Logger.Warning(err)
			continue
		}
		defer conn.Close()

		h := newHandler(conn, s)
		go h.run()
	}
	return nil
}

func (s *baseServer) stop() {
	s.mux.Lock()
	if s.state != running {
		l.Logger.Warning("server state is not 'running'")
		defer s.mux.Unlock()
		return
	}
	s.state = stopping
	s.mux.Unlock()

	s.listener.Close()
	s.wg.Wait() // wait for all active connection to close

	s.mux.Lock()
	s.state = stopped
	s.mux.Unlock()
}
