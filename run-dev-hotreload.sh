#!/bin/bash

# Development Quick Start Guide
# This script helps you run Authelia with frontend hot reload

set -e  # Exit on error (but we'll handle cleanup)

# Color codes for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global PIDs and cleanup state
BACKEND_PID=""
FRONTEND_PID=""
TAIL_PID=""
CLEANUP_DONE=0

# Comprehensive cleanup function
cleanup() {
    if [ $CLEANUP_DONE -eq 1 ]; then
        return
    fi
    CLEANUP_DONE=1

    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    
    # Kill tail process
    if [ -n "$TAIL_PID" ] && kill -0 "$TAIL_PID" 2>/dev/null; then
        kill "$TAIL_PID" 2>/dev/null || true
    fi
    
    # Kill frontend and its children
    if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        echo -e "${BLUE}Stopping frontend (PID: $FRONTEND_PID)...${NC}"
        # Kill process group to catch any child processes
        pkill -P "$FRONTEND_PID" 2>/dev/null || true
        kill "$FRONTEND_PID" 2>/dev/null || true
        sleep 1
        # Force kill if still running
        kill -9 "$FRONTEND_PID" 2>/dev/null || true
    fi
    
    # Kill backend and its children
    if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
        echo -e "${BLUE}Stopping backend (PID: $BACKEND_PID)...${NC}"
        # Kill process group to catch any child processes
        pkill -P "$BACKEND_PID" 2>/dev/null || true
        kill "$BACKEND_PID" 2>/dev/null || true
        sleep 1
        # Force kill if still running
        kill -9 "$BACKEND_PID" 2>/dev/null || true
    fi
    
    echo -e "${GREEN}Cleanup complete${NC}"
    exit 0
}

# Set up trap for various signals
trap cleanup INT TERM EXIT

# Function to check if a process is running
is_running() {
    local pid=$1
    kill -0 "$pid" 2>/dev/null
}

# Function to find listening ports for a PID
get_ports_for_pid() {
    local pid=$1
    local ports
    
    # Prefer lsof, fallback to ss, then netstat
    if command -v lsof >/dev/null 2>&1; then
        ports=$(lsof -Pan -p "$pid" -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR>1{split($9,a,":"); print a[length(a)]}' | sort -u | xargs)
    elif command -v ss >/dev/null 2>&1; then
        ports=$(ss -ltnp 2>/dev/null | grep -E "pid=$pid," | sed -E 's/.*:([0-9]+) .*/\1/' | sort -u | xargs)
    elif command -v netstat >/dev/null 2>&1; then
        ports=$(netstat -ltnp 2>/dev/null | grep "$pid" | sed -E 's/.*:([0-9]+) .*/\1/' | sort -u | xargs)
    else
        ports=""
    fi
    
    echo "$ports"
}

# Function to wait for process to listen on port
wait_for_port() {
    local pid=$1
    local timeout=${2:-10}
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        if ! is_running "$pid"; then
            echo -e "${RED}Process died unexpectedly${NC}"
            return 1
        fi
        
        local ports=$(get_ports_for_pid "$pid")
        if [ -n "$ports" ]; then
            echo "$ports"
            return 0
        fi
        
        sleep 1
        elapsed=$((elapsed + 1))
    done
    
    return 1
}

# Main execution
main() {
    echo -e "${GREEN}=========================================="
    echo "Authelia Development Environment"
    echo -e "==========================================${NC}"
    echo ""
    
    # Check if authelia binary exists
    if [ ! -f "./authelia" ]; then
        echo -e "${RED}Error: ./authelia binary not found${NC}"
        exit 1
    fi
    
    # Check if config exists
    if [ ! -f "config.yml" ]; then
        echo -e "${RED}Error: config.yml not found${NC}"
        exit 1
    fi
    
    # Start backend
    echo -e "${BLUE}Starting backend on port 9010...${NC}"
    ./authelia --config config.yml &
    BACKEND_PID=$!
    
    # Wait for backend to start
    echo "Backend PID: $BACKEND_PID"
    echo "Waiting for backend to listen on port..."
    
    if backend_ports=$(wait_for_port "$BACKEND_PID" 15); then
        echo -e "${GREEN}✓ Backend started successfully${NC}"
        echo "  Listening on ports: $backend_ports"
    else
        echo -e "${RED}✗ Backend failed to start${NC}"
        exit 1
    fi
    
    # Check if web directory exists
    if [ ! -d "web" ]; then
        echo -e "${RED}Error: web directory not found${NC}"
        exit 1
    fi
    
    # Start frontend
    echo ""
    echo -e "${BLUE}=========================================="
    echo "Starting frontend dev server..."
    echo -e "==========================================${NC}"
    echo ""
    echo -e "${GREEN}Hot reload is ENABLED${NC} - edit any .tsx/.ts file and see changes live!"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop both servers${NC}"
    echo ""
    
    # Start frontend and capture its output
    FRONTEND_LOG="$(pwd)/frontend.dev.log"
    cd web
    export PATH=/usr/bin:$PATH
    
    # Check if pnpm is available
    if ! command -v pnpm >/dev/null 2>&1; then
        echo -e "${RED}Error: pnpm not found${NC}"
        exit 1
    fi
    
    pnpm start > "$FRONTEND_LOG" 2>&1 &
    FRONTEND_PID=$!
    cd ..
    
    # Try to extract frontend URL from Vite output
    frontend_url=""
    echo "Waiting for frontend to start..."
    for i in $(seq 1 30); do
        if ! is_running "$FRONTEND_PID"; then
            echo -e "${RED}✗ Frontend process died${NC}"
            echo "Last 20 lines of frontend log:"
            tail -n 20 "$FRONTEND_LOG"
            exit 1
        fi
        
        if grep -E "Local:|Dev server running at" -m1 "$FRONTEND_LOG" >/dev/null 2>&1; then
            frontend_url=$(grep -oE "https?://[^[:space:]]+" "$FRONTEND_LOG" | head -n1)
            break
        fi
        
        # Fallback: check if the frontend process is listening
        pf=$(get_ports_for_pid "$FRONTEND_PID")
        if [ -n "$pf" ]; then
            frontend_url="http://localhost:$(echo "$pf" | awk '{print $1}')"
            break
        fi
        
        sleep 1
    done
    
    # Present running summary
    echo ""
    echo -e "${GREEN}================ Running Summary ================${NC}"
    echo -e "${BLUE}Backend:${NC}"
    echo "  PID: $BACKEND_PID"
    if [ -n "$backend_ports" ]; then
        echo "  Ports: $backend_ports"
    else
        echo "  Ports: (not detected)"
    fi
    echo ""
    echo -e "${BLUE}Frontend:${NC}"
    echo "  PID: $FRONTEND_PID"
    if [ -n "$frontend_url" ]; then
        echo -e "  ${GREEN}URL: $frontend_url${NC}"
    else
        echo "  URL: (starting up...)"
        echo "  Log: $FRONTEND_LOG"
    fi
    echo -e "${GREEN}================================================${NC}"
    echo ""
    
    # Tail the frontend log to the console
    tail -f "$FRONTEND_LOG" &
    TAIL_PID=$!
    
    # Monitor processes
    while true; do
        if ! is_running "$BACKEND_PID"; then
            echo -e "${RED}Backend process died!${NC}"
            exit 1
        fi
        
        if ! is_running "$FRONTEND_PID"; then
            echo -e "${RED}Frontend process died!${NC}"
            exit 1
        fi
        
        sleep 2
    done
}

# Run main function
main