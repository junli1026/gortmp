package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/junli1026/rtmp-server"
)

func main() {
	s := rtmp.NewServer(":1936")

	/* config log settings */
	s.ConfigLog(&rtmp.LogSetting{
		LogLevel:   rtmp.InfoLevel, //set loglevel, default value rtmp.DebugLevel
		Filename:   "./log.txt",    //set log file, default value empty, that is, logging to stderr
		MaxSize:    1,              //set log file size to 1 MB
		MaxBackups: 3,              //set maximum number of log files to 3
		MaxAge:     1,              //set maximum log file life to 1 day
	})

	c := make(chan int)

	f, _ := os.Create("./test.flv")
	defer f.Close()

	/* register callback when receiving flv header */
	s.OnFlvHeader(func(stream *rtmp.StreamMeta, timestamp uint32, data []byte) error {
		fmt.Printf("    encoder: %v\n", stream.Encoder())
		fmt.Printf(" stream url: %v\n", stream.URL())
		fmt.Printf("stream name: %v\n", stream.StreamName())
		fmt.Printf("video codec: %v\n", stream.VideoCodec())
		fmt.Printf(" frame rate: %v\n", stream.FrameRate())
		fmt.Printf("      width: %v\n", stream.Width())
		fmt.Printf("     height: %v\n", stream.Height())
		fmt.Printf("audio codec: %v\n", stream.AudioCodec())
		fmt.Printf("   channels: %v\n", stream.AudioChannels())
		fmt.Printf(" samplerate: %v\n", stream.AudioSampleRate())
		fmt.Printf(" samplesize: %v\n", stream.AudioSampleSize())
		fmt.Printf("     stereo: %v\n", stream.Stereo())
		f.Write(data)
		return nil
	})

	writeFile := func(stream *rtmp.StreamMeta, timestamp uint32, data []byte) error {
		f.Write(data)
		return nil
	}

	/* register callback when receiving script data */
	s.OnFlvScriptData(writeFile)

	/* register callback when receiving audio data */
	s.OnFlvAudioData(writeFile)

	/* register callback when receiving video data */
	s.OnFlvVideoData(writeFile)

	/* create a goroutine to run server */
	go func() {
		if err := s.Run(); err != nil {
			panic(err)
		}
	}()

	/* create a goroutine listening to stdin, exit when received "q" */
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}

			if b == 'q' {
				s.Stop() // stop the server
				c <- 0
			}
		}
	}()

	<-c
}
