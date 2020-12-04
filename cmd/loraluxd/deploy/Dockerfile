FROM golang:1.15-alpine AS builder

# Install git
RUN set -ex; \
    apk update; \
    apk add --no-cache git

# Copy Repository
WORKDIR /go/src/github.com/22arw/loralux/
COPY . ./

# Build Go Binary
RUN set -ex; \
    CGO_ENABLED=0 GOOS=linux go build -o ./loraluxd ./cmd/loraluxd;

# Final image, no source code
FROM alpine:latest

# Install Root Ceritifcates
RUN set -ex; \
    apk update; \
    apk add --no-cache \
     ca-certificates

WORKDIR /opt/
COPY --from=builder /go/src/github.com/22arw/loralux/loraluxd .

# Run Go Binary
CMD /opt/loraluxd -verbose