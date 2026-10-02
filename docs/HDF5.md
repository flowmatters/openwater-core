# HDF5 configuration

OpenWater talks to HDF5 through cgo (`gonum.org/v1/hdf5`). The library it links
against can be built with or without thread-safety, and OpenWater behaves
differently in each case.

**A thread-safe libhdf5 is the recommended configuration.** OpenWater works
correctly against a non-thread-safe one and CI covers both, so it is not a hard
requirement.

## Why thread-safety matters

`ow-sim` reads model inputs, parameters and states while a separate goroutine
writes results. HDF5 has one global internal state, so those calls have to be
serialized somewhere:

| libhdf5 | how calls are serialized | granularity |
| --- | --- | --- |
| built `--enable-threadsafe` | libhdf5's own recursive mutex | per C call |
| built without | a process-wide `sync.Mutex` in `io/hdf5_util.go` | per Go function — the whole open/read/close |

The Go-side fallback holds the lock across an entire file open, dataset
read and close, so a reader and the writer block each other for far longer than
they need to. `io/hdf5_util.go` detects which library it has at `init` via
`H5is_library_threadsafe` and picks the right strategy; nothing else in the
codebase needs to care.

Whether a packaged libhdf5 is thread-safe depends on who packaged it. Debian
and Ubuntu build their serial `libhdf5-dev` with thread-safety; Fedora's
`hdf5` and Homebrew's formula are without. Building it yourself is the only
way to get the recommended configuration everywhere.

Check what a binary ended up linked against:

```bash
./ow-sim -version
# ...
# HDF5 thread-safe: true
```

## Building a thread-safe libhdf5

```bash
./build/hdf5/build-hdf5.sh
```

Downloads and builds HDF5 (1.14.6 by default) into `$HOME/hdf5-install`, static
and `-fPIC` so it links into `libopenwater.so`/`.dylib` as well as the CLI
tools. Re-running is a no-op unless the requested configuration changed. Useful
knobs — see the header of the script for the full list:

| variable | default | |
| --- | --- | --- |
| `HDF5_INSTALL_DIR` | `$HOME/hdf5-install` | install prefix |
| `HDF5_VERSION` | `1.14.6` | release to build |
| `HDF5_THREADSAFE` | `ON` | set `OFF` to build a matched non-thread-safe library for comparison |
| `HDF5_SHARED` | `OFF` (`ON` on Windows) | shared instead of static |
| `HDF5_FORCE` | | `1` to rebuild regardless |

On Windows the library is built without zlib, because the build has no zlib
to link against there. Windows binaries therefore cannot write compressed
outputs (`-compress-outputs`) or read gzip-compressed datasets.

The build takes a few minutes. `ALLOW_UNSUPPORTED=ON` is required because the
HDF5 build system treats thread-safety plus the high-level (HL) library as an
unsupported combination, and gonum's bindings link `-lhdf5_hl`.

## Switching between libraries

`build/hdf5/env.sh` sets the cgo flags. For the current shell, source it:

```bash
source ./build/hdf5/env.sh            # the build from build-hdf5.sh
source ./build/hdf5/env.sh system     # whatever libhdf5 the system provides
```

Then rebuild (`./build/build.sh`, or `go build ./cmd/ow-sim`). Shared builds get
an `-rpath`, so binaries resolve the library without `LD_LIBRARY_PATH` set.

To make the choice apply to *every* shell, write it into `go env`
(`~/.config/go/env`) instead:

```bash
./build/hdf5/env.sh --persist         # note: run, don't source
./build/hdf5/env.sh --persist system  # undo
```

Use `--persist` if anything outside your terminal builds the project. The
project's tmux layout, for instance, runs a watcher:

```bash
while sleep 1 ; do find openwater-core -name '*.go' | entr -d ./build_all.sh; done
```

That shell has no `CGO_*` variables, so every time a `.go` file changes it
rebuilds and *overwrites* `ow-sim`, `libopenwater.so` and `../bin/*` against the
system libhdf5 — including the ones you just built with the flags exported. It
is not obvious when it happens: the binary keeps working, it just quietly loses
thread-safety. `go env` settings are inherited by those builds too, so
`--persist` keeps them consistent (exported variables still take precedence when
both are set).

Running `./build/build.sh` triggers this itself: the `ow-specgen` step rewrites
`generated_*.go`, which wakes the watcher.

Run the test suite both ways when touching anything in `io/` — the locking
paths are genuinely different:

```bash
source ./build/hdf5/env.sh
OW_TEST_PATH=$PWD/test/files go test -count=1 ./io/

source ./build/hdf5/env.sh system
OW_TEST_PATH=$PWD/test/files go test -count=1 ./io/
```

`io/hdf5_concurrent_test.go` is the part that exercises this: concurrent
readers against a writer, plus a guard that the Go-side mutex path does not
grow the OS thread pool with the number of waiters.

## CI

`build/ci/{linux,osx,windows}.sh` all call `build/hdf5/build-hdf5.sh` and write
the resulting cgo flags to `compilation_vars.txt`, which `build/ci/all.sh`
sources before building. The install directory is cached between runs by
GitHub Actions (`.github/workflows/build-ow-core.yml`); the cache key includes
the HDF5 version, and `build-hdf5.sh` rebuilds by itself if a cached prefix
was built with a different configuration. The cache is only saved when its key
misses, so bump the `-vN` suffix on the keys whenever the build configuration
changes — otherwise every run restores the stale prefix and rebuilds.

All three platforms build and test against the thread-safe library. The Linux
job then runs `build/ci/test-nonthreadsafe-hdf5.sh`, which builds a second
libhdf5 with `HDF5_THREADSAFE=OFF` (cached separately) and repeats the `io`
tests against it — the only place CI exercises the Go-side mutex. Ubuntu's own
`libhdf5-dev` is no use for this: it is thread-safe. Both runs set `OW_EXPECT_HDF5_THREADSAFE`, which
`TestHDF5ThreadSafetyReported` checks, so a run that links the wrong library
fails rather than silently covering the same path twice.

## Notes on the locking code

Two things in `io/hdf5_util.go` are easy to get wrong:

- **`h5_ensure_silenced` and `runtime.LockOSThread`.** In thread-safe builds,
  HDF5's auto-error-print setting lives on a *per-thread* error stack, so
  silencing it once on the main thread leaves worker threads printing warnings.
  Each locked region installs the NULL handler on the current thread (a
  thread-local guard makes this a branch after the first time) and pins the
  goroutine so the following cgo calls land on that same thread.

- **Lock order.** The mutex is acquired *before* `runtime.LockOSThread`. A
  goroutine that parks on a `sync.Mutex` while its M is locked takes that OS
  thread out of service, forcing the scheduler to spin up a replacement — one
  extra thread per waiter. Pinning after the mutex keeps waiters unpinned.
  `TestConcurrentAccessDoesNotLeakThreads` guards this.

- **The mutex is not reentrant.** Code running under a lock — including the
  callback passed to `WithReadFile`/`WithWriteFile` — must not call anything
  that locks again. It works against a thread-safe libhdf5, where the mutex is
  skipped, and deadlocks otherwise. Use `WithWriteFiles` to write to two files
  together, and the handle-taking variants (`LoadFromFile`, `WriteSliceToFile`,
  `CreateInFile`) inside a callback.

The pin itself costs about 2ns against a locked region that runs a file open,
dataset read and close, so it is not worth optimising further.
