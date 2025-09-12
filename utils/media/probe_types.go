package media

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type stream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"` // "video" | "audio"
	CodecName string `json:"codec_name"`
	Profile   string `json:"profile"`
	BitRate   string `json:"bit_rate"`
	// video
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"` // e.g. "24000/1001"
	Duration     string `json:"duration"`
	// audio
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
	SampleRate    string `json:"sample_rate"`
	// selection helpers
	Tags        map[string]string `json:"tags"` // language, etc.
	Disposition struct {
		Default int `json:"default"`
	} `json:"disposition"`
}

type format struct {
	FormatName string `json:"format_name"`
	Filename   string `json:"filename"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	Duration   string `json:"duration"`
}

type probeOutput struct {
	Streams []stream `json:"streams"`
	Format  format   `json:"format"`
}
