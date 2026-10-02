#!/bin/bash
#
# Build and install a thread-safe libhdf5 for OpenWater.
#
# Thread-safe libhdf5 (--enable-threadsafe / HDF5_ENABLE_THREADSAFE=ON) is
# OpenWater's *recommended* configuration. With it, libhdf5 serializes
# internally with a recursive mutex at the C-call level and io/hdf5_util.go
# skips its process-wide Go mutex entirely, so concurrent readers and the
# writer goroutine in ow-sim contend far less. Whether a packaged libhdf5 has
# it varies (Debian/Ubuntu do, Fedora and Homebrew do not), so we build our own
# rather than depend on it.
#
# OpenWater still works against a non-thread-safe libhdf5 — io/hdf5_util.go
# detects the difference at init via H5is_library_threadsafe and falls back
# to serializing every HDF5 call behind a global Go mutex.
#
# Usage:
#   ./build/hdf5/build-hdf5.sh              # thread-safe, static, ~/hdf5-install
#   HDF5_THREADSAFE=OFF ./build/hdf5/build-hdf5.sh   # non-TS build, for comparison
#
# Then point the Go build at it:
#   source ./build/hdf5/env.sh
#
# Environment:
#   HDF5_VERSION       HDF5 release to build          (default 1.14.6)
#   HDF5_INSTALL_DIR   install prefix                 (default $HOME/hdf5-install)
#   HDF5_THREADSAFE    ON|OFF                         (default ON)
#   HDF5_SHARED        ON|OFF, build shared libs      (default OFF — static
#                                                      avoids any runtime
#                                                      library-path juggling)
#   HDF5_BUILD_DIR     scratch dir for the source     (default a temp dir)
#   HDF5_FORCE         1 to rebuild even if installed

set -e

HDF5_VERSION="${HDF5_VERSION:-1.14.6}"
HDF5_INSTALL_DIR="${HDF5_INSTALL_DIR:-$HOME/hdf5-install}"
HDF5_THREADSAFE="${HDF5_THREADSAFE:-ON}"
HDF5_SHARED="${HDF5_SHARED:-OFF}"
HDF5_BUILD_DIR="${HDF5_BUILD_DIR:-${TMPDIR:-/tmp}/openwater-hdf5-build}"

# Windows CI builds shared — MSVC static libhdf5 + cgo is a poor combination.
# It also builds without zlib: the runner has no zlib for MSVC to find, so
# Windows binaries cannot write compressed outputs (-compress-outputs) or read
# gzip-compressed datasets.
IS_WINDOWS=0
HDF5_ZLIB=ON
case "$OSTYPE" in
  msys* | cygwin* | win*) IS_WINDOWS=1; HDF5_SHARED=ON; HDF5_ZLIB=OFF ;;
esac

# Marker recording exactly what is installed at the prefix. env.sh reads it to
# work out the link flags, and we compare against it to decide whether the
# cached install matches what was asked for.
CONFIG_FILE="${HDF5_INSTALL_DIR}/.openwater-hdf5-config"
WANTED="version=${HDF5_VERSION} threadsafe=${HDF5_THREADSAFE} shared=${HDF5_SHARED}"

if [ "${HDF5_FORCE}" != "1" ] && [ -f "${CONFIG_FILE}" ] && [ "$(cat "${CONFIG_FILE}")" = "${WANTED}" ]; then
    echo "HDF5 already installed at ${HDF5_INSTALL_DIR} (${WANTED})"
    exit 0
fi

if [ -f "${CONFIG_FILE}" ]; then
    echo "Reinstalling HDF5 at ${HDF5_INSTALL_DIR}"
    echo "  have: $(cat "${CONFIG_FILE}")"
    echo "  want: ${WANTED}"
fi

if command -v nproc >/dev/null 2>&1; then
    JOBS=$(nproc)
elif command -v sysctl >/dev/null 2>&1; then
    JOBS=$(sysctl -n hw.ncpu)
else
    JOBS=4
fi

echo "Building HDF5 ${HDF5_VERSION} (${WANTED}) -> ${HDF5_INSTALL_DIR}"

SRC_DIR="${HDF5_BUILD_DIR}/hdf5-${HDF5_VERSION}"
if [ ! -d "${SRC_DIR}" ]; then
    mkdir -p "${HDF5_BUILD_DIR}"
    echo "Downloading hdf5-${HDF5_VERSION}.tar.gz"
    curl -sL "https://github.com/HDFGroup/hdf5/releases/download/hdf5_${HDF5_VERSION}/hdf5-${HDF5_VERSION}.tar.gz" \
        | tar xz -C "${HDF5_BUILD_DIR}"
fi

# ALLOW_UNSUPPORTED is required: the HDF5 build system considers thread-safety
# combined with the high-level (HL) library an unsupported combination. gonum's
# bindings link -lhdf5_hl, so we need HL and have to opt in explicitly. The
# combination is what the HDF5 project itself ships in its thread-safe builds
# and what OpenWater CI has been using.
CMAKE_ARGS=(
    "-DCMAKE_INSTALL_PREFIX=${HDF5_INSTALL_DIR}"
    "-DHDF5_ENABLE_THREADSAFE=${HDF5_THREADSAFE}"
    "-DHDF5_BUILD_HL_LIB=ON"
    "-DALLOW_UNSUPPORTED=ON"
    "-DHDF5_ENABLE_SZIP_SUPPORT=OFF"
    "-DHDF5_ENABLE_Z_LIB_SUPPORT=${HDF5_ZLIB}"
    "-DHDF5_BUILD_TOOLS=ON"
    "-DBUILD_TESTING=OFF"
    "-DCMAKE_BUILD_TYPE=Release"
)

if [ "${HDF5_SHARED}" = "ON" ]; then
    CMAKE_ARGS+=("-DBUILD_SHARED_LIBS=ON" "-DBUILD_STATIC_LIBS=OFF")
else
    # -fPIC so the static archives can be linked into libopenwater.so/.dylib,
    # which cgo builds with -buildmode=c-shared.
    CMAKE_ARGS+=("-DBUILD_SHARED_LIBS=OFF" "-DCMAKE_C_FLAGS=-fPIC")
fi

BUILD_SUBDIR="${SRC_DIR}/build-ts${HDF5_THREADSAFE}-shared${HDF5_SHARED}"
mkdir -p "${BUILD_SUBDIR}"
cd "${BUILD_SUBDIR}"

if [ "${IS_WINDOWS}" = "1" ]; then
    # No -G: let CMake pick whichever Visual Studio the machine has. Naming a
    # release here breaks every time the windows-latest runner image moves on.
    cmake .. "${CMAKE_ARGS[@]}"
    cmake --build . --config Release --parallel
    cmake --install . --config Release
else
    cmake .. "${CMAKE_ARGS[@]}"
    make -j"${JOBS}"
    make install
fi

echo "${WANTED}" > "${CONFIG_FILE}"
echo "HDF5 installed to ${HDF5_INSTALL_DIR}"
echo "Use it with:  source ./build/hdf5/env.sh"
