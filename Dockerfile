# --- Stage 1: Build the binary ---
FROM golang:1.22 as builder

WORKDIR /app
COPY . .

# Optional: download go dependencies early for better caching
RUN go mod download

# Build statically-linked binary (except for sqlite, which uses cgo)
RUN CGO_ENABLED=1 go build -o gpu-tracker ./cmd/gpu-tracker

# --- Stage 2: Minimal runtime image ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libsqlite3-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# App files
WORKDIR /app
COPY --from=builder /app/gpu-tracker /usr/local/bin/gpu-tracker

# The app writes its sqlite db into the user's home by default
RUN useradd -ms /bin/bash gpuuser
USER gpuuser

ENTRYPOINT ["gpu-tracker"]
