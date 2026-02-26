#!/bin/bash

GO_VERSION="1.26.0"

set -e

install_dir=/tmp/kitup-install-$(date +%s)
mkdir -p "${install_dir}"

# Clone the repository to build
git clone --depth 1 https://github.com/PondWader/kit "$install_dir/kit"

if [[ $(uname -m) != "x86_64" ]]; then 
    echo "Unsupported architecture: $(uname -m)"
    exit 1
fi

# Download Go 
wget -O "$install_dir/go.tar.gz" https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz 
tar xf "$install_dir/go.tar.gz" -C "$install_dir"

# Build the execute
echo "Building..."
"$install_dir/go/bin/go" build -C "$install_dir/kit" -o "$(pwd)/kit" -trimpath -ldflags "-s -w" ./cmd
echo "Built!"

rm -rf /tmp/kitup-install-*