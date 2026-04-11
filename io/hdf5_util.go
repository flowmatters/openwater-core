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
*/
import "C"

import (
	"errors"
	"os"
	"reflect"
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

func rLockHDF5(fn string) {
	if !hdf5ThreadSafe {
		mu.Lock()
	}
}

func rUnlockHDF5(fn string) {
	if !hdf5ThreadSafe {
		mu.Unlock()
	}
}

func lockHDF5(fn string) {
	if !hdf5ThreadSafe {
		mu.Lock()
	}
}

func unlockHDF5(fn string) {
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

func openOrCreateDataset(f *hdf5.File, path string, shape []int, exampleValue interface{}, compress bool) (*hdf5.Dataset, error) {
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
	return createDataset(rootGroup, path, shape, exampleValue, compress)
}

func createDataset(g *hdf5.Group, path string, shape []int, exampleValue interface{}, compress bool) (*hdf5.Dataset, error) {
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

		if compress {
			dcpl, err := hdf5.NewPropList(hdf5.P_DATASET_CREATE)
			if err != nil {
				return nil, prefix("Cannot create property list", err)
			}
			defer dcpl.Close()

			dcpl.SetDeflate(hdf5.DefaultCompression)

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
	ds, err := createDataset(group, strings.Join(paths[1:], "/"), shape, exampleValue, compress)
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
