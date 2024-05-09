#!/bin/bash

set -e
set -o pipefail

VERSION=0.0.1
BIN_NAME=github-backup
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
BUILD_FROM_SOURCE="You can build from source: go install github.com/TBXark/${BIN_NAME}@latest"

# if CURRENT_ARCH is x86_64 or amd64, then use x86
if [ "$CURRENT_ARCH" == "x86_64" ] || [ "$CURRENT_ARCH" == "amd64" ]; then
  CURRENT_ARCH="x86"
fi

# Support arch: x86, arm64
if [ "$CURRENT_ARCH" != "x86" ] && [ "$CURRENT_ARCH" != "arm64" ]; then
  echo "Unsupported arch: $CURRENT_ARCH"
  echo $BUILD_FROM_SOURCE
  exit 1
fi

# Support os: linux, darwin, windows
if [ "$CURRENT_OS" != "linux" ] && [ "$CURRENT_OS" != "darwin" ] && [ "$CURRENT_OS" != "windows" ]; then
  echo "Unsupported OS: $CURRENT_OS"
  echo $BUILD_FROM_SOURCE
  exit 1
fi

TARGET=${BIN_NAME}_${CURRENT_OS}_${CURRENT_ARCH}

URL=https://github.com/TBXark/${BIN_NAME}/releases/download/${VERSION}/${TARGET}.tar.gz
echo "Downloading ${BIN_NAME} from ${URL}"

TEMP_DIR=$(mktemp -d) || exit 1
readonly TEMP_DIR
trap 'rm -rf ${TEMP_DIR}' EXIT

curl -L $URL | tar -xz -C ${TEMP_DIR}
mv ${TEMP_DIR}/${TARGET}/${BIN_NAME} /usr/local/bin/${BIN_NAME}
chmod +x /usr/local/bin/${BIN_NAME}