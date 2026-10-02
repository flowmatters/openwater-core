#!/bin/bash
#
# Point the Go/cgo build at a particular libhdf5.
#
# For the current shell only — source it, don't execute it:
#
#   source ./build/hdf5/env.sh            # the build from build-hdf5.sh
#   source ./build/hdf5/env.sh system     # whatever libhdf5 the system provides
#
# For every shell (writes CGO_CFLAGS/CGO_LDFLAGS into `go env`, i.e.
# ~/.config/go/env) — run it normally:
#
#   ./build/hdf5/env.sh --persist         # use the build from build-hdf5.sh
#   ./build/hdf5/env.sh --persist system  # clear the settings again
#
# Use --persist when something outside your shell also builds the project — a
# file-watcher auto-build (`entr ./build_all.sh`), an IDE, another terminal.
# Those inherit `go env` but not your exported variables, so without it they
# will quietly relink the binaries against the system libhdf5. Environment
# variables still win over `go env` settings, so the two forms compose.
#
# Rebuild after switching, then check what you actually linked against:
#   ./ow-sim -version        (prints "HDF5 thread-safe: true|false")
#
# Environment:
#   HDF5_INSTALL_DIR   prefix written by build-hdf5.sh (default $HOME/hdf5-install)

_ow_persist=0
_ow_mode=threadsafe
for _ow_arg in "$@"; do
    case "${_ow_arg}" in
        --persist) _ow_persist=1 ;;
        system | threadsafe) _ow_mode="${_ow_arg}" ;;
        *)
            echo "usage: [source] env.sh [threadsafe|system] [--persist]" >&2
            unset _ow_persist _ow_mode _ow_arg
            return 1 2>/dev/null || exit 1
            ;;
    esac
done
unset _ow_arg

if [ "${_ow_mode}" = "system" ]; then
    if [ "${_ow_persist}" = "1" ]; then
        go env -u CGO_CFLAGS CGO_LDFLAGS
        echo "HDF5: cleared CGO_CFLAGS/CGO_LDFLAGS from go env — all shells now use the system libhdf5"
    else
        unset CGO_CFLAGS
        unset CGO_LDFLAGS
        echo "HDF5: using system libhdf5 (CGO_CFLAGS/CGO_LDFLAGS cleared)"
    fi
else
    _ow_prefix="${HDF5_INSTALL_DIR:-$HOME/hdf5-install}"
    _ow_config="${_ow_prefix}/.openwater-hdf5-config"

    if [ ! -f "${_ow_config}" ]; then
        echo "No HDF5 install at ${_ow_prefix}." >&2
        echo "Build one first:  ./build/hdf5/build-hdf5.sh" >&2
        unset _ow_persist _ow_mode _ow_prefix _ow_config
        return 1 2>/dev/null || exit 1
    fi

    _ow_version=$(tr ' ' '\n' < "${_ow_config}" | sed -n 's/^version=//p')
    _ow_threadsafe=$(tr ' ' '\n' < "${_ow_config}" | sed -n 's/^threadsafe=//p')
    _ow_shared=$(tr ' ' '\n' < "${_ow_config}" | sed -n 's/^shared=//p')

    # The install lib dir is lib or lib64 depending on the platform/CMake.
    _ow_libdir="${_ow_prefix}/lib"
    if [ ! -d "${_ow_libdir}" ] && [ -d "${_ow_prefix}/lib64" ]; then
        _ow_libdir="${_ow_prefix}/lib64"
    fi

    _ow_cflags="-I${_ow_prefix}/include"

    # -lhdf5_hl before -lhdf5: HL depends on the core library, and link order
    # matters for the static case.
    case "$OSTYPE" in
        msys* | cygwin* | win*)
            _ow_ldflags="-L${_ow_libdir} -lhdf5_hl -lhdf5"
            ;;
        *)
            if [ "${_ow_shared}" = "ON" ]; then
                # rpath so built binaries resolve the library without
                # LD_LIBRARY_PATH/DYLD_LIBRARY_PATH being set at run time.
                _ow_ldflags="-L${_ow_libdir} -lhdf5_hl -lhdf5 -Wl,-rpath,${_ow_libdir}"
            else
                _ow_ldflags="-L${_ow_libdir} -lhdf5_hl -lhdf5 -lz -ldl -lm -lpthread"
            fi
            ;;
    esac

    if [ "${_ow_persist}" = "1" ]; then
        go env -w CGO_CFLAGS="${_ow_cflags}" CGO_LDFLAGS="${_ow_ldflags}"
    else
        export CGO_CFLAGS="${_ow_cflags}"
        export CGO_LDFLAGS="${_ow_ldflags}"
        case "$OSTYPE" in
            msys* | cygwin* | win*) export PATH="${_ow_prefix}/bin:$PATH" ;;
        esac
    fi

    echo "HDF5: ${_ow_prefix} (version=${_ow_version} threadsafe=${_ow_threadsafe} shared=${_ow_shared})"
    echo "  CGO_CFLAGS=${_ow_cflags}"
    echo "  CGO_LDFLAGS=${_ow_ldflags}"
    [ "${_ow_persist}" = "1" ] && echo "  written to go env — applies to every shell until: $0 --persist system"

    unset _ow_prefix _ow_config _ow_version _ow_threadsafe _ow_shared _ow_libdir
    unset _ow_cflags _ow_ldflags
fi

unset _ow_persist _ow_mode
