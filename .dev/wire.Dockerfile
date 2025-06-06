# Dockerfile for wire development container
# Simple approach that relies on volume mount for source code
FROM golang:1.24.4-bullseye

# Set up Go environment
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

# Create workspace directory
WORKDIR /workspace

# Default command
CMD ["bash"]