## HLS package

This package provides on-demand HLS transcoding and segmentation with adaptive prefetching and hardware acceleration.

### Architecture

- **Transcoder**: Orchestrates HLS for an asset. Manages `StreamWrapper` instances and client tracking.
- **StreamWrapper**: Represents a single input file. Holds discovered `MediaInfo`, and per-track streams.
- **VideoStream / AudioStream**: Implement the `StreamHandle` to produce segments for either video variants or audio tracks.
- **Stream**: Core engine shared by audio/video. Manages heads (parallel encoders) and segments.
- **Tracker**: Observes client requests and updates heads accordingly.
- **Quality ladder**: Predefined qualities and helpers for bitrates/resolutions.
- **HW Accel**: Flags for decode/encode and scaling filters.

### Data flow

1. API requests `GetMaster`/`GetIndex`/`GetSegment` on `Transcoder`.
2. `Transcoder` creates/returns a `StreamWrapper` for the asset.
3. `StreamWrapper` probes metadata (`MediaInfo`) from DB-provided metadata.
4. For index/segment requests, `StreamWrapper` provides a `VideoStream` or `AudioStream`.
5. `Stream` schedules transcoding heads, invokes ffmpeg with segment times derived from keyframes, writes `.ts` files, and returns paths.

### Files

- `stream_wrapper.go`: `StreamWrapper`, `MediaInfo`, master playlist generation, audio/video stream accessors
- `stream_video.go`: `VideoStream` specifics and ffmpeg args for video
- `stream_audio.go`: `AudioStream` specifics and ffmpeg args for audio
- `stream.go`: Shared `Stream` engine, segment scheduling, head management
- `tracker.go`: Client tracking and heuristics
- `quality.go`: Quality ladder and bitrate calculations
- `hwaccel.go`: Hardware acceleration flags and filters
- `settings.go`: Package settings (cache path, defaults)
- `transcoder.go`: Entry point used by API layer to service HLS requests

### Notes

- Decorative separators are kept once between major sections
- Comments avoid trailing periods and remain concise
- All references to third-party implementations are removed from comments
