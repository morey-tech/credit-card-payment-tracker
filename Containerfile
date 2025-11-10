# Multi-stage build for minimal final image

# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Version argument (defaults to "dev" if not provided)
ARG VERSION=dev

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" to reduce binary size
# -X main.version=${VERSION} to inject version at build time
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION}" -o server cmd/server/main.go

# Stage 2: Create minimal runtime image
FROM registry.access.redhat.com/ubi10-minimal:10.0

# Set working directory
WORKDIR /app
RUN chown 1000:0 /app

# Set user to 1000
USER 1000

# Copy the binary from builder
COPY --from=builder /app/server .

# Set environment variable to load sample data for dev builds
ARG VERSION=dev
ENV LOAD_SAMPLE_DATA=${VERSION:+true}

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./server"]
