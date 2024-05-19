# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Set environment variables for the build
ARG PACKAGE="github.com/BullionBear/crypto-feed"
ARG VERSION
ARG COMMIT_HASH
ARG BUILD_TIMESTAMP
ARG LDFLAGS

# Build the binary
RUN VERSION=$(git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null) && \
    COMMIT_HASH=$(git rev-parse --short HEAD) && \
    BUILD_TIMESTAMP=$(date '+%Y-%m-%dT%H:%M:%S') && \
    LDFLAGS="-X '${PACKAGE}/internal/env.Version=${VERSION}' \
             -X '${PACKAGE}/internal/env.CommitHash=${COMMIT_HASH}' \
             -X '${PACKAGE}/internal/env.BuildTime=${BUILD_TIMESTAMP}'" && \
    env GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ./bin/cfeed-linux-x86 cmd/server/*.go

# Stage 2: Create the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/bin/cfeed-linux-x86 .

# Expose the port your application runs on
EXPOSE 50051

# Command to run the binary
CMD ["./cfeed-linux-x86", "--config", "./config/config.json"]
