#!/bin/bash

# Redis Analyzer - Build and Deploy Script
# ========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# Functions
print_header() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_dependencies() {
    print_header "Checking Dependencies"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_success "Go $(go version | cut -d' ' -f3)"
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        exit 1
    fi
    print_success "Node.js $(node --version)"
    
    # Check pnpm
    if ! command -v pnpm &> /dev/null; then
        print_error "pnpm is not installed"
        exit 1
    fi
    print_success "pnpm $(pnpm --version)"
    
    # Check Redis
    if ! command -v redis-server &> /dev/null; then
        print_warning "Redis server not found - install with: sudo apt install redis-server"
    else
        print_success "Redis server available"
    fi
    
    echo
}

clean() {
    print_header "Cleaning Build Artifacts"
    
    rm -rf "$BUILD_DIR"
    rm -rf "$DIST_DIR"
    rm -rf "$FRONTEND_DIR/dist"
    rm -rf "$FRONTEND_DIR/build"
    
    print_success "Cleaned build directories"
    echo
}

test_backend() {
    print_header "Testing Backend"
    
    cd "$BACKEND_DIR"
    
    # Run unit tests
    print_info "Running unit tests..."
    go test ./... -v
    
    print_success "Backend tests passed"
    echo
}

build_backend() {
    print_header "Building Backend"
    
    cd "$BACKEND_DIR"
    
    # Create build directory
    mkdir -p "$BUILD_DIR"
    
    # Build for current platform
    print_info "Building for current platform..."
    go build -o "$BUILD_DIR/redis-analyzer" .
    
    # Build for Linux (common deployment target)
    print_info "Building for Linux..."
    GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/redis-analyzer-linux" .
    
    # Build for Windows
    print_info "Building for Windows..."
    GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/redis-analyzer.exe" .
    
    # Build for macOS
    print_info "Building for macOS..."
    GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/redis-analyzer-macos" .
    
    print_success "Backend built successfully"
    echo
}

