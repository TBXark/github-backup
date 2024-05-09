#!/bin/bash

set -e
set -o pipefail

VERSION=0.0.1
BIN_NAME=github-backup
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
TARGET=${BIN_NAME}_${CURRENT_OS}_${CURRENT_ARCH}

URL=https://github.com/TBXark/${BIN_NAME}/releases/download/${VERSION}/${TARGET}.tar.gz
echo "Downloading ${BIN_NAME} from ${URL}"
curl -L $URL | tar xz
chmod +x ${TARGET}/github-backup
mv ${TARGET}/${BIN_NAME} /usr/local/bin/${BIN_NAME}
rm -rf ${TARGET}