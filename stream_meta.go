package rtmp

//StreamMeta describes stream metadata
type StreamMeta struct {
	url             string
	streamID        int
	streamName      string
	hasVideo        bool
	hasAudio        bool
	width           int
	height          int
	frameRate       int
	videoCodec      string
	videoDataRate   int
	audioCodec      string
	audioDataRate   int
	audioChannels   int
	audioSampleRate int
	audioSampleSize int
	stereo          bool
	encoder         string
}

//URL returns stream url
func (st *StreamMeta) URL() string {
	return st.url
}

//StreamID returns stream id
func (st *StreamMeta) StreamID() int {
	return st.streamID
}

//StreamName returns stream name
func (st *StreamMeta) StreamName() string {
	return st.streamName
}

//Width returns video width
func (st *StreamMeta) Width() int {
	return st.width
}

//Height returns video height
func (st *StreamMeta) Height() int {
	return st.height
}

//FrameRate returns video frame rate
func (st *StreamMeta) FrameRate() int {
	return st.frameRate
}

//VideoCodec returns video codec fourcc
func (st *StreamMeta) VideoCodec() string {
	return st.videoCodec
}

//VideoDataRate returns video data rate
func (st *StreamMeta) VideoDataRate() int {
	return st.videoDataRate
}

//AudioCodec returns audio codec
func (st *StreamMeta) AudioCodec() string {
	return st.audioCodec
}

//AudioDataRate return audio data rate
func (st *StreamMeta) AudioDataRate() int {
	return st.audioDataRate
}

//AudioChannels returns number of audio channels
func (st *StreamMeta) AudioChannels() int {
	return st.audioChannels
}

//AudioSampleRate returns audio sample rate
func (st *StreamMeta) AudioSampleRate() int {
	return st.audioSampleRate
}

//AudioSampleSize returns audio sample size
func (st *StreamMeta) AudioSampleSize() int {
	return st.audioSampleSize
}

//Stereo returns boolean indicating whether the audio is stereo
func (st *StreamMeta) Stereo() bool {
	return st.stereo
}

//Encoder returns encoder name
func (st *StreamMeta) Encoder() string {
	return st.encoder
}
