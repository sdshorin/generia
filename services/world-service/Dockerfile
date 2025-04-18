FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy module files for dependency download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/world-service ./services/world-service/cmd/main.go

# Create minimal runtime image
FROM alpine:3.16

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/world-service /app/world-service

# Set executable permissions
RUN chmod +x /app/world-service

# Expose port
EXPOSE 8085

# Run the service
CMD ["/app/world-service"]