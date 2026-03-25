# Stage 1: Build Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /logfindr ./cmd/logfindr

# Stage 2: Minimal runtime
FROM debian:bookworm-slim

# Install Fluent Bit from official repo
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl gnupg && \
    curl -fsSL https://packages.fluentbit.io/fluentbit.key | gpg --dearmor -o /usr/share/keyrings/fluentbit-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/fluentbit-keyring.gpg] https://packages.fluentbit.io/debian/bookworm bookworm main" \
        > /etc/apt/sources.list.d/fluent-bit.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends fluent-bit && \
    apt-get purge -y --auto-remove curl gnupg && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /logfindr /usr/local/bin/logfindr
COPY configs/fluent-bit/fluent-bit.conf /etc/fluent-bit/fluent-bit.conf
COPY configs/fluent-bit/parsers.conf /etc/fluent-bit/parsers.conf
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

VOLUME /data
EXPOSE 8080 24224

ENTRYPOINT ["/entrypoint.sh"]
