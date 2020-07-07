package main

import (
	"bufio"
	"fmt"
	"os"

	rtmp "github.com/junli1026/gortmp"
)

func main() {
	s := rtmp.NewServer()

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

	/* register handler for stream data */
	s.OnStreamData(func(meta *rtmp.StreamMeta, streamData *rtmp.StreamData) error {
		//header data only shows once, at the beginning of stream data
		if streamData.Type == rtmp.FlvHeader {
			fmt.Printf("    encoder: %v\n", meta.Encoder())
			fmt.Printf(" stream url: %v\n", meta.URL())
			fmt.Printf("stream name: %v\n", meta.StreamName())
			fmt.Printf("video codec: %v\n", meta.VideoCodec())
			fmt.Printf(" frame rate: %v\n", meta.FrameRate())
			fmt.Printf("      width: %v\n", meta.Width())
			fmt.Printf("     height: %v\n", meta.Height())
			fmt.Printf("audio codec: %v\n", meta.AudioCodec())
			fmt.Printf("   channels: %v\n", meta.AudioChannels())
			fmt.Printf(" samplerate: %v\n", meta.AudioSampleRate())
			fmt.Printf(" samplesize: %v\n", meta.AudioSampleSize())
			fmt.Printf("     stereo: %v\n", meta.Stereo())
		}

		// simply write binary to file
		f.Write(streamData.Data)
		return nil
	})

	s.OnStreamClose(func(meta *rtmp.StreamMeta, err error) {
		fmt.Printf("stream-'%v' name-'%v' closed for reason: %v\n",
			meta.URL(), meta.StreamName(), err)
	})

	go func() {
		if err := s.Run(":1936"); err != nil {
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
