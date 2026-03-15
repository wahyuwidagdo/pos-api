# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o pos-api ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /root/

# Install necessary runtime dependencies (e.g., certificates)
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/pos-api .

# Copy environment file (optional, but good for defaults)
COPY .env.example .env

# Expose port
EXPOSE 8080

# Command to run
CMD ["./pos-api"]
