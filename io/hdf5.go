package io

import (
	"bytes"
	"errors"
	"reflect"
	"strings"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/slice"
	"gonum.org/v1/hdf5"
)


type H5Ref[T data.Number] struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5Ref[T]) Load() (data.ND[T], error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return h.loadFromOpenFile(f)
}

// LoadFromFile loads data from an already-open HDF5 file handle. The
// caller must hold the global HDF5 mutex and manage the file lifecycle.
// Use this via WithReadFile for batching multiple reads under a single
// mutex acquisition.
func (h H5Ref[T]) LoadFromFile(f *hdf5.File) (data.ND[T], error) {
	return h.loadFromOpenFile(f)
}

func (h H5Ref[T]) loadFromOpenFile(f *hdf5.File) (data.ND[T], error) {
	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	if h.Slice != nil {
		for _, s := range h.Slice {
			if s != nil {
				return h.loadSubset(ds)
			}
		}
	}

	space := ds.Space()
	defer space.Close()

	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}

	shape := conv.UintsToInts(dims)
	result := data.NewArray[T](shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5Ref[T]) loadSubset(ds *hdf5.Dataset) (data.ND[T], error) {
	space := ds.Space()
	defer space.Close()

	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}
	shape := conv.UintsToInts(dims)

	offset, stride, count, block := makeHyperslab(h.Slice, shape)
	filespace := space
	err = filespace.SelectHyperslab(offset, stride, count, block)
	if err != nil {
		return nil, err
	}

	for dim, size := range shape {
		if h.Slice[dim] != nil {
			newSize := sliceSize(h.Slice[dim], size)
			shape[dim] = newSize
		}
	}
	ushape := conv.IntsToUints(shape)
	memSpace, err := hdf5.CreateSimpleDataspace(ushape, ushape)
	if err != nil {
		return nil, err
	}
	defer memSpace.Close()

	result := data.NewArray[T](shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5Ref[T]) Write(data data.ND[T]) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)
	// mu.Lock()
	// defer mu.Unlock()
	f, err := openWriteOrCreate(h.Filename, true)
	if err != nil {
		return err
	}
	defer f.Close()

	ds, err := openOrCreateDataset(f, h.Dataset, data.Shape(), data.Get(data.NewIndex(0)), false)
	if err != nil {
		return err
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5Ref[T]) Create(shape []int, fillValue T, compress bool) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)

	f, err := openWriteOrCreate(h.Filename, true)
	if err != nil {
		return err
	}
	defer f.Close()

	ds, err := openOrCreateDataset(f, h.Dataset, shape, fillValue, compress)
	if err == nil {
		ds.Close()
	}
	return err
}

// CreateInFile creates a dataset in an already-open HDF5 file. The caller
// must hold the global HDF5 mutex and manage the file lifecycle.
func (h H5Ref[T]) CreateInFile(f *hdf5.File, shape []int, fillValue T, compress bool) error {
	ds, err := openOrCreateDataset(f, h.Dataset, shape, fillValue, compress)
	if err == nil {
		ds.Close()
	}
	return err
}

func (h H5Ref[T]) WriteSlice(data data.ND[T], loc []int) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)

	f, err := openWriteOrCreate(h.Filename, false)
	if err != nil {
		return err
	}
	defer f.Close()

	return h.writeSliceToOpenFile(f, data, loc)
}

// WriteSliceToFile writes a slice of data into an already-open HDF5 file.
// The caller must hold the global HDF5 mutex and manage the file lifecycle.
// Use via WithWriteFile for batching multiple writes under a single mutex
// acquisition and file open.
func (h H5Ref[T]) WriteSliceToFile(f *hdf5.File, data data.ND[T], loc []int) error {
	return h.writeSliceToOpenFile(f, data, loc)
}

func (h H5Ref[T]) writeSliceToOpenFile(f *hdf5.File, data data.ND[T], loc []int) error {
	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return err
	}
	defer ds.Close()

	filespace := ds.Space()
	defer filespace.Close()

	shp := conv.IntsToUints(data.Shape())
	stride_count := conv.IntsToUints(slice.Ones(len(loc)))
	err = filespace.SelectHyperslab(conv.IntsToUints(loc), stride_count, stride_count, shp)
	if err != nil {
		return err
	}

	memSpace, err := hdf5.CreateSimpleDataspace(shp, shp)
	if err != nil {
		return err
	}
	defer memSpace.Close()

	impl := data.Unroll()
	err = ds.WriteSubset(&impl, memSpace, filespace)

	return nil
}

