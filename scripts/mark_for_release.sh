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

# Check if tag already exists locally
TAG_EXISTS_LOCALLY=false
TAG_EXISTS_REMOTELY=false

if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    TAG_EXISTS_LOCALLY=true
    LOCAL_TAG_COMMIT=$(git rev-parse "$TAG_NAME")
    CURRENT_COMMIT=$(git rev-parse HEAD)

    echo -e "${YELLOW}WARNING: Tag '${TAG_NAME}' already exists locally${NC}"
    echo "Current commit: ${CURRENT_COMMIT:0:7}"
    echo "Tagged commit:  ${LOCAL_TAG_COMMIT:0:7}"
    echo

    if [ "$LOCAL_TAG_COMMIT" = "$CURRENT_COMMIT" ]; then
        echo -e "${YELLOW}Tag points to the current commit.${NC}"
    else
        echo -e "${YELLOW}Tag points to a different commit!${NC}"
    fi
fi

# Check if tag exists on remote
if git ls-remote --tags origin | grep -q "refs/tags/$TAG_NAME"; then
    TAG_EXISTS_REMOTELY=true
    echo -e "${YELLOW}WARNING: Tag '${TAG_NAME}' already exists on remote${NC}"
fi

# Handle existing tags
if [ "$TAG_EXISTS_LOCALLY" = true ] || [ "$TAG_EXISTS_REMOTELY" = true ]; then
    echo
    echo -e "${YELLOW}Do you want to replace the existing tag?${NC}"
    echo "This will:"
    if [ "$TAG_EXISTS_LOCALLY" = true ]; then
        echo "  - Delete local tag '${TAG_NAME}'"
    fi
    if [ "$TAG_EXISTS_REMOTELY" = true ]; then
        echo "  - Force push to replace remote tag '${TAG_NAME}'"
        echo -e "  ${RED}WARNING: This will trigger a new build and may affect existing releases!${NC}"
    fi
    echo
    read -p "Replace existing tag? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi

    # Delete local tag if it exists
    if [ "$TAG_EXISTS_LOCALLY" = true ]; then
        echo "Deleting local tag '${TAG_NAME}'..."
        git tag -d "$TAG_NAME"
    fi
fi

# Confirm tagging (only if not already confirmed above)
if [ "$TAG_EXISTS_LOCALLY" = false ] && [ "$TAG_EXISTS_REMOTELY" = false ]; then
    read -p "Create and push tag '${TAG_NAME}'? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Create annotated tag
COMMIT_SHA=$(git rev-parse HEAD)
echo "Creating tag for commit ${COMMIT_SHA:0:7}..."
git tag -a "$TAG_NAME" -m "Release $VERSION"

# Push tag to origin
echo "Pushing tag to origin..."
if [ "$TAG_EXISTS_REMOTELY" = true ]; then
    # Force push to replace remote tag
    git push --force origin "$TAG_NAME"
    echo -e "${YELLOW}Force pushed to replace remote tag${NC}"
else
    git push origin "$TAG_NAME"
fi

echo
echo -e "${GREEN}✓ Tag '${TAG_NAME}' created and pushed${NC}"
echo
echo "The GitHub workflow will now build and publish this release."
echo "Monitor progress at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
