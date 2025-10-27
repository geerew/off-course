FROM node:22-alpine AS ui

WORKDIR /src/ui

# Copy package.json and lock file then install dependencies
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# COPY the ui and build
COPY ui/ ./
RUN pnpm run build

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

FROM golang:1.25-alpine AS backend

RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev

ENV CGO_ENABLED=1 

WORKDIR /src/

COPY . .
RUN rm -rf ui

COPY --from=ui /src/ui/build ./ui/build
COPY --from=ui /src/ui/embed.go ./ui/embed.go

RUN go mod download
RUN go build -o offcourse .

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    libva-utils \
    mesa-va-gallium

COPY  --from=backend /src/offcourse /usr/local/bin/offcourse

RUN chmod +x /usr/local/bin/offcourse

EXPOSE 80

VOLUME [ "/offcourse" ]

ENTRYPOINT [ "/usr/local/bin/offcourse", "serve", "--data-dir", "/offcourse", "--http", "0.0.0.0:80" ]