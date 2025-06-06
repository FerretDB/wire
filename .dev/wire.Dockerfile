# Dockerfile for wire development container
FROM golang:1.24.4-bullseye

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set up Go environment
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

# Create workspace directory
WORKDIR /workspace

# Copy and cache Go modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy and build tools first for better caching
COPY tools/go.mod tools/go.sum ./tools/
RUN cd tools && go mod download

COPY tools/tools.go ./tools/
RUN cd tools && go generate -x

# Default command
CMD ["bash"]