FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk update && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o post-service ./services/post-service/cmd

# Final stage
FROM alpine:latest

# Install required packages
RUN apk --no-cache add ca-certificates tzdata

# Create a non-privileged user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/post-service .

# Switch to non-privileged user
USER appuser

# Expose the service port
EXPOSE 8082

# Run the service
CMD ["./post-service"]