#!/usr/bin/env bash
# Preflight checks for Authelia development run

set -u

# Color codes for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track overall pass/fail
CHECKS_FAILED=0

# Enhanced logging functions
log_error() {
    echo -e "${RED}✗ Error:${NC} $1" >&2
    CHECKS_FAILED=1
}

log_warning() {
    echo -e "${YELLOW}⚠ Warning:${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

check_file() {
    local file=$1
    local description=${2:-"$file"}
    
    if [ ! -f "$file" ]; then
        log_error "Required file '$description' not found at: $file"
        return 1
    fi
    
    if [ ! -r "$file" ]; then
        log_error "File '$description' exists but is not readable"
        return 1
    fi
    
    log_success "Found $description"
    return 0
}

check_cmd() {
    local cmd=$1
    local description=${2:-"$cmd"}
    
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Required command '$description' not found in PATH"
        return 1
    fi
    
    log_success "Found $description"
    return 0
}

check_node_version() {
    if ! command -v node >/dev/null 2>&1; then
        log_error "'node' is not installed. Install Node.js 22+ or 20.19+"
        echo "  Install via: https://nodejs.org/ or use a version manager like nvm"
        return 1
    fi
    
    local ver
    ver=$(node -v 2>/dev/null | sed 's/^v//') || {
        log_error "Failed to get node version"
        return 1
    }
    
    local major minor patch
    major=$(echo "$ver" | cut -d. -f1)
    minor=$(echo "$ver" | cut -d. -f2)
    patch=$(echo "$ver" | cut -d. -f3)
    
    # Validate version components are numeric
    if ! [[ "$major" =~ ^[0-9]+$ ]] || ! [[ "$minor" =~ ^[0-9]+$ ]]; then
        log_error "Invalid node version format: v$ver"
        return 1
    fi
    
    # Check version requirements
    if [ "$major" -ge 22 ]; then
        log_success "Node.js v$ver (meets requirement: >=22.0.0)"
        return 0
    fi
    
    if [ "$major" -eq 20 ] && [ "$minor" -ge 19 ]; then
        log_success "Node.js v$ver (meets requirement: >=20.19.0)"
        return 0
    fi
    
    log_warning "Node.js v$ver detected. Vite requires Node >=20.19 or >=22.0"
    log_warning "Builds may fail. Consider upgrading Node.js"
    return 0
}

check_pnpm_version() {
    if ! command -v pnpm >/dev/null 2>&1; then
        log_error "'pnpm' is not installed"
        echo "  Install via: corepack enable && corepack prepare pnpm@latest --activate"
        echo "  Or: npm install -g pnpm"
        return 1
    fi
    
    local pnpm_ver
    pnpm_ver=$(pnpm -v 2>/dev/null) || {
        log_warning "Found pnpm but couldn't determine version"
        return 0
    }
    
    log_success "pnpm v$pnpm_ver"
    return 0
}

check_go_version() {
    if ! command -v go >/dev/null 2>&1; then
        return 1  # Silent fail, handled by caller
    fi
    
    local go_ver
    go_ver=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//') || return 1
    
    log_info "Go $go_ver available for building"
    return 0
}

check_web_directory() {
    if [ ! -d "web" ]; then
        log_error "Frontend directory 'web' not found"
        return 1
    fi
    
    if [ ! -f "web/package.json" ]; then
        log_error "web/package.json not found"
        return 1
    fi
    
    log_success "Frontend directory structure valid"
    return 0
}

check_node_modules() {
    if [ ! -d "web/node_modules" ]; then
        log_warning "node_modules not found in web directory"
        echo "  Run: cd web && pnpm install"
        return 0
    fi
    
    log_success "node_modules present"
    return 0
}

check_static_files() {
    # Check that the built frontend static files are present where the Go server expects them
    local public_dir="internal/server/public_html"
    local static_dir="$public_dir/static"

    if [ -d "$public_dir" ] && [ -f "$public_dir/index.html" ]; then
        if [ -d "$static_dir" ]; then
            log_success "Built static frontend files present in '$public_dir'"
            return 0
        else
            log_warning "Public directory found but '$static_dir' missing"
            echo "  You may need to build the frontend assets: cd web && pnpm install && pnpm run build"
            return 1
        fi
    fi

    log_warning "Built frontend files not found in '$public_dir'"
    echo "  Build frontend with: cd web && pnpm install && pnpm run build"
    echo "  Or use: ./build-local.sh or ./build-docker.sh to produce the files"
    return 1
}

build_authelia_if_missing() {
    # Check if binary exists and is executable
    if [ -x "./authelia" ]; then
        log_success "Authelia binary found and executable"
        return 0
    fi
    
    if [ -f "./authelia" ] && [ ! -x "./authelia" ]; then
        log_warning "Authelia binary exists but is not executable"
        read -r -p "Make it executable? [y/N] " makex
        if [[ "${makex,,}" =~ ^y(es)?$ ]]; then
            chmod +x ./authelia && log_success "Made authelia executable" || {
                log_error "Failed to make authelia executable"
                return 1
            }
            return 0
        fi
    fi
    
    log_warning "Authelia binary './authelia' not found or not executable"
    
    # Check if running non-interactively
    if [ ! -t 0 ]; then
        log_error "Running non-interactively and binary is missing"
        return 1
    fi
    
    read -r -p "Build the binary now? [y/N] " buildnow
    if ! [[ "${buildnow,,}" =~ ^y(es)?$ ]]; then
        log_error "Binary required to run. Please build or provide './authelia'"
        return 1
    fi
    
    # Check Go availability
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go toolchain not found"
        echo "  Install Go from: https://go.dev/dl/"
        echo "  Or provide a prebuilt './authelia' binary"
        return 1
    fi
    
    log_info "Building backend binary (this may take a few minutes)..."
    
    # Build with proper error handling
    if GOEXPERIMENT="nosynchashtriemap" CGO_ENABLED=1 GOMEMLIMIT=1GiB \
       go build -p 1 -tags dev -ldflags "-s -w" -trimpath -o authelia ./cmd/authelia; then
        log_success "Binary built successfully"
        chmod +x ./authelia
        return 0
    else
        log_error "Go build failed"
        echo "  Check that all dependencies are available"
        echo "  Try: go mod download && go mod verify"
        return 1
    fi
}

check_port_availability() {
    local port=${1:-9010}
    
    if command -v lsof >/dev/null 2>&1; then
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warning "Port $port is already in use"
            return 0
        fi
    elif command -v ss >/dev/null 2>&1; then
        if ss -ltn | grep -q ":$port "; then
            log_warning "Port $port is already in use"
            return 0
        fi
    elif command -v netstat >/dev/null 2>&1; then
        if netstat -ln | grep -q ":$port "; then
            log_warning "Port $port is already in use"
            return 0
        fi
    fi
    
    log_success "Port $port is available"
    return 0
}

run_preflight_checks() {
    echo -e "${BLUE}=========================================="
    echo "Authelia Development Preflight Checks"
    echo -e "==========================================${NC}"
    echo ""
    
    # Reset failure counter
    CHECKS_FAILED=0
    
    # Configuration files
    echo -e "${BLUE}Checking configuration files...${NC}"
    check_file "config.yml" "Authelia config" || true
    check_file "users_database.yml" "Users database" || true
    echo ""
    
    # Runtime environment
    echo -e "${BLUE}Checking runtime environment...${NC}"
    check_node_version || true
    check_pnpm_version || true
    check_go_version || true  # Optional, just for info
    echo ""
    
    # Project structure
    echo -e "${BLUE}Checking project structure...${NC}"
    check_web_directory || true
    check_node_modules || true
    # Static built frontend files
    check_static_files || true
    echo ""
    
    # Binary availability
    echo -e "${BLUE}Checking backend binary...${NC}"
    build_authelia_if_missing || true
    echo ""
    
    # Port availability
    echo -e "${BLUE}Checking port availability...${NC}"
    check_port_availability 9010 || true
    echo ""
    
    # Summary
    if [ $CHECKS_FAILED -eq 0 ]; then
        echo -e "${GREEN}=========================================="
        echo "✓ All preflight checks passed"
        echo -e "==========================================${NC}"
        echo ""
        echo "Now run ./run-dev-hotreload.sh to start the development environment.":
        echo ""
        

        return 0
    else
        echo -e "${RED}=========================================="
        echo "✗ Some preflight checks failed"
        echo -e "==========================================${NC}"
        echo "Please resolve the errors above before continuing."
        return 1
    fi
}

# Allow sourcing without side-effects; callers should call run_preflight_checks
# If this script is executed directly, run the checks and exit with an appropriate code
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    if run_preflight_checks; then
        exit 0
    else
        exit 1
    fi
fi