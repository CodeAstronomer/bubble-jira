#!/bin/bash

# --- Configuration ---
# Stop the script on any error
set -euo pipefail

# App-specific variables
APP_NAME="jira"
BINARY_NAME="jira"
MAIN_GO="main.go"
PKG_ID="com.jira.cli"
STAGING_DIR="jira-pkg"

# ANSI Color Codes
GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[0;33m"
CYAN="\033[0;36m"
NC="\033[0m" # No Color

# --- 1. Parse Version Flag ---
VERSION=""
if [[ "$#" -ge 2 ]] && [[ "$1" == "--version" ]]; then
  VERSION="$2"
else
  echo -e "${RED}Error: Missing or invalid version flag.${NC}"
  echo "Usage: ./build-pkg.sh --version <version_number>"
  echo "Example: ./build-pkg.sh --version 1.0.0"
  exit 1
fi

echo -e "${CYAN}Starting build for ${APP_NAME} version ${VERSION}...${NC}"

# --- 2. Check Prerequisites ---
echo "Checking dependencies..."
# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: 'go' command not found. Please install Go.${NC}"
    exit 1
fi

# Check for pkgbuild (Xcode Command Line Tools)
if ! command -v pkgbuild &> /dev/null; then
    echo -e "${RED}Error: 'pkgbuild' command not found.${NC}"
    echo "Please install Xcode Command Line Tools by running:"
    echo "xcode-select --install"
    exit 1
fi

# Check for source files
if [ ! -f "$MAIN_GO" ] || [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Missing '$MAIN_GO' or 'go.mod'.${NC}"
    echo "Please run this script from your project's root directory."
    exit 1
fi

echo "Checking for embedded file paths..."
if [ ! -f "ui/LICENSE" ]; then
    echo -e "${RED}Error: Missing 'LICENSE' file in project root.${NC}"
    echo "The 'ui/commands.go' file expects to embed '../LICENSE'."
    exit 1
fi
if [ ! -d "ui/languages" ]; then
    echo -e "${RED}Error: Missing 'ui/languages' directory.${NC}"
    echo "The 'ui/variables.go' file expects to embed 'languages/*.json' from within the 'ui/' dir."
    exit 1
fi
echo -e "${GREEN}All dependencies found.${NC}"

# --- 3. Cleanup ---
echo "Cleaning up old build artifacts..."
rm -rf "$STAGING_DIR"
rm -f "${APP_NAME}-installer-v*.pkg"
rm -f "$BINARY_NAME" # Remove old binary

# --- 4. Compile Go Binary ---
echo -e "\n${CYAN}Step 1: Compiling Go binary for darwin/arm64...${NC}"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$BINARY_NAME" "$MAIN_GO"

if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${RED}Build failed! Binary '$BINARY_NAME' was not created.${NC}"
    exit 1
fi
echo -e "${GREEN}Go binary compiled successfully.${NC}"

# --- 5. Prepare Package Structure ---
echo -e "\n${CYAN}Step 2: Creating package structure...${NC}"
INSTALL_PATH="$STAGING_DIR/root/usr/local/bin"
SCRIPT_PATH="$STAGING_DIR/scripts" # Wird nicht mehr verwendet, aber schadet nicht

mkdir -p "$INSTALL_PATH"
# mkdir -p "$SCRIPT_PATH" # Nicht mehr nötig

mv "$BINARY_NAME" "$INSTALL_PATH/"
echo "Package structure created and binary moved."

# WICHTIG: Sicherstellen, dass die Binärdatei ausführbar ist
chmod 755 "$INSTALL_PATH/$BINARY_NAME"
echo "Binary permissions set to 755."

# --- 6. Build the Installer Package ---
echo -e "\n${CYAN}Step 3: Building the .pkg installer (without scripts)...${NC}"
FINAL_PKG_NAME="${APP_NAME}-installer-v${VERSION}.pkg"

pkgbuild \
  --root "$STAGING_DIR/root" \
  --install-location "/" \
  --identifier "$PKG_ID" \
  --version "$VERSION" \
  "$FINAL_PKG_NAME"

echo -e "\n${GREEN}Success!${NC}"
echo -e "Installer created: ${GREEN}${FINAL_PKG_NAME}${NC}"

# --- 7. Final Instructions ---
echo -e "\n--- Next Steps ---"
echo "To test the installer:"
echo "1. Double-click '${FINAL_PKG_NAME}' to install."
echo "2. Open a NEW terminal window."
echo "3. Type 'bubble-jira -h' to test (the binary name)."