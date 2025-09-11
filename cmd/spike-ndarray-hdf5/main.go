// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
	"gonum.org/v1/hdf5"
)

const (
	fname  string = "NDArray.h5"
	dsname string = "Array2D"
)

func main() {
	arr := data.ARange[float64](80).MustReshape([]int{8, 10}).(data.ND2[float64])

	log.Info().Any("data", arr).Msg("Array data")

	// create data space
	dims := conv.IntsToUints(arr.Shape())
	log.Info().Any("Shape", dims).Msg("Array shape")
	space, err := hdf5.CreateSimpleDataspace(dims, nil)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}

	// create the file
	f, err := hdf5.CreateFile(fname, hdf5.F_ACC_TRUNC)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}
	defer f.Close()

	// create the memory data type
	dtype, err := hdf5.NewDatatypeFromValue(0.0)
	if err != nil {
		log.Panic().Msg("could not create a dtype")
	}

	// create the dataset
	dset, err := f.CreateDataset(dsname, dtype, space)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}

	// write data to the dataset
	log.Info().Msg("Writing to dataset")
	arrAsSlice := arr.Unroll()
	err = dset.Write(&arrAsSlice)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}
	log.Info().Msg("Dataset write successful")

	// release resources
	dset.Close()
	f.Close()

	// open the file and the dataset
	f, err = hdf5.OpenFile(fname, hdf5.F_ACC_RDONLY)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}
	dset, err = f.OpenDataset(dsname)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}

	space = dset.Space()
	dims, _, err = space.SimpleExtentDims()
	log.Info().Any("Dimensions", dims).Msg("Read array shape")

	// // read it back into a new slice
	// s2 := make([]s1Type, length)
	dest := data.NewArray[float64](conv.UintsToInts(dims))
	destAsSlice := dest.Unroll()
	err = dset.Read(&destAsSlice)
	if err != nil {
		log.Panic().Stack().Err(err).Msg("")
	}

	// display the fields
	log.Info().Any("Fields", dest).Msg("Fields")

	// release resources
	space.Close()
	dset.Close()
	f.Close()
}
