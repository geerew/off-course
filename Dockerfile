FROM node:18-alpine AS ui

WORKDIR /src/ui

# Copy package.json and lock file then install dependencies
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# COPY the ui and build
COPY ui/ ./
RUN pnpm run build

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

FROM golang:1.23-alpine AS backend

WORKDIR /src/

COPY . .
RUN rm -rf ui

COPY --from=ui /src/ui/build ./ui/build
COPY --from=ui /src/ui/embed.go ./ui/embed.go

RUN echo 1
RUN ls -l 

RUN go mod download
RUN go build -o offcourse .

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY  --from=backend /src/offcourse /usr/local/bin/offcourse

EXPOSE 9081

ENTRYPOINT [ "/usr/local/bin/offcourse", "serve" ]