#!/bin/bash
# lufy-ai installer
# Usage: curl -fsSL https://raw.githubusercontent.com/adrianrojas/lufy-ai/main/scripts/install.sh | bash
# Or: ./scripts/install.sh

set -e

REPO_URL="https://github.com/adrianrojas/lufy-ai.git"
INSTALL_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="${2:-.}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║         lufy-ai Installer             ║"
    echo "║   AI-First Software Development      ║"
    echo "╚════════════════════════════════════════╝${NC}"
}

print_step() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[i]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

check_dependencies() {
    print_step "Checking dependencies..."
    
    local missing=()
    
    if ! command -v git &> /dev/null; then
        missing+=("git")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        print_error "Missing dependencies: ${missing[*]}"
        echo "Please install them and try again."
        exit 1
    fi
    
    print_step "Dependencies OK"
}

detect_existing() {
    print_step "Checking existing configuration..."
    
    if [ -d ".opencode" ]; then
        print_warn "Found existing .opencode/ directory"
        echo "This project may already have lufy-ai configured."
        echo ""
        read -p "Continue anyway? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled"
            exit 0
        fi
    fi
    
    if [ -f "AGENTS.md" ]; then
        print_warn "Found existing AGENTS.md"
        read -p "Backup and continue? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled"
            exit 0
        fi
        cp AGENTS.md AGENTS.md.backup
    fi
    
    print_step "No conflicts detected"
}

detect_stack() {
    print_step "Detecting project stack..."
    
    local stack="unknown"
    
    # Check for package.json (npm/Node.js projects)
    if [ -f "package.json" ]; then
        if grep -q '"expo"' package.json 2>/dev/null; then
            stack="mobile-expo"
        elif grep -q '"react"' package.json 2>/dev/null; then
            stack="frontend-react"
        else
            stack="frontend-node"
        fi
    fi
    
    # Check for pom.xml (Java/Maven)
    if [ -f "pom.xml" ]; then
        stack="backend-spring"
    fi
    
    # Check for go.mod (Go)
    if [ -f "go.mod" ]; then
        stack="backend-go"
    fi
    
    # Check for requirements.txt (Python)
    if [ -f "requirements.txt" ]; then
        stack="backend-python"
    fi
    
    # If unknown, ask user
    if [ "$stack" = "unknown" ]; then
        print_warn "Could not auto-detect stack"
        echo "Available stacks:"
        echo "  1) frontend-react  (React, Next.js, Vue)"
        echo "  2) mobile-expo    (Expo, React Native)"
        echo "  3) backend-spring (Spring Boot, Java)"
        echo "  4) backend-node   (Node.js, Express, Nest)"
        echo "  5) backend-python (Python, Django, FastAPI)"
        echo ""
        read -p "Select stack [1-5] or press Enter to skip: " stack_choice
        
        case $stack_choice in
            1) stack="frontend-react" ;;
            2) stack="mobile-expo" ;;
            3) stack="backend-spring" ;;
            4) stack="backend-node" ;;
            5) stack="backend-python" ;;
            *) print_info "Skipping stack detection" ;;
        esac
    fi
    
    echo $stack
}

copy_files() {
    local stack="$1"
    print_step "Copying lufy-ai configuration files..."
    
    # Create .opencode directory
    mkdir -p .opencode
    
    # Copy agents
    mkdir -p .opencode/agents
    cp -r "$INSTALL_DIR/.opencode/agents/"* .opencode/agents/ 2>/dev/null || true
    
    # Copy skills
    mkdir -p .opencode/skills
    cp -r "$INSTALL_DIR/.opencode/skills/"* .opencode/skills/ 2>/dev/null || true
    
    # Copy policies
    mkdir -p .opencode/policies
    cp -r "$INSTALL_DIR/.opencode/policies/"* .opencode/policies/ 2>/dev/null || true
    
    # Copy commands
    mkdir -p .opencode/commands
    cp -r "$INSTALL_DIR/.opencode/commands/"* .opencode/commands/ 2>/dev/null || true
    
    # Copy plugins
    mkdir -p .opencode/plugins
    cp -r "$INSTALL_DIR/.opencode/plugins/"* .opencode/plugins/ 2>/dev/null || true
    
    # Copy agent-observatory
    mkdir -p .opencode/agent-observatory
    cp -r "$INSTALL_DIR/.opencode/agent-observatory/"* .opencode/agent-observatory/ 2>/dev/null || true
    
    # Copy templates
    mkdir -p .opencode/templates
    cp -r "$INSTALL_DIR/.opencode/templates/"* .opencode/templates/ 2>/dev/null || true
    
    print_step "Files copied"
}

copy_agents_md() {
    print_step "Setting up AGENTS.md..."
    
    if [ ! -f "AGENTS.md" ]; then
        if [ -f "$INSTALL_DIR/AGENTS.md.template" ]; then
            cp "$INSTALL_DIR/AGENTS.md.template" AGENTS.md
            print_step "Created AGENTS.md from template"
        else
            print_warn "No AGENTS.md.template found"
        fi
    else
        print_info "AGENTS.md already exists, skipping"
    fi
}

copy_tui_config() {
    print_step "Setting up TUI configuration..."
    
    if [ -f "$INSTALL_DIR/tui.json" ]; then
        cp "$INSTALL_DIR/tui.json" . 2>/dev/null || true
        print_step "Copied tui.json"
    else
        print_warn "No tui.json found, skipping"
    fi
}

check_engram() {
    print_step "Checking for Engram..."
    
    # Check if Engram is installed
    if command -v engram &> /dev/null || [ -d "$HOME/.engram" ]; then
        print_step "Engram found"
        read -p "Integrate Engram memory? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Engram integration: Add 'memory' skill to your workflow"
            echo "Use the memory skill for persistent sessions."
        fi
    else
        print_info "Engram not found (optional)"
        echo "For memory integration: https://github.com/Gentleman-Programming/gentle-ai"
    fi
}

print_next_steps() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║       Installation Complete!       ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review AGENTS.md for project conventions"
    echo "  2. Restart OpenCode to load new agents"
    echo "  3. Use /opsx-explore to start exploring"
    echo ""
    echo "Available commands:"
    echo "  /opsx-explore  - Explore codebase"
    echo "  /opsx-propose - Create proposal"
    echo "  /opsx-apply  - Implement tasks"
    echo "  /opsx-verify - Verify implementation"
    echo "  /opsx-archive - Archive completed"
    echo ""
    echo "Documentation: $INSTALL_DIR/docs/"
    echo ""
}

# Main execution
main() {
    print_header
    
    cd "$TARGET_DIR" || exit 1
    
    check_dependencies
    detect_existing
    
    local stack
    stack=$(detect_stack)
    if [ -n "$stack" ]; then
        print_info "Detected stack: $stack"
    fi
    
    copy_files "$stack"
    copy_agents_md
    copy_tui_config
    check_engram
    print_next_steps
}

main "$@"