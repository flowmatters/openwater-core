package io

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"fmt"

	"github.com/joelrahman/genny/generic"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/slice"
	"gonum.org/v1/hdf5"
)

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ArrayType=float64,float32,int32,uint32,int64,uint64,int,uint"

type ArrayType generic.Type

type H5RefArrayType struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefArrayType) Load() (data.NDArrayType, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)
	// mu.RLock()
	// defer mu.RUnlock()

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
	result := data.NewArrayArrayType(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefArrayType) loadSubset(ds *hdf5.Dataset) (data.NDArrayType, error) {
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

	result := data.NewArrayArrayType(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefArrayType) Write(data data.NDArrayType) error {
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

func (h H5RefArrayType) Create(shape []int, fillValue ArrayType, compress bool) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)
	// mu.Lock()
	// defer mu.Unlock()
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

func (h H5RefArrayType) WriteSlice(data data.NDArrayType, loc []int) error {
	lockHDF5(h.Filename)
	defer unlockHDF5(h.Filename)
	// mu.Lock()
	// defer mu.Unlock()
	f, err := openWriteOrCreate(h.Filename, false)
	if err != nil {
		return err
	}
	defer f.Close()

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

func (h H5RefArrayType) LoadText() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)
	// mu.RLock()
	// defer mu.RUnlock()

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

func (h H5RefArrayType) GetDatasets() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)
	// mu.RLock()
	// defer mu.RUnlock()

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

		ds, err := g.OpenDataset(name)
		if err == nil && ds != nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func (h H5RefArrayType) GetGroups() ([]string, error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)
	// mu.RLock()
	// defer mu.RUnlock()

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

		ds, err := g.OpenDataset(name)
		if err != nil && ds == nil {
			result = append(result, name)
		} else {
			ds.Close()
		}
	}
	return result, nil
}

func ParseH5RefArrayType(path string) H5RefArrayType {
	components := strings.Split(path, ":")
	return H5RefArrayType{components[0], components[1], nil}
}

func (h H5RefArrayType) Exists() bool {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)

	components := strings.Split(h.Dataset,"/")

	path := "/"
	for ix, comp := range(components) {
		ref := H5RefArrayType{Filename:h.Filename,Dataset: path}

		if ix == (len(components)-1) {
			datasets, err := ref.GetDatasets()

			if err != nil {
				return false
			}
			if findInSlice(datasets,comp) >= 0 {
				return true
			}
		}

		groups, err := ref.GetGroups()

		if err != nil {
			return false
		}
		if findInSlice(groups,comp) < 0 {
			return false
		}

		path = fmt.Sprintf("%s/%s",path,comp)
	}

	return true
}

func (h H5RefArrayType) Shape() ([]int,error) {
	rLockHDF5(h.Filename)
	defer rUnlockHDF5(h.Filename)
	// mu.RLock()
	// defer mu.RUnlock()

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
	return shape,nil
}
