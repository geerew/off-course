# Docker

## Build

Note: replace `x` with the appropriate tag

```bash
docker build --platform linux/amd64 -t geerew/offcourse:x -f docker/Dockerfile .
```

## Push

Note: replace `x` with the appropriate tag

```bash
docker push docker.io/geerew/offcourse:x
```

## Compose

The following is a `docker-compose.yaml`

```yaml
services:
  offcourse:
    container_name: OffCourse
    image: geerew/offcourse:x
    environment:
      - OC_ENABLE_SIGNUP=true
    restart: unless-stopped
    volumes:
      - /path/to/data:/offcourse # Path to persist the application data
      - /path/to/courses1:/courses1 # Path to directory containing courses
      - /path/to/courses2:/courses2 # Path to directory containing courses
    ports:
      - 9081:80
```

### Image Version

You can see the available versions [here](https://hub.docker.com/r/geerew/offcourse/tags).

### Volumes

A minimum of two volumes are required

- `xxx:/offcourse` - A persistent location on the host machine to store application data
- `xxx:/courses` - An optional location on the host machine where courses exist. `/courses` is an optional mount inside the container, however this can be called anything

Additional volumes can be mounted as needed

### Environment Variables

There are several environment variables that can be set

- OC_DEV - Whether to run in development mode. Defaults to `false`
- OC_ENABLE_SIGNUP - Whether to enable signup. Defaults to `false`

### Hardware Acceleration

OffCourse supports hardware-accelerated video transcoding for improved performance. The following environment variables control hardware acceleration:

- OC_HWACCEL - Hardware acceleration type. Options: `disabled` (default), `cpu`, `vaapi`, `qsv`, `intel`, `nvidia`
- OC_PRESET - FFmpeg encoding preset. Defaults to `fast`. Options: `ultrafast`, `superfast`, `veryfast`, `faster`, `fast`, `medium`, `slow`, `slower`, `veryslow`
- OC_VAAPI_RENDERER - VAAPI render device path. Defaults to `/dev/dri/renderD128`

#### Hardware Acceleration Types

- **disabled/cpu**: Software-only transcoding using CPU
- **vaapi**: Intel/AMD GPU acceleration via VAAPI (Linux)
- **qsv/intel**: Intel Quick Sync Video acceleration
- **nvidia**: NVIDIA GPU acceleration via CUDA/NVENC

#### Device Mounting

For hardware acceleration to work, you may need to mount GPU devices:

```yaml
services:
  offcourse:
    container_name: OffCourse
    image: geerew/offcourse:x
    environment:
      - OC_HWACCEL=nvidia
      - OC_ENABLE_SIGNUP=true
    restart: unless-stopped
    volumes:
      - /path/to/data:/offcourse
      - /path/to/courses:/courses
      - /dev/dri:/dev/dri # For Intel/AMD GPU (VAAPI)
    ports:
      - 9081:80
```
