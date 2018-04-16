// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"gonum.org/v1/hdf5"
)

const (
	fname  string = "NDArray.h5"
	dsname string = "Array2D"
)

func main() {
	arr := data.ARangeFloat64(80).MustReshape([]int{8, 10}).(data.ND2Float64)

	fmt.Printf(":: data: %v\n", arr)

	// create data space
	dims := conv.IntsToUints(arr.Shape())
	fmt.Printf(":: data shape: %v\n", dims)
	space, err := hdf5.CreateSimpleDataspace(dims, nil)
	if err != nil {
		panic(err)
	}

	// create the file
	f, err := hdf5.CreateFile(fname, hdf5.F_ACC_TRUNC)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Printf(":: file [%s] created (id=%d)\n", fname, f.Id())

	// create the memory data type
	dtype, err := hdf5.NewDatatypeFromValue(0.0)
	if err != nil {
		panic("could not create a dtype")
	}

	// create the dataset
	dset, err := f.CreateDataset(dsname, dtype, space)
	if err != nil {
		panic(err)
	}
	fmt.Printf(":: dset (id=%d)\n", dset.Id())

	// write data to the dataset
	fmt.Printf(":: dset.Write...\n")
	arrAsSlice := arr.Unroll()
	err = dset.Write(&arrAsSlice)
	if err != nil {
		panic(err)
	}
	fmt.Printf(":: dset.Write... [ok]\n")

	// release resources
	dset.Close()
	f.Close()

	// open the file and the dataset
	f, err = hdf5.OpenFile(fname, hdf5.F_ACC_RDONLY)
	if err != nil {
		panic(err)
	}
	dset, err = f.OpenDataset(dsname)
	if err != nil {
		panic(err)
	}

	space = dset.Space()
	dims, _, err = space.SimpleExtentDims()
	fmt.Printf(":: data shape (R): %v\n", dims)

	// // read it back into a new slice
	// s2 := make([]s1Type, length)
	dest := data.NewArrayFloat64(conv.UintsToInts(dims))
	destAsSlice := dest.Unroll()
	err = dset.Read(&destAsSlice)
	if err != nil {
		panic(err)
	}

	// display the fields
	fmt.Printf(":: data: %v\n", dest)

	// release resources
	space.Close()
	dset.Close()
	f.Close()
}
