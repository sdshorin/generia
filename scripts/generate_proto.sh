#!/bin/bash
set -e

# Define directories
PROTO_DIR="api/proto"
OUT_DIR="api/grpc"

# Create output directory if it doesn't exist
mkdir -p $OUT_DIR

# Check for required tools
if ! command -v protoc &> /dev/null; then
  echo "Error: protoc is not installed. Please install Protocol Buffers compiler."
  exit 1
fi

if ! command -v protoc-gen-go &> /dev/null || ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo "Error: protoc-gen-go and/or protoc-gen-go-grpc plugins are not installed."
  echo "Install them with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
  exit 1
fi

# Generate code for each service directory
for service_dir in "$PROTO_DIR"/*; do
  if [ -d "$service_dir" ]; then
    service=$(basename "$service_dir")
    echo "Generating gRPC code for $service service..."
    
    # Create output directory for the service
    mkdir -p "$OUT_DIR/$service"
    
    # Generate Go code
    protoc --proto_path="$PROTO_DIR" \
      --go_out="$OUT_DIR" --go_opt=paths=source_relative \
      --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
      "$service_dir"/*.proto
    
    echo "Generated gRPC code for $service service"
  fi
done

echo "All gRPC code generated successfully!"
