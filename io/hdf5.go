package io

import (
	"os"
	"strings"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"gonum.org/v1/hdf5"
)

type H5Ref struct {
	filename string
	dataset  string
}

func (h H5Ref) Load() (data.NDFloat64, error) {
	f, err := hdf5.OpenFile(h.filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.dataset)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	space := ds.Space()
	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		return nil, err
	}

	shape := conv.UintsToInts(dims)
	result := data.NewArray(shape)
	impl := result.Unroll()
	ds.Read(&impl)
	return result, nil
}

func (h H5Ref) Write(data data.NDFloat64) error {
	f, err := hdf5.OpenFile(h.filename, hdf5.F_ACC_RDWR)
	if err != nil {
		if _, err := os.Stat(h.filename); os.IsNotExist(err) {
			f, err = hdf5.CreateFile(h.filename, hdf5.F_ACC_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	ds, err := f.OpenDataset(h.dataset)
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

		ds, err = f.CreateDataset(h.dataset, dtype, space)
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

func ParseH5Ref(path string) H5Ref {
	components := strings.Split(path, ":")
	return H5Ref{components[0], components[1]}
}
