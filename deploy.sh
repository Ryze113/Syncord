#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
SOURCE_DIR="${SOURCE_DIR:-$SCRIPT_DIR/src/Vencord}"
SYNCORD_DIR="${SYNCORD_DIR:-$SCRIPT_DIR/src/Installer}"
OUTPUT_DIR="${OUTPUT_DIR:-$SCRIPT_DIR/build}"
REPO_DIR="${REPO_DIR:-$SCRIPT_DIR}"
PUSH="${PUSH:-1}"
COMMIT_MESSAGE="${COMMIT_MESSAGE:-Build Syncord $(date -u +%Y-%m-%dT%H:%M:%SZ)}"

info() {
    printf '\n==> %s\n' "$1"
}

fail() {
    printf '\nERROR: %s\n' "$1" >&2
    exit 1
}

command -v node >/dev/null 2>&1 || fail "Node.js is required (Vencord needs Node 22 or newer)."
command -v pnpm >/dev/null 2>&1 || fail "pnpm is required."
command -v go >/dev/null 2>&1 || fail "Go is required."
command -v git >/dev/null 2>&1 || fail "git is required."

[[ -d "$SOURCE_DIR" ]] || fail "Source directory not found: $SOURCE_DIR"
[[ -d "$SYNCORD_DIR" ]] || fail "Syncord directory not found: $SYNCORD_DIR"
git -C "$REPO_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1 || fail "Git repository not found: $REPO_DIR"

info "Installing JavaScript dependencies"
(
    cd "$SOURCE_DIR"
    pnpm install --frozen-lockfile
)

info "Building Syncord/Vencord"
(
    cd "$SOURCE_DIR"
    pnpm buildStandalone
)

required_files=(
    patcher.js
    preload.js
    renderer.js
    renderer.css
)

for file in "${required_files[@]}"; do
    [[ -f "$SOURCE_DIR/dist/$file" ]] || fail "Build did not create Vencord-src/dist/$file"
done

info "Publishing built Discord files"
mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR"/{patcher.js,preload.js,renderer.js,renderer.css}
for file in "${required_files[@]}"; do
    cp -- "$SOURCE_DIR/dist/$file" "$OUTPUT_DIR/$file"
done

version="$(cd "$SOURCE_DIR" && node -p "require('./package.json').version")"
printf '%s\n' "$version" > "$OUTPUT_DIR/version.txt"

info "Building the Linux CLI installer"
(
    cd "$SYNCORD_DIR"
    make clean
    make PLATFORM=linux ARCH=amd64 GUI=0 VERSION="$version"
)
cp -- "$SYNCORD_DIR/build/VencordInstallerCli-linux" "$OUTPUT_DIR/"

info "Build complete"
printf 'Artifacts: %s\n' "$OUTPUT_DIR"
printf 'Version:   %s\n' "$version"

if [[ "$PUSH" != "1" ]]; then
    printf 'Push skipped (set PUSH=1 to commit and push).\n'
    exit 0
fi

info "Pushing build to GitHub"
(
    cd "$REPO_DIR"
    git rm -r --cached --ignore-unmatch src >/dev/null 2>&1 || true
    git add -A -- build settings deploy.sh README.md .gitignore
    if git diff --cached --quiet; then
        printf 'No changes to push.\n'
        exit 0
    fi
    git commit -m "$COMMIT_MESSAGE"
    git push origin main
)

printf '\nPublished version %s successfully.\n' "$version"
