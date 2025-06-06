package media

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type stream struct {
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  string `json:"duration"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type format struct {
	Filename       string `json:"filename"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	Duration       string `json:"duration"`
	FormatLongName string `json:"format_long_name"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// probeOutput represents the output of ffprobe in JSON format
type probeOutput struct {
	Streams []stream `json:"streams"`
	Format  format   `json:"format"`
}
