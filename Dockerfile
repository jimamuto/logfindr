# Stage 1: Build Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /logfindr ./cmd/logfindr

# Stage 2: Runtime with Fluent Bit
FROM fluent/fluent-bit:3.1

COPY --from=builder /logfindr /usr/local/bin/logfindr
COPY configs/fluent-bit/fluent-bit.conf /fluent-bit/etc/fluent-bit.conf
COPY configs/fluent-bit/parsers.conf /fluent-bit/etc/parsers.conf
COPY scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

VOLUME /data
EXPOSE 8080 24224

ENTRYPOINT ["/entrypoint.sh"]
