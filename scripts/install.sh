#!/bin/bash
# lufy-ai installer
# Usage: curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/main/scripts/install.sh | bash
# Or: /path/to/lufy-ai/scripts/install.sh [target-project-dir]

set -e

REPO_URL="https://github.com/adrotech/lufy-ai.git"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="${1:-.}"
TEMP_INSTALL_DIR=""

if [ -d "$SCRIPT_DIR/../.opencode" ] && [ -f "$SCRIPT_DIR/../AGENTS.md.template" ]; then
    INSTALL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
else
    TEMP_INSTALL_DIR="$(mktemp -d)"
    INSTALL_DIR="$TEMP_INSTALL_DIR/lufy-ai"
fi

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

cleanup() {
    if [ -n "$TEMP_INSTALL_DIR" ] && [ -d "$TEMP_INSTALL_DIR" ]; then
        rm -rf "$TEMP_INSTALL_DIR"
    fi
}

prepare_install_source() {
    if [ -d "$INSTALL_DIR/.opencode" ]; then
        return
    fi

    print_step "Fetching lufy-ai source..."
    git clone --depth 1 "$REPO_URL" "$INSTALL_DIR" >/dev/null 2>&1
    print_step "Source ready"
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
        echo "  1) frontend-react  (React)"
        echo "  2) frontend-nextjs (Next.js)"
        echo "  3) frontend-astro  (Astro)"
        echo "  4) mobile-expo     (Expo, React Native)"
        echo "  5) backend-spring  (Spring Boot, Java)"
        echo "  6) backend-python  (Python, Django, FastAPI)"
        echo ""
        read -p "Select stack [1-6] or press Enter to skip: " stack_choice
        
        case $stack_choice in
            1) stack="frontend-react" ;;
            2) stack="frontend-nextjs" ;;
            3) stack="frontend-astro" ;;
            4) stack="mobile-expo" ;;
            5) stack="backend-spring" ;;
            6) stack="backend-python" ;;
            *) print_info "Skipping stack detection" ;;
        esac
    fi
    
    DETECTED_STACK="$stack"
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

    # Copy local tooling metadata without installing dependencies
    for file in README.md package.json package-lock.json .gitignore; do
        if [ -f "$INSTALL_DIR/.opencode/$file" ]; then
            cp "$INSTALL_DIR/.opencode/$file" ".opencode/$file"
        fi
    done
    
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

copy_openspec() {
    print_step "Setting up OpenSpec structure..."

    if [ ! -d "openspec" ]; then
        if [ -d "$INSTALL_DIR/openspec" ]; then
            cp -r "$INSTALL_DIR/openspec" .
            print_step "Copied openspec/"
        else
            print_warn "No openspec/ template found"
        fi
    else
        print_info "openspec/ already exists, skipping"
    fi
}

write_opencode_config() {
    print_step "Setting up OpenCode configuration..."

    if [ -f "opencode.json" ]; then
        print_info "opencode.json already exists, skipping"
        return
    fi

    local project_name
    project_name="$(basename "$(pwd)")"

    cat > opencode.json <<EOF
{
  "\$schema": "https://opencode.ai/config.json",
  "default_agent": "orchestrator",
  "instructions": [
    "AGENTS.md"
  ],
  "plugin": [
    "./.opencode/plugins/anthropic-tool-streaming-compat.ts"
  ],
  "share": "manual",
  "watcher": {
    "ignore": [
      ".git/**",
      ".opencode/node_modules/**",
      "node_modules/**",
      "target/**",
      "dist/**",
      "build/**"
    ]
  },
  "permission": {
    "edit": "ask",
    "external_directory": "ask",
    "bash": {
      "*": "ask",
      "rg *": "allow",
      "git status*": "allow",
      "git diff*": "allow",
      "git log*": "allow"
    }
  },
  "mcp": {
    "engram": {
      "type": "local",
      "command": [
        "/opt/homebrew/bin/engram",
        "mcp",
        "--tools=agent",
        "--project",
        "$project_name"
      ],
      "enabled": false,
      "timeout": 3000
    }
  }
}
EOF

    print_step "Created opencode.json"
}

customize_observatory_id() {
    local project_name
    project_name="$(basename "$(pwd)")"

    if [ -f "tui.json" ]; then
        perl -0pi -e "s/lufy-ai\.observatory/${project_name}.observatory/g" tui.json
    fi

    if [ -f ".opencode/plugins/agent-observatory.tsx" ]; then
        perl -0pi -e "s/lufy-ai\.observatory/${project_name}.observatory/g" .opencode/plugins/agent-observatory.tsx
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
            if [ -f "opencode.json" ]; then
                perl -0pi -e 's/"enabled": false/"enabled": true/' opencode.json
            fi
            print_info "Engram MCP enabled in opencode.json"
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
    trap cleanup EXIT
    check_dependencies
    prepare_install_source
    
    cd "$TARGET_DIR" || exit 1
    
    detect_existing
    
    DETECTED_STACK=""
    detect_stack
    if [ -n "$DETECTED_STACK" ]; then
        print_info "Detected stack: $DETECTED_STACK"
    fi
    
    copy_files "$DETECTED_STACK"
    copy_agents_md
    copy_tui_config
    copy_openspec
    write_opencode_config
    customize_observatory_id
    check_engram
    print_next_steps
}

main "$@"
