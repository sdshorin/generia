#!/bin/bash

# Generia Static Server Launcher
# Simple script to start the development server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default port
PORT=${1:-8000}

# Function to print colored output
print_color() {
    echo -e "${1}${2}${NC}"
}

# Function to check if port is available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Function to find available port
find_available_port() {
    local port=$1
    while ! check_port $port; do
        print_color $YELLOW "⚠️  Port $port is already in use, trying $((port + 1))..."
        port=$((port + 1))
    done
    echo $port
}

# Change to script directory
cd "$(dirname "$0")"

print_color $BLUE "🌟 GENERIA STATIC SERVER LAUNCHER"
print_color $BLUE "=================================="

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    print_color $RED "❌ Python 3 is not installed or not in PATH"
    print_color $YELLOW "💡 Please install Python 3 to run the server"
    exit 1
fi

# Check if the port is available
AVAILABLE_PORT=$(find_available_port $PORT)

if [ "$AVAILABLE_PORT" != "$PORT" ]; then
    print_color $YELLOW "📍 Using port $AVAILABLE_PORT instead of $PORT"
    PORT=$AVAILABLE_PORT
fi

print_color $GREEN "🚀 Starting server on port $PORT..."
print_color $CYAN "📁 Serving files from: $(pwd)"
print_color $CYAN "🌐 Server will be available at: http://localhost:$PORT"
echo ""

# Check if server.py exists
if [ ! -f "server.py" ]; then
    print_color $RED "❌ server.py not found in current directory"
    print_color $YELLOW "💡 Make sure you're running this script from the correct directory"
    exit 1
fi

# Make server.py executable
chmod +x server.py

print_color $GREEN "✅ All checks passed!"
print_color $BLUE "🎯 Starting server now..."
echo ""

# Start the server
python3 server.py $PORT