func (h H5Ref[T]) LoadText() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	dt, err := ds.Datatype()
	defer dt.Close()

	if dt.GoType() != reflect.TypeOf("a string") {
		return nil, errors.New("Not a string type")
	}

	space := ds.Space()
	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}

	if len(dims) > 1 {
		return nil, errors.New("Can only read 1D data as text")
	}
	maxLen := int(dt.Size())
	nStrings := int(dims[0])
	characters := make([]byte, nStrings*maxLen)
	ds.Read(&characters)

	result := make([]string, dims[0])
	for i := 0; i < nStrings; i++ {
		theBytes := characters[(i * maxLen):((i + 1) * maxLen)]
		end := bytes.Index(theBytes, []byte{0})
		if end < 0 {
			end = maxLen
		}
		theBytes = theBytes[0:end]
		result[i] = string(theBytes)
	}
	return result, nil
}

func (h H5Ref[T]) GetDatasets() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := f.OpenGroup(h.Dataset)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	n, err := g.NumObjects()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for i := 0; i < int(n); i++ {
		name, err := g.ObjectNameByIndex(uint(i))
		if err != nil {
			return nil, err
		}

		objType, err := g.ObjectTypeByIndex(uint(i))
		if err == nil && objType == hdf5.H5G_DATASET {
			result = append(result, name)
		}
	}
	return result, nil
}

func (h H5Ref[T]) GetGroups() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := f.OpenGroup(h.Dataset)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	n, err := g.NumObjects()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for i := 0; i < int(n); i++ {
		name, err := g.ObjectNameByIndex(uint(i))
		if err != nil {
			return nil, err
		}

		objType, err := g.ObjectTypeByIndex(uint(i))
		if err == nil && objType == hdf5.H5G_GROUP {
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5Ref[T data.Number](path string) H5Ref[T] {
	components := strings.Split(path, ":")
	return H5Ref[T]{components[0], components[1], nil}
}

// Exists reports whether the object (dataset or group) at h.Dataset is
// present in h.Filename. It walks the path one component at a time using
// H5Lexists (a cheap library-native check), rather than enumerating every
// child of every group along the path as the earlier implementation did.
//
// For a path like /MODELS/Foo/inputs on a file containing 27 top-level
// models the old implementation did four fresh file opens and four
// full-group enumerations; the new implementation does one file open, a
// handful of H5Lexists calls, and a couple of OpenGroup/Close pairs, all
// under a single acquisition of the global HDF5 mutex.
func (h H5Ref[T]) Exists() bool {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return false
	}
	defer f.Close()

	path := strings.Trim(h.Dataset, "/")
	if path == "" {
		return true
	}
	components := strings.Split(path, "/")

	// Track groups we open along the walk so we can close them all on exit
	// via a single defer — matters if some intermediate OpenGroup succeeds
	// but a later check fails.
	var openGroups []*hdf5.Group
	defer func() {
		for _, g := range openGroups {
			g.Close()
		}
	}()

	// Check first component at the file root.
	if !f.LinkExists(components[0]) {
		return false
	}
	if len(components) == 1 {
		return true
	}
	grp, err := f.OpenGroup(components[0])
	if err != nil {
		return false
	}
	openGroups = append(openGroups, grp)

	// Walk the remaining components inside the currently-open group.
	for i := 1; i < len(components); i++ {
		if !grp.LinkExists(components[i]) {
			return false
		}
		if i == len(components)-1 {
			return true
		}
		next, err := grp.OpenGroup(components[i])
		if err != nil {
			return false
		}
		openGroups = append(openGroups, next)
		grp = next
	}
	return true
}

func (h H5Ref[T]) Shape() ([]int, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	space := ds.Space()
	defer space.Close()

	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}

	shape := conv.UintsToInts(dims)
	return shape, nil
}
