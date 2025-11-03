#!/bin/bash

# --- Configuration ---
# Stop the script on any error
set -euo pipefail

# App-specific variables
APP_NAME="bubble-jira"
BINARY_NAME="bubble-jira"
MAIN_GO="main.go"
PKG_ID="com.bubblejira.cli"
STAGING_DIR="bubble-jira-pkg"

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

echo "Cleaning up old build artifacts..."
rm -rf "$STAGING_DIR"
rm -f "${APP_NAME}-installer-v*.pkg"
rm -f "$BINARY_NAME" # Remove old binary

# --- 4. Compile Go Binary (Step 2 from tutorial) ---
echo -e "\n${CYAN}Step 1: Compiling Go binary for darwin/arm64...${NC}"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$BINARY_NAME" "$MAIN_GO"

if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${RED}Build failed! Binary '$BINARY_NAME' was not created.${NC}"
    exit 1
fi
echo -e "${GREEN}Go binary compiled successfully.${NC}"

# --- 5. Prepare Package Structure (Step 3) ---
echo -e "\n${CYAN}Step 2: Creating package structure...${NC}"
INSTALL_PATH="$STAGING_DIR/root/usr/local/bin"
SCRIPT_PATH="$STAGING_DIR/scripts"

mkdir -p "$INSTALL_PATH"
mkdir -p "$SCRIPT_PATH"

mv "$BINARY_NAME" "$INSTALL_PATH/"
echo "Package structure created and binary moved."

# --- 6. Create Post-Installation Script (Step 4) ---
echo -e "\n${CYAN}Step 3: Creating postinstall script...${NC}"
POSTINSTALL_SCRIPT="$SCRIPT_PATH/postinstall"

# Use 'cat << EOF' to write the script. The quotes around 'EOF' are
# critical to prevent this build script from expanding variables like $SUDO_USER.
cat > "$POSTINSTALL_SCRIPT" << 'EOF'
#!/bin/bash

# Define the function to add
# This is the corrected version that calls the compiled binary
ALIAS_FUNCTION="

# === bubble-jira alias ===
function jira() {
  # Calls the binary located in /usr/local/bin/bubble-jira
  bubble-jira "$@"
}
# === end bubble-jira alias ===
"

# This function checks if the alias exists and adds it if not
add_alias_to_file() {
  local file_path="$1"
  local file_name=$(basename "$file_path")

  # Get the home directory of the user who ran the installer
  # SUDO_USER is set by 'sudo' which is how installers run
  local home_dir=$(eval echo ~$SUDO_USER)
  local full_path="$home_dir/$file_name"

  # Use the direct path from the argument if we're not using $SUDO_USER
  # Fallback just in case SUDO_USER is not set
  if [ -z "$SUDO_USER" ] || [ -z "$home_dir" ]; then
      home_dir=$(eval echo ~)
      full_path="$home_dir/$file_name"
  fi

  echo "Checking for $full_path..."

  # Create the file if it doesn't exist
  if [ ! -f "$full_path" ]; then
      echo "$full_path not found, creating it."
      touch "$full_path"
      # Ensure the user owns the file
      chown $SUDO_USER "$full_path"
  fi

  # Check if the alias is already in the file
  if ! grep -q "# === bubble-jira alias ===" "$full_path"; then
    echo "Adding jira alias to $full_path..."
    # Append the function
    echo "$ALIAS_FUNCTION" >> "$full_path"
    echo "Alias added. User must open a new terminal for changes to take effect."
  else
    echo "jira alias already exists in $full_path. No action taken."
  fi
}

# Add to .zshrc (default for modern macOS)
add_alias_to_file ".zshrc"

# Add to .bash_profile (for users who use bash)
add_alias_to_file ".bash_profile"

exit 0
EOF

# Make the postinstall script executable
chmod +x "$POSTINSTALL_SCRIPT"
echo "Postinstall script created and made executable."

# --- 7. Build the Installer Package (Step 5) ---
echo -e "\n${CYAN}Step 4: Building the .pkg installer...${NC}"
FINAL_PKG_NAME="${APP_NAME}-installer-v${VERSION}.pkg"

pkgbuild \
  --root "$STAGING_DIR/root" \
  --scripts "$STAGING_DIR/scripts" \
  --install-location "/" \
  --identifier "$PKG_ID" \
  --version "$VERSION" \
  "$FINAL_PKG_NAME"

echo -e "\n${GREEN}Success!${NC}"
echo -e "Installer created: ${GREEN}${FINAL_PKG_NAME}${NC}"

# --- 8. Final Instructions (Step 6) ---
echo -e "\n--- Next Steps ---"
echo "To test the installer:"
echo "1. Double-click '${FINAL_PKG_NAME}' to install."
echo "2. Open a NEW terminal window."
echo "3. Type 'jira -h' to test."

echo -e "\nTo uninstall:"
echo "1. sudo rm /usr/local/bin/$BINARY_NAME"
echo "2. sudo pkgutil --forget $PKG_ID"
echo "3. Manually remove the 'jira function' block from your ~/.zshrc and ~/.bash_profile"