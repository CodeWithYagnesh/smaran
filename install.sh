#!/bin/sh
# didfix installer
# Usage: curl -fsSL https://raw.githubusercontent.com/CodeWithYagnesh/didfix/main/install.sh | sh

set -e

REPO="CodeWithYagnesh/didfix"
BINARY="didfix"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ---- detect OS ----
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "didfix: unsupported OS: $OS" >&2
    echo "Download manually from https://github.com/${REPO}/releases" >&2
    exit 1
    ;;
esac

# ---- detect architecture ----
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "didfix: unsupported architecture: $ARCH" >&2
    echo "Download manually from https://github.com/${REPO}/releases" >&2
    exit 1
    ;;
esac

# ---- resolve version ----
VERSION="${DIDFIX_VERSION:-}"
if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "didfix: could not determine latest version, falling back to v1.0.0" >&2
  VERSION="v1.0.0"
fi

ARCHIVE="didfix-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "didfix: installing ${VERSION} for ${OS}/${ARCH}"
echo "didfix: downloading ${URL}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -fsSL "$URL" -o "${TMP_DIR}/${ARCHIVE}"; then
  echo "didfix: download failed. Check that a release exists for ${OS}/${ARCH}:" >&2
  echo "  https://github.com/${REPO}/releases/tag/${VERSION}" >&2
  exit 1
fi

tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

if [ ! -w "$INSTALL_DIR" ]; then
  echo "didfix: installing to ${INSTALL_DIR} (requires sudo)"
  sudo install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "didfix: installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Test it:"
echo "  didfix --help"