build_frontend() {
    print_header "Building Frontend"
    
    cd "$FRONTEND_DIR"
    
    # Install dependencies
    print_info "Installing dependencies..."
    pnpm install
    
    # Build for production
    print_info "Building for production..."
    pnpm run build
    
    # Copy build to main build directory
    mkdir -p "$BUILD_DIR/web"
    cp -r dist/* "$BUILD_DIR/web/"
    
    print_success "Frontend built successfully"
    echo
}

create_distribution() {
    print_header "Creating Distribution Package"
    
    mkdir -p "$DIST_DIR"
    
    # Create directory structure
    mkdir -p "$DIST_DIR/redis-analyzer"
    mkdir -p "$DIST_DIR/redis-analyzer/bin"
    mkdir -p "$DIST_DIR/redis-analyzer/web"
    mkdir -p "$DIST_DIR/redis-analyzer/docs"
    
    # Copy binaries
    cp "$BUILD_DIR"/redis-analyzer* "$DIST_DIR/redis-analyzer/bin/"
    
    # Copy web files
    cp -r "$BUILD_DIR/web"/* "$DIST_DIR/redis-analyzer/web/"
    
    # Copy documentation
    cp "$PROJECT_ROOT/README.md" "$DIST_DIR/redis-analyzer/docs/" 2>/dev/null || true
    
    # Create startup scripts
    cat > "$DIST_DIR/redis-analyzer/start.sh" << 'EOF'
#!/bin/bash
# Redis Analyzer Startup Script

echo "Starting Redis Analyzer..."

# Check if Redis is running
if ! pgrep redis-server > /dev/null; then
    echo "Warning: Redis server is not running"
    echo "Please start Redis with: sudo systemctl start redis-server"
    echo "Or: redis-server"
fi

# Determine the correct binary
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    BINARY="./bin/redis-analyzer-linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    BINARY="./bin/redis-analyzer-macos"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    BINARY="./bin/redis-analyzer.exe"
else
    BINARY="./bin/redis-analyzer"
fi

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "Error: Binary $BINARY not found"
    exit 1
fi

# Make binary executable
chmod +x "$BINARY"

# Start the server
echo "Starting server on http://localhost:8080"
echo "Press Ctrl+C to stop"
"$BINARY"
EOF

    cat > "$DIST_DIR/redis-analyzer/start.bat" << 'EOF'
@echo off
echo Starting Redis Analyzer...

REM Check if Redis is running (Windows)
tasklist /FI "IMAGENAME eq redis-server.exe" 2>NUL | find /I /N "redis-server.exe">NUL
if "%ERRORLEVEL%"=="1" (
    echo Warning: Redis server is not running
    echo Please start Redis server first
)

REM Start the server
echo Starting server on http://localhost:8080
echo Press Ctrl+C to stop
bin\redis-analyzer.exe
EOF

    chmod +x "$DIST_DIR/redis-analyzer/start.sh"
    
    # Create archive
    cd "$DIST_DIR"
    tar -czf "redis-analyzer-$(date +%Y%m%d).tar.gz" redis-analyzer/
    
    print_success "Distribution package created: $DIST_DIR/redis-analyzer-$(date +%Y%m%d).tar.gz"
    echo
}

run_integration_tests() {
    print_header "Running Integration Tests"
    
    cd "$PROJECT_ROOT"
    
    # Make sure Redis is running
    if ! pgrep redis-server > /dev/null; then
        print_warning "Starting Redis server..."
        sudo systemctl start redis-server || redis-server --daemonize yes
        sleep 2
    fi
    
    # Run integration tests
    go run integration.go
    
    print_success "Integration tests completed"
    echo
}

start_dev() {
    print_header "Starting Development Environment"
    
    # Check if Redis is running
    if ! pgrep redis-server > /dev/null; then
        print_info "Starting Redis server..."
        sudo systemctl start redis-server || redis-server --daemonize yes
        sleep 2
    fi
    
    print_info "Starting backend server..."
    cd "$BACKEND_DIR"
    go run main.go &
    BACKEND_PID=$!
    
    sleep 3
    
    print_info "Starting frontend server..."
    cd "$FRONTEND_DIR"
    pnpm run dev --host &
    FRONTEND_PID=$!
    
    print_success "Development environment started"
    print_info "Backend: http://localhost:8080"
    print_info "Frontend: http://localhost:5173"
    print_info "API Docs: http://localhost:8080/api/v1/health"
    
    echo
    print_warning "Press Ctrl+C to stop all servers"
    
    # Wait for interrupt
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT
    wait
}

deploy_production() {
    print_header "Production Deployment"
    
    # Build everything
    clean
    test_backend
    build_backend
    build_frontend
    create_distribution
    
    print_success "Production build completed"
    print_info "Distribution package: $DIST_DIR/redis-analyzer-$(date +%Y%m%d).tar.gz"
    print_info "Extract and run: ./start.sh (Linux/macOS) or start.bat (Windows)"
    echo
}

show_help() {
    echo -e "${CYAN}Redis Analyzer Build Script${NC}"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  clean          Clean build artifacts"
    echo "  deps           Check dependencies"
    echo "  test           Run backend tests"
    echo "  build          Build backend and frontend"
    echo "  build-backend  Build only backend"
    echo "  build-frontend Build only frontend"
    echo "  integration    Run integration tests"
    echo "  dev            Start development environment"
    echo "  deploy         Build for production deployment"
    echo "  help           Show this help message"
    echo
    echo "Examples:"
    echo "  $0 dev         # Start development servers"
    echo "  $0 deploy      # Create production build"
    echo "  $0 integration # Run full integration tests"
    echo
}

# Main script logic
case "${1:-help}" in
    clean)
        clean
        ;;
    deps)
        check_dependencies
        ;;
    test)
        test_backend
        ;;
    build)
        clean
        build_backend
        build_frontend
        ;;
    build-backend)
        build_backend
        ;;
    build-frontend)
        build_frontend
        ;;
    integration)
        run_integration_tests
        ;;
    dev)
        start_dev
        ;;
    deploy)
        deploy_production
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo
        show_help
        exit 1
        ;;
esac

