FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk update && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy only necessary files
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY configs/ ./configs/
COPY services/cache-service/ ./services/cache-service/
COPY .env* ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cache-service ./services/cache-service/cmd

# Final stage
FROM alpine:latest

# Install required packages
RUN apk --no-cache add ca-certificates tzdata

# Create a non-privileged user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/cache-service .

# Switch to non-privileged user
USER appuser

# Expose the service port
EXPOSE 8086

# Run the service
CMD ["./cache-service"]