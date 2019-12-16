package rtmp

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func senddata(conn net.Conn, data []string) []string {
	buf := make([]byte, 128)
	replies := make([]string, 0)
	defer conn.Close()
	for _, s := range data {
		fmt.Fprintf(conn, s+"\n")
		length, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		replies = append(replies, string(buf[0:length]))
	}
	return replies
}

type echoServer struct {
	*baseServer
}

func newEchoServer(addr string) *echoServer {
	s := &echoServer{}
	s.baseServer = newBaseServer(addr, s)
	return s
}

func (*echoServer) newContext(con net.Conn) interface{} {
	return nil
}

func (*echoServer) read(arr []byte, opaque interface{}) (int, []byte, error) {
	for i, d := range arr {
		if d == '\n' {
			return i + 1, arr[0:i], nil
		}
	}
	return 0, nil, nil
}

func Test_Echo(t *testing.T) {
	s := newEchoServer("127.0.0.1:1234")
	go s.run()
	time.Sleep(1 * time.Second)

	num := 200
	conns := make([]net.Conn, 0)
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:1234")
		if err != nil {
			continue
		}
		conns = append(conns, conn)
	}
	for i := 0; i < len(conns); i++ {
		wg.Add(1)
		go func(c net.Conn) {
			replies := senddata(c, []string{"this", "is", "a", "test"})
			if len(replies) != 4 ||
				replies[0] != "this" ||
				replies[1] != "is" ||
				replies[2] != "a" ||
				replies[3] != "test" {
				t.Fail()
			}
			c.Close()
			wg.Done()
		}(conns[i])
	}
	wg.Wait()
	s.stop()
	if s.numConnections() != 0 {
		t.Fail()
	}
}
