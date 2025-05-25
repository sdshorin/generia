#!/bin/bash
set -e

# Define directories
ROOT_DIR=$(dirname $(dirname $0))
# Get absolute path to repo root
REPO_ROOT="$(cd "$ROOT_DIR/../.." && pwd)"
PROTO_DIR="$REPO_ROOT/api/proto"
GRPC_DIR="$ROOT_DIR/src/grpc"

echo "Generating gRPC Python code for AI Worker service"
echo "================================================="
echo "Proto files from: $PROTO_DIR"
echo "Output to: $GRPC_DIR"

# Create output directory if it doesn't exist
mkdir -p $GRPC_DIR

source $ROOT_DIR/.venv/bin/activate

# Check for required tools
if ! command -v python3 &> /dev/null; then
  echo "Error: Python 3 is not installed."
  exit 1
fi

if ! command -v pip3 &> /dev/null; then
  echo "Error: pip3 is not installed."
  exit 1
fi
which python3
# Install required packages if not already installed
python3 -m pip install grpcio grpcio-tools
PIP_PYTHON=$(which python3)

# Create empty __init__.py in the root grpc directory
echo "# gRPC generated code package" > "$GRPC_DIR/__init__.py"

# Generate code for each service directory we need
SERVICES=("character" "media" "post" "world")

for service in "${SERVICES[@]}"; do
  service_dir="$PROTO_DIR/$service"
  
  if [ -d "$service_dir" ]; then
    echo "Generating Python gRPC code for $service service..."
    
    # Create output directory for the service
    mkdir -p "$GRPC_DIR/$service"
    echo "# $service service gRPC generated code" > "$GRPC_DIR/$service/__init__.py"
    
    # Generate Python code
    $PIP_PYTHON -m grpc_tools.protoc \
      --proto_path="$PROTO_DIR" \
      --python_out="$GRPC_DIR" \
      --grpc_python_out="$GRPC_DIR" \
      "$service_dir"/*.proto
    
    echo "Generated Python gRPC code for $service service"
  else
    echo "Warning: Service directory not found: $service_dir"
  fi
done

echo "All Python gRPC code generated successfully!"
echo ""
echo "Next steps:"
echo "1. Verify the generated files in $GRPC_DIR"
echo "2. Add the files to git using 'git add $GRPC_DIR'"
echo "3. Commit the changes"
echo "4. Update the Dockerfile to remove runtime gRPC generation"