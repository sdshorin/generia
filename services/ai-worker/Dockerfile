FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy module files for dependency download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ai-worker ./services/ai-worker/cmd/main.go

# Create minimal runtime image
FROM alpine:3.16

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/ai-worker /app/ai-worker

# Set executable permissions
RUN chmod +x /app/ai-worker

# Run the service
CMD ["/app/ai-worker"]