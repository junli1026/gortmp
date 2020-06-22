package rtmp

import (
	"errors"
	l "github.com/junli1026/gortmp/logging"
)

type handshakeState struct {
	c0 bool
	c1 bool
	c2 bool
}

func newHandshakeState() *handshakeState {
	return &handshakeState{
		c0: false,
		c1: false,
		c2: false,
	}
}

func (hs *handshakeState) generateS0S1S2(c1data []byte) (s0s1s2 [1 + 1536 + 1536]byte) {
	s0s1s2[0] = 3
	copy(s0s1s2[1:1536+1], c1data[:])
	s0s1s2[1] = 1
	s0s1s2[2] = 0
	s0s1s2[3] = 2
	s0s1s2[4] = 6
	copy(s0s1s2[1+1536:], c1data[:])
	return
}

func (hs *handshakeState) generateS0S1() (s0s1 [1537]byte) {
	s0s1[0] = 3
	s0s1[1] = 1
	s0s1[2] = 0
	s0s1[3] = 2
	s0s1[4] = 6
	return
}

func (hs *handshakeState) generateS2(c1data []byte) (s2 [1536]byte) {
	if c1data[0] != 1 || c1data[1] != 0 || c1data[2] != 2 || c1data[3] != 6 {
		l.Logger.Error("client does not hornor s1 timetamp")
	}
	copy(s2[0:1536], c1data[:])
	return
}

func (hs *handshakeState) checkC2(c2data []byte) error {
	if c2data[0] != 1 || c2data[1] != 0 || c2data[2] != 2 || c2data[3] != 6 {
		return errors.New("client does not hornor s1 timetamp")
	}
	return nil
}

func (hs *handshakeState) done() bool {
	return hs.c2
}

func (hs *handshakeState) handshake(data []byte) (int, []byte, error) {
	if !hs.c0 {
		hs.c0 = true
		version := int(data[0])
		l.Logger.Infof("version %v", version)
		if len(data) >= 1536+1 {
			hs.c1 = true
			reply := hs.generateS0S1S2(data[1 : 1+1536])
			return 1537, reply[:], nil
		}
		reply := hs.generateS0S1()
		return 1, reply[:], nil
	}
	if !hs.c1 {
		if len(data) < 1536 {
			return 0, nil, nil
		}
		hs.c1 = true
		reply := hs.generateS2(data[0:1536])
		return 1536, reply[:], nil
	}
	if !hs.c2 {
		if len(data) < 1536 {
			return 0, nil, nil
		}
		if err := hs.checkC2(data); err != nil {
			return 0, nil, err
		}
		hs.c2 = true
		l.Logger.Info("handshake done")
		return 1536, nil, nil
	}
	return 0, nil, nil
}
