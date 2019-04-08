package io

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/util/m"
	"github.com/flowmatters/openwater-core/util/slice"
	"gonum.org/v1/hdf5"
)

var mu sync.RWMutex

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func prefix(msg string, e error) error {
	return &errorString{msg + e.Error()}
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
	return m.MaxInt(0, (m.MinInt(size, slice[1])-m.MinInt(size, slice[0]))) / slice[2]
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

func openOrCreateDataset(f *hdf5.File, path string, shape []int, exampleValue interface{}) (*hdf5.Dataset, error) {
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
	return createDataset(rootGroup, path, shape, exampleValue)
}

func createDataset(g *hdf5.Group, path string, shape []int, exampleValue interface{}) (*hdf5.Dataset, error) {
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
	ds, err := createDataset(group, strings.Join(paths[1:], "/"), shape, exampleValue)
	if err != nil {
		return nil, prefix(paths[0]+": ", err)
	}
	return ds, nil
}
