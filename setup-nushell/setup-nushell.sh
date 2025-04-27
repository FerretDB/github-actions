#!/bin/bash

set -eu

VERSION=0.103.0
AMD64_CHECKSUM=4e717dbdebd95579a842fa832f73cc197d6ae3eb1b9bb47649a1d0b0ab2f9c87
ARM64_CHECKSUM=df0ef0cd87556997346936d65e7d1962675a13bd605d9a5920622b3f433b3dab

for cmd in curl tar sha256sum uname; do
    if ! command -v "${cmd}" &> /dev/null; then
        echo "Error: ${cmd} is not installed. Please install it first."
        exit 1
    fi
done

OS=$(uname -s)
if [ "${OS}" != "Linux" ]; then
    echo "Error: This script only supports Linux."
    exit 1
fi

ARCH=$(uname -m)
NAME="nu-${VERSION}-${ARCH}-unknown-linux-musl"

curl -fLO "https://github.com/nushell/nushell/releases/download/${VERSION}/${NAME}.tar.gz"

EXPECTED_CHECKSUM=""
case "${ARCH}" in
    x86_64)
        EXPECTED_CHECKSUM="${AMD64_CHECKSUM}"
        ;;
    aarch64)
        EXPECTED_CHECKSUM="${ARM64_CHECKSUM}"
        ;;
    *)
        echo "Error: Checksum not defined for architecture: ${ARCH}."
        exit 1
        ;;
esac

echo "${EXPECTED_CHECKSUM} ${NAME}.tar.gz" | sha256sum -c -
if [ ${?} -ne 0 ]; then
    echo "Error: Checksum verification failed!"
    exit 1
fi

tar -xzf "${NAME}.tar.gz"

if [ ! -d "${NAME}" ] || [ ! -f "${NAME}/nu" ]; then
    echo "Error: Failed to find extracted 'nu' binary in ${NAME}."
    exit 1
fi

INSTALL_PATH="/usr/local/bin/nu"
if sudo mv "${NAME}/nu" "${INSTALL_PATH}"; then
    sudo chmod +x "${INSTALL_PATH}"
else
    echo "Error: Failed to move binary to ${INSTALL_PATH}."
    exit 1
fi

rm -f "${NAME}.tar.gz"
rm -rf "${NAME}"

echo "Nushell installed successfully to ${INSTALL_PATH}:"
${INSTALL_PATH} --version
