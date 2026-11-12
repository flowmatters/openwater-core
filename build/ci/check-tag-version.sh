#!/bin/bash
#
# On a tag build, fail unless the tag names exactly the version this build
# produced: v{MAJOR.MINOR.PATCH}+{commit}.{model signature}.
#
# The tag is made by scripts/mark_for_release.sh from a local build, and a
# local build compiles whatever is in the working tree. An untracked model
# file there changes the signature, so the tag can name a signature that a
# clean checkout (this one) does not produce. Publishing under that name would
# tell users the release is compatible with model files it is not.
#
# Off a tag (branch pushes, nightlies) this does nothing.
#
# Run from openwater-core/ after ./build/build.sh.

set -e

if [ "${GITHUB_REF_TYPE}" != "tag" ]; then
    echo "Not a tag build — skipping tag/version check"
    exit 0
fi

BUILT=$(./ow-sim --version 2>/dev/null | awk 'NR==1 {print $2}')
if [ -z "${BUILT}" ]; then
    echo "ERROR: could not read the version from ./ow-sim --version" >&2
    exit 1
fi

if [ "v${BUILT}" != "${GITHUB_REF_NAME}" ]; then
    echo "ERROR: tag ${GITHUB_REF_NAME} does not match this build (v${BUILT})." >&2
    echo "The tag was probably made from a working tree with uncommitted or" >&2
    echo "untracked model files. Delete the tag and re-tag from a clean checkout." >&2
    exit 1
fi

echo "Tag ${GITHUB_REF_NAME} matches the built version"
