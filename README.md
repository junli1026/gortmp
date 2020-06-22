# Simple RTMP Server
A lightweight rtmp/1.0 ingestion server implementation.

## Compatibility
Support OBS, Wirecast and FFmpeg ingestion.

## Installation
```
go get "github.com/junli1026/gortmp"
```

## Example
The following example simply dump binary to flv file.
```go
s := rtmp.NewServer(":1936")

/* config log settings */
s.ConfigLog(&rtmp.LogSetting{
    LogLevel:   rtmp.InfoLevel, //set loglevel, default value rtmp.DebugLevel
    Filename:   "./log.txt",    //set log file, default value empty, that is, logging to stderr
    MaxSize:    1,              //set log file size to 1 MB
    MaxBackups: 3,              //set maximum number of log files to 3
    MaxAge:     1,              //set maximum log file life to 1 day
})

/* try dump stream to flv file */
f, _ := os.Create("./test.flv")
defer f.Close()

/* simply write binary to disk */
writeFile := func(stream *rtmp.StreamMeta, timestamp uint32, data []byte) error {
    f.Write(data)
    return nil
}

/* register callback when receiving flv header */
s.OnFlvHeader(writeFile)

/* register callback when receiving script data */
s.OnFlvScriptData(writeFile)

/* register callback when receiving audio data */
s.OnFlvAudioData(writeFile)

/* register callback when receiving video data */
s.OnFlvVideoData(writeFile)

go s.Run()

```
