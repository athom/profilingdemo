# builder stage
FROM golang:1.14.4-stretch AS builder

# Install build dependencies
RUN apt-get -qq update && \
    apt-get -qq install -y --no-install-recommends \
      build-essential \
      git \
      openssh-client \
    && rm -rf /var/lib/apt/lists/*

# Update timezone
ENV TZ=Asia/Singapore
WORKDIR /app

## download and cache go dependencies
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o demo


# application stage
FROM debian:stretch-slim as application
WORKDIR /app

# Install runtime dependencies
RUN apt-get -qq update && \
    apt-get -qq install -y --no-install-recommends \
      curl \
    && rm -rf /var/lib/apt/lists/*

# Update timezone
ENV TZ=Asia/Singapore
ENV ROOT_DIR=/app

EXPOSE 12000

#HEALTHCHECK --start-period=10s \
#            --interval=15s \
#            --timeout=5s \
#            --retries=3 \
#            CMD curl -sSf http://localhost:12000 || exit 1

COPY --from=builder /app/demo ./demo
COPY --from=builder /app/data.json ./data.json

CMD ["./demo"]
