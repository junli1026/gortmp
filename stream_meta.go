package rtmp

//StreamMeta describes stream metadata
type StreamMeta struct {
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

//GetStreamID returns message stream id
func (st *StreamMeta) GetStreamID() int {
	return st.streamID
}

//GetStreamName returns stream name
func (st *StreamMeta) GetStreamName() string {
	return st.streamName
}

//GetWidth returns video width
func (st *StreamMeta) GetWidth() int {
	return st.width
}

//GetHeight returns video Height
func (st *StreamMeta) GetHeight() int {
	return st.height
}

//GetFrameRate returns video frame rate
func (st *StreamMeta) GetFrameRate() int {
	return st.frameRate
}

//GetVideoCodec returns video codec fourcc
func (st *StreamMeta) GetVideoCodec() string {
	return st.videoCodec
}

//GetVideoDataRate returns video data rate
func (st *StreamMeta) GetVideoDataRate() int {
	return st.videoDataRate
}

//GetAudioCodec returns audio codec
func (st *StreamMeta) GetAudioCodec() string {
	return st.audioCodec
}

//GetAudioDataRate return audio data rate
func (st *StreamMeta) GetAudioDataRate() int {
	return st.audioDataRate
}

//GetAudioChannels returns number of audio channels
func (st *StreamMeta) GetAudioChannels() int {
	return st.audioChannels
}

//GetAudioSampleRate returns audio sample rate
func (st *StreamMeta) GetAudioSampleRate() int {
	return st.audioSampleRate
}

//GetAudioSampleSize returns audio sample size
func (st *StreamMeta) GetAudioSampleSize() int {
	return st.audioSampleSize
}

//IsStereo returns boolean indicating whether the audio is stereo
func (st *StreamMeta) IsStereo() bool {
	return st.stereo
}

//GetEncoder returns encoder name
func (st *StreamMeta) GetEncoder() string {
	return st.encoder
}
