package io

/*
#include "hdf5.h"

// h5_is_threadsafe wraps H5is_library_threadsafe which is always present
// in libhdf5 (even non-threadsafe builds export it — it returns false).
static int h5_is_threadsafe() {
    hbool_t ts = 0;
    if (H5is_library_threadsafe(&ts) < 0) return 0;
    return (int)ts;
}

// In threadsafe libhdf5 builds the auto-error-print setting (the callback
// installed via H5Eset_auto2) lives on a per-thread error stack. Silencing
// on the main thread does not silence worker threads that the Go runtime
// uses to service cgo calls for parallel goroutines. h5_ensure_silenced
// installs the NULL handler on the current thread's default stack the
// first time it runs there; subsequent calls on the same thread are a
// trivial branch on the thread-local guard. Callers must hold the OS
// thread (runtime.LockOSThread) so the next cgo call lands on the same
// thread we just silenced.
static __thread int _h5_thread_silenced = 0;
static void h5_ensure_silenced() {
    if (!_h5_thread_silenced) {
        H5Eset_auto2(H5E_DEFAULT, NULL, NULL);
        _h5_thread_silenced = 1;
    }
}
*/
import "C"

import (
	"errors"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/util/m"
	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/rs/zerolog/log"
	"gonum.org/v1/hdf5"
)

// hdf5ThreadSafe is true when the linked libhdf5 was built with
// --enable-threadsafe. Detected once at package init via the C API
// H5is_library_threadsafe. When true, libhdf5 handles its own internal
// locking and the Go-side lock functions become no-ops. When false, we
// fall back to a process-wide sync.Mutex that serializes every HDF5 call.
var hdf5ThreadSafe bool

// mu serializes HDF5 operations when libhdf5 is NOT thread-safe. When
// libhdf5 IS thread-safe, rLockHDF5/lockHDF5 are no-ops and the library's
// internal recursive mutex handles serialization at a finer granularity
// (per-cgo-call rather than per-Go-function), which significantly reduces
// contention between the main goroutine's reads and the writer goroutine's
// writes.
var mu sync.Mutex

func init() {
	hdf5ThreadSafe = C.h5_is_threadsafe() != 0
	if hdf5ThreadSafe {
		log.Info().Msg("HDF5 library is thread-safe — Go-side locking disabled")
	} else {
		log.Debug().Msg("HDF5 library is NOT thread-safe — using Go-side global mutex")
	}
}

