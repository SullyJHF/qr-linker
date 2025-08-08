# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application and CLI tools
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o qr-linker .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o adduser cmd/adduser/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o manageusers cmd/manageusers/main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S qr-linker && \
    adduser -u 1001 -S qr-linker -G qr-linker

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/qr-linker .
COPY --from=builder /app/adduser .
COPY --from=builder /app/manageusers .

# Create data directory for database
RUN mkdir -p /app/data && \
    chown -R qr-linker:qr-linker /app

# Switch to non-root user
USER qr-linker

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/login || exit 1

# Run the application
CMD ["./qr-linker"]