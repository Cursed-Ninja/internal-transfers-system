# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o internal-transfers-system cmd/server/main.go

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/internal-transfers-system .

# Copy config directory
COPY --from=builder /app/config ./config

# Set environment variable
ENV APP_ENV=development

# Expose port
EXPOSE 8080

# Run the application
CMD ["./internal-transfers-system"]