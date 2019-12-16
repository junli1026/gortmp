package main

import (
	"bufio"
	"os"

	"github.com/junli1026/rtmp-server"
)

func main() {
	s := rtmp.NewServer(":1936")
	c := make(chan int)

	f, err := os.Create("./test.flv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writeFile := func(stream *rtmp.StreamMeta, timestamp uint32, data []byte) error {
		f.Write(data)
		return nil
	}
	s.OnFlvHeader(writeFile)
	s.OnFlvScriptData(writeFile)
	s.OnFlvAudioData(writeFile)
	s.OnFlvVideoData(writeFile)

	go func() {
		if err := s.Run(); err != nil {
			panic(err)
		}
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}

			if b == 'q' {
				s.Stop()
				c <- 0
			}
		}
	}()

	<-c
}
