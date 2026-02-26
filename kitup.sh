#!/bin/bash

GO_VERSION="1.26.0"
BIN_PATH="/usr/local/bin/kit"

set -e

install_dir=/tmp/kitup-install-$(date +%s)
mkdir -p "${install_dir}"
# Register cleanup
trap 'rm -rf /tmp/kitup-install-*' EXIT

# Clone the repository to build
git clone --depth 1 https://github.com/PondWader/kit "$install_dir/kit"

case $(uname -m) in
    x86_64) GOARCH="amd64" ;;
    aarch64|arm64) GOARCH="arm64" ;;
    armv6l) GOARCH="armv6l" ;;
    i386|i686) GOARCH="386" ;;
    *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
esac

# Download Go
wget -O "$install_dir/go.tar.gz" "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz"
tar xf "$install_dir/go.tar.gz" -C "$install_dir"

# Build the execute
echo "Building..."
"$install_dir/go/bin/go" build -C "$install_dir/kit" -o "$BIN_PATH" -trimpath -ldflags "-s -w" ./cmd
echo "Built and wrote kit to $BIN_PATH"

