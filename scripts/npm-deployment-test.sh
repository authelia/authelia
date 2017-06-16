#!/bin/bash

NPM_UNPACK_DIR=/tmp/npm-unpack/test

echo "Packing npm package into a tarball"
npm pack

AUTHELIA_PACKAGE=`ls | grep "authelia-\([0-9]\+.\)\{2\}[0-9]\+.tgz"`
echo "Authelia package is ${AUTHELIA_PACKAGE}"

echo "Copy package into "${NPM_UNPACK_DIR}" to test unpacking"
mkdir -p ${NPM_UNPACK_DIR}
cp ${AUTHELIA_PACKAGE} ${NPM_UNPACK_DIR}

pushd ${NPM_UNPACK_DIR}

echo "Test unpacking..."
npm install ${AUTHELIA_PACKAGE}

RET_CODE=$?
# echo ${RET_CODE}

popd

if [ "$RET_CODE" != "0" ]
then
    echo "Unpacking failed..."
    exit 1
else
    echo "Unpacking succeeded"
    exit 0
fi



