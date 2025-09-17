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
      - /path/to/courses:/courses # Path to additional directories containing courses
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
