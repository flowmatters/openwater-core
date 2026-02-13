#!/bin/bash
# Mark for Release Script
# Tags the current commit with version string for publication

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "OpenWater Core - Mark for Release"
echo "=================================="
echo

cd "$CORE_DIR"

# Check for uncommitted changes
echo "Checking for uncommitted changes..."

# Check for staged changes
if ! git diff --cached --quiet 2>/dev/null; then
    echo -e "${YELLOW}WARNING: You have staged but uncommitted changes:${NC}"
    git diff --cached --name-status
    echo
fi

# Check for modified files
if ! git diff --quiet 2>/dev/null; then
    echo -e "${YELLOW}WARNING: You have modified files:${NC}"
    git diff --name-status
    echo
fi

# Check for untracked files
UNTRACKED=$(git ls-files --others --exclude-standard)
if [ -n "$UNTRACKED" ]; then
    echo -e "${YELLOW}WARNING: You have untracked files:${NC}"
    echo "$UNTRACKED"
    echo
fi

# If any warnings were shown, ask for confirmation
if ! git diff --cached --quiet 2>/dev/null || ! git diff --quiet 2>/dev/null || [ -n "$UNTRACKED" ]; then
    echo -e "${YELLOW}Uncommitted changes detected. The tag will reference the current commit,${NC}"
    echo -e "${YELLOW}not these changes. Continue anyway?${NC}"
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Get version from ow-sim
if [ ! -f "./ow-sim" ]; then
    echo -e "${RED}Error: ow-sim not found. Build the project first.${NC}"
    exit 1
fi

VERSION_OUTPUT=$(./ow-sim --version 2>/dev/null | head -1)
VERSION=$(echo "$VERSION_OUTPUT" | awk '{print $2}')

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not extract version from ow-sim${NC}"
    exit 1
fi

# Add 'v' prefix for git tag
TAG_NAME="v${VERSION}"

echo -e "Version: ${GREEN}${VERSION}${NC}"
echo -e "Tag name: ${GREEN}${TAG_NAME}${NC}"
echo

# Check if tag already exists
if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag '${TAG_NAME}' already exists${NC}"
    echo "Delete it first with: git tag -d ${TAG_NAME}"
    exit 1
fi

# Confirm tagging
read -p "Create and push tag '${TAG_NAME}'? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create annotated tag
COMMIT_SHA=$(git rev-parse HEAD)
echo "Creating tag for commit ${COMMIT_SHA:0:7}..."
git tag -a "$TAG_NAME" -m "Release $VERSION"

# Push tag to origin
echo "Pushing tag to origin..."
git push origin "$TAG_NAME"

echo
echo -e "${GREEN}✓ Tag '${TAG_NAME}' created and pushed${NC}"
echo
echo "The GitHub workflow will now build and publish this release."
echo "Monitor progress at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
