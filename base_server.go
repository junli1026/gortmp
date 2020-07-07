package rtmp

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	l "github.com/junli1026/gortmp/logging"
)

type serverImpl interface {
	newContext(conn net.Conn) interface{}
	read(data []byte, context interface{}) (int, []byte, error)
	close(err error, context interface{})
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

func (h *connHandler) logIOError(err error) error {
	if h.s.serverState() == stopping { // server is in stopping state, supress the error log
		l.Logger.Warnf("server is in STOPPING state, %v", err)
		return err
	}
	if err != io.EOF {
		l.Logger.Error(err)
	}
	return err
}

func (h *connHandler) read() error {
	var buf = make([]byte, 1024*10)

	// if no data for 60 seconds, we close conn with error timeout
	if err := h.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return h.logIOError(err)
	}

	length, err := h.conn.Read(buf)
	if err != nil {
		return h.logIOError(err)
	}
	h.readbuf = append(h.readbuf, buf[:length]...)
	return nil
}

func (h *connHandler) writeAll(data []byte) error {
	for data != nil && len(data) > 0 {
		length, err := h.conn.Write(data)
		if err != nil {
			return h.logIOError(err)
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
			h.s.impl.close(err, h.context)
			return
		}

		for {
			length, reply, err := h.s.impl.read(h.readbuf, h.context)
			if err != nil {
				l.Logger.Errorf("application 'read' returns error: %v", err)
				return
			}

			if err = h.writeAll(reply); err != nil {
				h.s.impl.close(err, h.context)
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
	addr     *net.TCPAddr
	listener *net.TCPListener
	mux      sync.Mutex
	state    serverState
	impl     serverImpl
	wg       sync.WaitGroup
}

func newBaseServer(impl serverImpl) *baseServer {
	return &baseServer{
		listener: nil,
		state:    stopped,
		wg:       sync.WaitGroup{},
		impl:     impl,
	}
}

func (s *baseServer) listenAndServe(addr string) error {
	s.mux.Lock()
	if s.state != stopped {
		defer s.mux.Unlock()
		return errors.New("server is not in STOPPED state")
	}
	s.state = running
	s.mux.Unlock()

	var err error
	s.addr, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		l.Logger.Fatal(err)
	}

	s.listener, err = net.ListenTCP("tcp", s.addr)
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
			if err != io.EOF {
				l.Logger.Warning(err)
			}
			continue
		}
		defer conn.Close()
		l.Logger.Infof("new connection accepted from %v\n", conn.RemoteAddr().String())

		h := newHandler(conn, s)
		go h.run()
	}
	return nil
}

func (s *baseServer) serverState() serverState {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.state
}

func (s *baseServer) stop() {
	s.mux.Lock()
	l.Logger.Info("trying to stop the server...")
	if s.state != running {
		l.Logger.Warning("server state is not in RUNNING state")
		defer s.mux.Unlock()
		return
	}
	s.state = stopping
	s.mux.Unlock()

	s.listener.Close()
	s.wg.Wait() // wait for all active connection to close

	s.mux.Lock()
	s.state = stopped
	l.Logger.Info("server stopped")
	s.mux.Unlock()
}
