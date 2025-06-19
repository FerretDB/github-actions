#!/bin/bash

set -eu

VERSION=0.105.1
AMD64_CHECKSUM=ed26ae2708b1b74e8d84cda001ebd4b3b6b17b3292c123aef523a22bff463338
ARM64_CHECKSUM=3686cb9eb71d75505bf089140ac72fa6d33c2bef413676dafc464795c857c6b3

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