// IsHDF5ThreadSafe reports whether the linked HDF5 library was built with
// thread-safety support. Exposed for diagnostics (e.g. ow-sim -version).
func IsHDF5ThreadSafe() bool {
	return hdf5ThreadSafe
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Each lock/unlock pair pins the goroutine to its OS thread and ensures
// HDF5's per-thread auto-error-print is silenced on that thread. The OS
// thread pin is required even when libhdf5 is thread-safe (so go-side
// mutex is skipped) because cgo otherwise lets the goroutine migrate
// between cgo calls, and we need the silence to apply to the same thread
// that subsequently runs the HDF5 ops.
//
// Order matters: acquire mu FIRST, then pin. A goroutine that parks on a
// sync.Mutex while its M is locked takes that M out of service, forcing
// the scheduler to spin up a replacement thread for the P. Pinning after
// the mutex keeps waiters unpinned, so the thread pool stays at ~GOMAXPROCS
// instead of growing to one thread per concurrent HDF5 caller.

func rLockHDF5(fn string) {
	if !hdf5ThreadSafe {
		mu.Lock()
	}
	runtime.LockOSThread()
	C.h5_ensure_silenced()
}

func rUnlockHDF5(fn string) {
	runtime.UnlockOSThread()
	if !hdf5ThreadSafe {
		mu.Unlock()
	}
}

func lockHDF5(fn string) {
	if !hdf5ThreadSafe {
		mu.Lock()
	}
	runtime.LockOSThread()
	C.h5_ensure_silenced()
}

func unlockHDF5(fn string) {
	runtime.UnlockOSThread()
	if !hdf5ThreadSafe {
		mu.Unlock()
	}
}

func prefix(msg string, e error) error {
	return &errorString{msg + e.Error()}
}

// WithReadFile opens an HDF5 file for reading under the global HDF5 mutex
// (when needed) and passes the open file handle to fn. The file is closed
// and the mutex released when fn returns. This allows multiple datasets
// from the same file to be read under a single mutex acquisition rather
// than paying the open/lock/close cycle per dataset.
func WithReadFile(filename string, fn func(f *hdf5.File) error) error {
	lockHDF5(filename)
	defer unlockHDF5(filename)
	f, err := hdf5.OpenFile(filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

// WithWriteFile opens an HDF5 file for read-write under the global HDF5
// mutex and passes the open file handle to fn. The file is closed and
// the mutex released when fn returns. Mirrors WithReadFile for writes.
// fn must not call back into anything that takes the HDF5 mutex — see
// WithWriteFiles for writing to two files together.
func WithWriteFile(filename string, fn func(f *hdf5.File) error) error {
	lockHDF5(filename)
	defer unlockHDF5(filename)
	f, err := openWriteOrCreate(filename, true)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

// WithWriteFiles is WithWriteFile for two files at once: both are opened
// read-write under a single acquisition of the global HDF5 mutex. When the
// two names are the same file it is opened once and fn receives the same
// handle twice.
//
// Use this rather than nesting WithWriteFile calls. The Go-side mutex is not
// reentrant, so a nested call deadlocks whenever libhdf5 is not thread-safe.
func WithWriteFiles(first, second string, fn func(f1, f2 *hdf5.File) error) error {
	lockHDF5(first)
	defer unlockHDF5(first)
	f1, err := openWriteOrCreate(first, true)
	if err != nil {
		return err
	}
	defer f1.Close()
	if second == first {
		return fn(f1, f1)
	}
	f2, err := openWriteOrCreate(second, true)
	if err != nil {
		return err
	}
	defer f2.Close()
	return fn(f1, f2)
}

func makeHyperslab(slice [][]int, dims []int) (offset, stride, count, block []uint) {
	offset = make([]uint, len(slice), len(slice))
	stride = make([]uint, len(slice), len(slice))
	count = make([]uint, len(slice), len(slice))
	block = make([]uint, len(slice), len(slice))

	for i, dim := range slice {
		if dim == nil {
			offset[i] = 0
			stride[i] = 1
			count[i] = uint(dims[i])
		} else {
			offset[i] = uint(dim[0])
			stride[i] = uint(dim[2])
			count[i] = uint(sliceSize(dim, dims[i]))
		}
		block[i] = 1
	}
	return offset, stride, count, block
}

func sliceSize(slice []int, size int) int {
	return m.Max[int](0, (m.Min[int](size, slice[1])-m.Min[int](size, slice[0]))) / slice[2]
}

func openWriteOrCreate(fn string, createIfNotExist bool) (*hdf5.File, error) {
	f, err := hdf5.OpenFile(fn, hdf5.F_ACC_RDWR)
	if err != nil {
		if !createIfNotExist {
			return nil, prefix("Cannot open file: "+fn, err)
		}

		if _, err := os.Stat(fn); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(fn, hdf5.F_ACC_TRUNC)
			if err != nil {
				return nil, prefix("Cannot create file: ", err)
			}
		}
	}
	return f, nil
}

func shapesMatch(ds *hdf5.Dataset, shape []int) bool {
	space := ds.Space()
	defer space.Close()

	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return false
	}

	dsShape := conv.UintsToInts(dims)

	return slice.Equal(dsShape, shape)
}

func openOrCreateDataset(f *hdf5.File, path string, shape []int, exampleValue interface{}, compressLevel int) (*hdf5.Dataset, error) {
	ds, err := f.OpenDataset(path)
	if err == nil {
		if !shapesMatch(ds, shape) {
			ds.Close()
			return nil, errors.New("Cannot resize datasets")
		}
		return ds, nil
	}

	rootGroup, err := f.OpenGroup("/")
	if err != nil {
		return nil, prefix("Cannot open root group in file "+f.FileName()+": ", err)
	}
	defer rootGroup.Close()
	return createDataset(rootGroup, path, shape, exampleValue, compressLevel)
}

func createDataset(g *hdf5.Group, path string, shape []int, exampleValue interface{}, compressLevel int) (*hdf5.Dataset, error) {
	paths := strings.Split(path, "/")
	if paths[0] == "" {
		paths = paths[1:]
	}
	if len(paths) == 1 {
		dtype, err := hdf5.NewDataTypeFromType(reflect.TypeOf(exampleValue))
		if err != nil {
			return nil, prefix("Cannot match datatype", err)
		}
		defer dtype.Close()

		dims := conv.IntsToUints(shape)
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return nil, prefix("Cannot create dataspace", err)
		}
		defer space.Close()

		if compressLevel > 0 {
			dcpl, err := hdf5.NewPropList(hdf5.P_DATASET_CREATE)
			if err != nil {
				return nil, prefix("Cannot create property list", err)
			}
			defer dcpl.Close()

			// Deflate requires chunked storage. Chunk along the first
			// dimension (nodes): [1, dim1, dim2, ...]. This matches the
			// per-node write pattern in WriteSlice and gives good
			// compression (each chunk is one node's full timeseries).
			chunkDims := make([]uint, len(dims))
			copy(chunkDims, dims)
			chunkDims[0] = 1
			if err := dcpl.SetChunk(chunkDims); err != nil {
				return nil, prefix("Cannot set chunk dimensions", err)
			}
			dcpl.SetDeflate(compressLevel)

			ds, err := g.CreateDatasetWith(paths[0], dtype, space, dcpl)
			if err != nil {
				return nil, prefix("Cannot create dataset  "+path+": ", err)
			}
			return ds, nil
		}

		ds, err := g.CreateDataset(paths[0], dtype, space)
		if err != nil {
			return nil, prefix("Cannot create dataset  "+path+": ", err)
		}
		return ds, nil
	}

	group, err := g.OpenGroup(paths[0])
	if err != nil {
		group, err = g.CreateGroup(paths[0])
		if err != nil {
			return nil, prefix("Cannot open or create group "+paths[0]+": ", err)
		}
	}
	defer group.Close()
	ds, err := createDataset(group, strings.Join(paths[1:], "/"), shape, exampleValue, compressLevel)
	if err != nil {
		return nil, prefix(paths[0]+": ", err)
	}
	return ds, nil
}

func findInSlice(strings []string, target string) int {
	for i, v := range strings {
		if v == target {
			return i
		}
	}
	return -1
}
