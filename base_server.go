package rtmp

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	l "github.com/junli1026/gortmp/logging"
)

type serverImpl interface {
	newContext(conn net.Conn) interface{}
	read(data []byte, context interface{}) (int, []byte, error)
}

type connHandler struct {
	conn      net.Conn
	readbuf   []byte
	mux       *sync.Mutex
	s         *baseServer
	stopping  int32
	readstop  chan int
	writestop chan int
	outch     chan []byte
	context   interface{}
}

func newHandler(conn net.Conn, s *baseServer) *connHandler {
	handler := &connHandler{
		conn:      conn,
		readbuf:   make([]byte, 0),
		mux:       &sync.Mutex{},
		s:         s,
		stopping:  0,
		outch:     make(chan []byte, 16),
		readstop:  make(chan int, 1),
		writestop: make(chan int, 1),
		context:   s.impl.newContext(conn),
	}
	return handler
}

func (h *connHandler) close() {
	if !atomic.CompareAndSwapInt32(&h.stopping, 0, 1) {
		return
	}
	h.conn.Close()
	h.outch <- []byte{}
	<-h.readstop
	<-h.writestop
	h.s.removeHandler(h.conn)
}

func (h *connHandler) readLoop() {
	var buf = make([]byte, 1024*10)
	l.Logger.Info("connection handler starts working\n")
Loop:
	for {
		// if not data for 60 seconds, we close conn with error timeout
		if err := h.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			l.Logger.Error(err)
			break
		}
		length, err := h.conn.Read(buf)
		if err != nil {
			if atomic.LoadInt32(&h.stopping) != 0 {
				l.Logger.Info("trying to stop read loop")
			} else if err == io.EOF {
				go h.close()
			} else {
				l.Logger.Error("read conn failed ", err)
				go h.close()
			}
			break
		}

		h.readbuf = append(h.readbuf, buf[:length]...)

		for {
			length, reply, err := h.s.impl.read(h.readbuf, h.context)
			if err != nil {
				l.Logger.Error(err)
				break Loop
			}

			if reply != nil && len(reply) > 0 {
				h.outch <- reply
			}

			if length != 0 {
				h.readbuf = h.readbuf[length:]
			} else {
				break
			}
		}
	}
	h.readstop <- 1
}

func (h *connHandler) writeLoop() {
	for {
		reply := <-h.outch
		if reply == nil || len(reply) == 0 {
			break
		}

		for {
			length, err := h.conn.Write(reply)
			if err != nil {
				if atomic.LoadInt32(&h.stopping) != 0 {
					l.Logger.Info("trying to stop write loop")
				} else if err == io.EOF {
					l.Logger.Info("eof")
				} else {
					l.Logger.Error("write conn failed ", err)
				}
				break
			}

			reply = reply[length:]
			if len(reply) == 0 {
				break
			}
		}
	}
	h.writestop <- 1
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
	wg       sync.WaitGroup
	mux      sync.Mutex
	state    serverState
	handlers map[net.Conn]*connHandler
	stopch   chan int
	impl     serverImpl
}

func newBaseServer(addr string, impl serverImpl) *baseServer {
	return &baseServer{
		addr:     addr,
		listener: nil,
		state:    stopped,
		handlers: make(map[net.Conn]*connHandler),
		stopch:   make(chan int, 1),
		impl:     impl,
	}
}

func (s *baseServer) numConnections() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return len(s.handlers)
}

func (s *baseServer) addHandler(conn net.Conn) *connHandler {
	handler := newHandler(conn, s)
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.handlers[conn]; !ok {
		s.handlers[conn] = handler
		s.wg.Add(1)
	}
	l.Logger.Infof("%v active connections", len(s.handlers))
	return handler
}

func (s *baseServer) removeHandler(conn net.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.handlers[conn]; ok {
		delete(s.handlers, conn)
		s.wg.Done()
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

		handler := s.addHandler(conn)
		go handler.readLoop()
		go handler.writeLoop()
	}
	s.stopch <- 1
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

	// stop the listener, then stop all active connections
	s.listener.Close()
	for _, handler := range s.handlers {
		go handler.close()
	}
	s.wg.Wait()
	<-s.stopch

	s.mux.Lock()
	s.state = stopped
	s.mux.Unlock()
}
