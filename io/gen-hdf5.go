// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package io

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"gonum.org/v1/hdf5"
)

type H5RefFloat64 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefFloat64) Load() (data.NDFloat64, error) {
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
	result := data.NewArrayFloat64(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefFloat64) loadSubset(ds *hdf5.Dataset) (data.NDFloat64, error) {
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

	result := data.NewArrayFloat64(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefFloat64) Write(data data.NDFloat64) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefFloat64) LoadText() ([]string, error) {
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

func (h H5RefFloat64) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefFloat64(path string) H5RefFloat64 {
	components := strings.Split(path, ":")
	return H5RefFloat64{components[0], components[1], nil}
}

type H5RefFloat32 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefFloat32) Load() (data.NDFloat32, error) {
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
	result := data.NewArrayFloat32(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefFloat32) loadSubset(ds *hdf5.Dataset) (data.NDFloat32, error) {
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

	result := data.NewArrayFloat32(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefFloat32) Write(data data.NDFloat32) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefFloat32) LoadText() ([]string, error) {
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

func (h H5RefFloat32) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefFloat32(path string) H5RefFloat32 {
	components := strings.Split(path, ":")
	return H5RefFloat32{components[0], components[1], nil}
}

type H5RefInt32 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefInt32) Load() (data.NDInt32, error) {
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
	result := data.NewArrayInt32(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefInt32) loadSubset(ds *hdf5.Dataset) (data.NDInt32, error) {
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

	result := data.NewArrayInt32(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefInt32) Write(data data.NDInt32) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefInt32) LoadText() ([]string, error) {
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

func (h H5RefInt32) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefInt32(path string) H5RefInt32 {
	components := strings.Split(path, ":")
	return H5RefInt32{components[0], components[1], nil}
}

type H5RefUint32 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefUint32) Load() (data.NDUint32, error) {
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
	result := data.NewArrayUint32(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefUint32) loadSubset(ds *hdf5.Dataset) (data.NDUint32, error) {
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

	result := data.NewArrayUint32(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefUint32) Write(data data.NDUint32) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefUint32) LoadText() ([]string, error) {
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

func (h H5RefUint32) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefUint32(path string) H5RefUint32 {
	components := strings.Split(path, ":")
	return H5RefUint32{components[0], components[1], nil}
}

type H5RefInt64 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefInt64) Load() (data.NDInt64, error) {
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
	result := data.NewArrayInt64(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefInt64) loadSubset(ds *hdf5.Dataset) (data.NDInt64, error) {
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

	result := data.NewArrayInt64(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefInt64) Write(data data.NDInt64) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefInt64) LoadText() ([]string, error) {
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

func (h H5RefInt64) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefInt64(path string) H5RefInt64 {
	components := strings.Split(path, ":")
	return H5RefInt64{components[0], components[1], nil}
}

type H5RefUint64 struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefUint64) Load() (data.NDUint64, error) {
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
	result := data.NewArrayUint64(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefUint64) loadSubset(ds *hdf5.Dataset) (data.NDUint64, error) {
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

	result := data.NewArrayUint64(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefUint64) Write(data data.NDUint64) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefUint64) LoadText() ([]string, error) {
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

func (h H5RefUint64) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefUint64(path string) H5RefUint64 {
	components := strings.Split(path, ":")
	return H5RefUint64{components[0], components[1], nil}
}

type H5RefInt struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefInt) Load() (data.NDInt, error) {
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
	result := data.NewArrayInt(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefInt) loadSubset(ds *hdf5.Dataset) (data.NDInt, error) {
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

	result := data.NewArrayInt(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefInt) Write(data data.NDInt) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefInt) LoadText() ([]string, error) {
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

func (h H5RefInt) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefInt(path string) H5RefInt {
	components := strings.Split(path, ":")
	return H5RefInt{components[0], components[1], nil}
}

type H5RefUint struct {
	Filename string
	Dataset  string
	Slice    [][]int
}

func (h H5RefUint) Load() (data.NDUint, error) {
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
	result := data.NewArrayUint(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5RefUint) loadSubset(ds *hdf5.Dataset) (data.NDUint, error) {
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

	result := data.NewArrayUint(shape)
	impl := result.Unroll()
	err = ds.ReadSubset(&impl, memSpace, filespace)
	return result, err
}

func (h H5RefUint) Write(data data.NDUint) error {
	f, err := hdf5.OpenFile(h.Filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.Filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.Filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.Dataset)
	if err == nil {
		// Ensure dataspace is the write size...
		// OR. just complain for now...

	} else {
		dtype, err := hdf5.NewDatatypeFromValue(0.0)
		if err != nil {
			return err
		}

		dims := conv.IntsToUints(data.Shape())
		space, err := hdf5.CreateSimpleDataspace(dims, nil)
		if err != nil {
			return err
		}
		defer space.Close()

		ds, err = f.CreateDataset(h.Dataset, dtype, space)
		if err != nil {
			return err
		}
	}
	defer ds.Close()

	arrAsSlice := data.Unroll()
	err = ds.Write(&arrAsSlice)
	if err != nil {
		return err
	}

	return nil
}

func (h H5RefUint) LoadText() ([]string, error) {
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

func (h H5RefUint) GetDatasets() ([]string, error) {
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
		if err == nil {
			ds.Close()
			result = append(result, name)
		}
	}
	return result, nil
}

func ParseH5RefUint(path string) H5RefUint {
	components := strings.Split(path, ":")
	return H5RefUint{components[0], components[1], nil}
}
