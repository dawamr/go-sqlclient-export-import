FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/sqlclient-export-import ./cmd/app

# Create a smaller final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    postgresql-client \
    mysql-client \
    mesa-dev \
    xorg-server-dev \
    bash \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/sqlclient-export-import .

# Copy required directories
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/static ./static

# Create necessary directories
RUN mkdir -p ./exports ./uploads

# Expose the application port
EXPOSE 3000

# Command to run the application
CMD ["./sqlclient-export-import"] 