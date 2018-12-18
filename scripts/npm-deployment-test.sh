#!/bin/bash

set -e

NPM_UNPACK_DIR=/tmp/npm-unpack

echo "--- Packing npm package into a tarball"
npm pack

AUTHELIA_PACKAGE=`ls | grep "authelia-\([0-9]\+.\)\{2\}[0-9]\+.tgz"`
echo "--- Authelia package is ${AUTHELIA_PACKAGE}"

tar -tzvf ${AUTHELIA_PACKAGE}

echo "--- Copy package into "${NPM_UNPACK_DIR}" to test unpacking"
mkdir -p ${NPM_UNPACK_DIR}
cp ${AUTHELIA_PACKAGE} ${NPM_UNPACK_DIR}

pushd ${NPM_UNPACK_DIR}

echo "--- Test unpacking..."
npm install ${AUTHELIA_PACKAGE}

RET_CODE_INSTALL=$?
# echo ${RET_CODE}

# The binary must start and display the help menu
./node_modules/.bin/authelia | grep "No config file has been provided."
RET_CODE_RUN=$?

popd

if [ "$RET_CODE_INSTALL" != "0" ] || [ "$RET_CODE_RUN" != "0" ]
then
    echo "--- Unpacking failed..."
    exit 1
else
    echo "+++ Unpacking succeeded"
    exit 0
fi



