package rtmp

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

func sendc0(conn net.Conn) error {
	c0 := make([]byte, 1)
	c0[0] = 0x03 // version 3
	l, err := conn.Write(c0)
	if err != nil {
		return err
	}
	if l != 1 {
		return errors.New("sendc0 fail")
	}
	return nil
}

func sendc1(conn net.Conn) error {
	c1 := make([]byte, 1536)
	for len(c1) > 0 {
		l, err := conn.Write(c1)
		if err != nil {
			return err
		}
		c1 = c1[l:]
	}
	return nil
}

func Test_HandshakeC0C1(t *testing.T) {
	s := newRtmpServer()
	go s.listenAndServe(":1234")
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fail()
	}
	defer conn.Close()
	defer s.stop()

	if err = sendc0(conn); err != nil {
		t.Fail()
	}
	if err = sendc1(conn); err != nil {
		t.Fail()
	}

	buf := make([]byte, 1024*100)
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fail()
	}
	index := 0
	for {
		l, err := conn.Read(buf[index:])
		if err != nil {
			fmt.Println(err)
			break
		}
		index += l
	}
	if index != 1536*2+1 {
		t.Fail()
	}
}

func Test_HandshakeC0(t *testing.T) {
	s := newRtmpServer()
	go s.listenAndServe(":1234")
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fail()
	}
	defer conn.Close()
	defer s.stop()

	if err = sendc0(conn); err != nil {
		t.Fail()
	}

	buf := make([]byte, 1024*100)
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fail()
	}
	index := 0
	for {
		l, err := conn.Read(buf[index:])
		if err != nil {
			fmt.Println(err)
			break
		}
		index += l
	}
	if index != 1536+1 {
		t.Fail()
	}
}
