# Stage 1: Build Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /logfindr ./cmd/logfindr
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Stage 2: Runtime with Fluent Bit
FROM fluent/fluent-bit:3.1-debug

COPY --from=builder /logfindr /usr/local/bin/logfindr
COPY --from=builder /entrypoint.sh /entrypoint.sh
COPY configs/fluent-bit/fluent-bit.conf /fluent-bit/etc/fluent-bit.conf
COPY configs/fluent-bit/parsers.conf /fluent-bit/etc/parsers.conf

VOLUME /data
EXPOSE 8080 24224

ENTRYPOINT ["/entrypoint.sh"]
