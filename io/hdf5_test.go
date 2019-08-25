package io

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/flowmatters/openwater-core/data"

	"github.com/stretchr/testify/assert"
)

const TEST_PATH = "openwater-core/test/files"
const TEST_FILE = "test_hdf5.h5"

func findInSlice(strings []string, target string) int {
	for i, v := range strings {
		if v == target {
			return i
		}
	}
	return -1
}

func test_filename() string {
	return strings.Join([]string{
		os.Getenv("GOPATH"),
		TEST_PATH,
		TEST_FILE}, string(os.PathSeparator))
}

func TestReadStrings(t *testing.T) {
	assert := assert.New(t)
	fn := test_filename()
	ref := H5RefFloat64{Filename: fn, Dataset: "simple/strings"}
	theStrings, err := ref.LoadText()

	assert.Nil(err)
	assert.Equal(4, len(theStrings))
	assert.Equal("one string", theStrings[0])
	assert.Equal("three strings", theStrings[2])
}

func TestGetDatasetNames(t *testing.T) {
	assert := assert.New(t)
	fn := test_filename()

	ref := H5RefFloat64{Filename: fn, Dataset: "simple"}
	datasetNames, err := ref.GetDatasets()

	assert.Nil(err)
	assert.Equal(3, len(datasetNames))
	assert.True(findInSlice(datasetNames, "strings") >= 0)
	assert.True(findInSlice(datasetNames, "ints") >= 0)
	assert.True(findInSlice(datasetNames, "doubles") >= 0)
	assert.Equal(-1, findInSlice(datasetNames, "not there"))
}

func TestReadDouble(t *testing.T) {
	assert := assert.New(t)
	fn := test_filename()

	ref := H5RefFloat64{Filename: fn, Dataset: "simple/doubles"}
	all, err := ref.Load()
	assert.Nil(err)

	assert.Equal(1, len(all.Shape()))
	assert.Equal(8, all.Shape()[0])

	slice := [][]int{[]int{2, 6, 1}}
	subset := H5RefFloat64{Filename: fn, Dataset: "simple/doubles", Slice: slice}
	all, err = subset.Load()
	assert.Nil(err)

	assert.Equal(1, len(all.Shape()))
	assert.Equal(4, all.Shape()[0])
	all1 := all.(data.ND1Float64)
	assert.Equal(2.0, all1.Get1(0))
	assert.Equal(8.0, all1.Get1(3))
}

func TestRead3DInts(t *testing.T) {
	assert := assert.New(t)
	fn := test_filename()
	ds := "ints3d"

	ref := H5RefInt32{Filename: fn, Dataset: ds}
	all, err := ref.Load()
	assert.Nil(err)

	assert.True(all.Contiguous())
	assert.Equal(3, len(all.Shape()))
	assert.Equal(5, all.Shape()[0])
	assert.Equal(10, all.Shape()[1])
	assert.Equal(4, all.Shape()[2])
	all3d := all.(data.ND3Int32)
	assert.Equal(int32(19), all3d.Get3(0, 4, 3))
	assert.Equal(int32(154), all3d.Get3(3, 8, 2))

	unrolled := all.Unroll()
	assert.Equal(int32(19), unrolled[19])
	assert.Equal(int32(154), unrolled[154])

	slice := [][]int{nil, []int{2, 6, 1}, []int{1, 3, 1}}
	subset := H5RefInt32{Filename: fn, Dataset: ds, Slice: slice}
	all, err = subset.Load()
	assert.Nil(err)

	assert.Equal(3, len(all.Shape()))
	assert.Equal(5, all.Shape()[0])
	assert.Equal(4, all.Shape()[1])
	assert.Equal(2, all.Shape()[2])

	all3d = all.(data.ND3Int32)
	assert.Equal(int32(61), all3d.Get3(1, 3, 0))
	assert.Equal(int32(134), all3d.Get3(3, 1, 1))
}

func TestWrite3DFloat64Whole(t *testing.T) {
	assert := assert.New(t)
	test_fn := "_test_write_whole.h5"
	ds := "float64_3d"

	ref := H5RefFloat64{Filename: test_fn, Dataset: ds}

	the_data, err := data.ARangeFloat64(1000).Reshape([]int{10, 20, 5})
	assert.Nil(err)

	indices := [][]int{
		[]int{0, 0, 0},
		[]int{5, 14, 2},
		[]int{7, 3, 4},
		[]int{9, 19, 4},
	}

	values := make([]float64, 4)
	for i, idx := range indices {
		values[i] = the_data.Get(idx)
	}

	err = ref.Write(the_data)
	assert.Nil(err)

	ref_read := H5RefFloat64{Filename: test_fn, Dataset: ds}
	read_data, err := ref_read.Load()
	assert.Nil(err)

	shp := read_data.Shape()
	assert.Equal(3, len(shp))
	assert.Equal(10, shp[0])
	assert.Equal(20, shp[1])
	assert.Equal(5, shp[2])

	for i, idx := range indices {
		assert.Equal(values[i], read_data.Get(idx),
			"Error in test data point %d [%d,%d,%d]", i, idx[0], idx[1], idx[2])
	}
}

func TestWriteTwice(t *testing.T) {
	assert := assert.New(t)
	test_fn := "_test_write_twice.h5"
	ds := "float64_1d"

	ref := H5RefFloat64{Filename: test_fn, Dataset: ds}

	the_data := data.ARangeFloat64(10)

	err := ref.Write(the_data)
	assert.Nil(err)

	ref_read := H5RefFloat64{Filename: test_fn, Dataset: ds}
	read_data, err := ref_read.Load()
	assert.Nil(err)

	shp := read_data.Shape()
	assert.Equal(1, len(shp))
	assert.Equal(10, shp[0])

	assert.Equal(4.0, read_data.Get([]int{4}))

	new_data := data.NewArray1DFloat64(10)
	data.ScaleFloat64Array(new_data, the_data, 2)

	err = ref.Write(new_data)
	assert.Nil(err)

	new_ref := H5RefFloat64{Filename: test_fn, Dataset: ds}
	new_read_data, err := new_ref.Load()
	assert.Nil(err)

	shp = new_read_data.Shape()
	assert.Equal(1, len(shp))
	assert.Equal(10, shp[0])

	assert.Equal(8.0, new_read_data.Get([]int{4}))
}

// func TestResizeOnRewrite(t *testing.T) {
// 	assert := assert.New(t)
// 	test_fn := "_test_write_twice_resize.h5"
// 	ds := "float64_1d"

// 	ref := H5RefFloat64{Filename: test_fn, Dataset: ds}

// 	the_data := data.ARangeFloat64(10)

// 	err := ref.Write(the_data)
// 	assert.Nil(err)

// 	ref_read := H5RefFloat64{Filename: test_fn, Dataset: ds}
// 	read_data, err := ref_read.Load()
// 	assert.Nil(err)

// 	shp := read_data.Shape()
// 	assert.Equal(1, len(shp))
// 	assert.Equal(10, shp[0])

// 	assert.Equal(4.0, read_data.Get([]int{4}))

// 	new_data := data.ARangeFloat64(5)
// 	data.ScaleFloat64Array(new_data, new_data, 2)

// 	err = ref.Write(new_data)
// 	assert.Nil(err)

// 	new_ref := H5RefFloat64{Filename: test_fn, Dataset: ds}
// 	new_read_data, err := new_ref.Load()
// 	assert.Nil(err)

// 	shp = new_read_data.Shape()
// 	assert.Equal(1, len(shp))
// 	assert.Equal(5, shp[0])

// 	assert.Equal(8.0, new_read_data.Get([]int{4}))

// 	new_data = data.ARangeFloat64(20)
// 	data.ScaleFloat64Array(new_data, new_data, 2)

// 	err = ref.Write(new_data)
// 	assert.Nil(err)

// 	new_ref = H5RefFloat64{Filename: test_fn, Dataset: ds}
// 	new_read_data, err = new_ref.Load()
// 	assert.Nil(err)

// 	shp = new_read_data.Shape()
// 	assert.Equal(1, len(shp))
// 	assert.Equal(20, shp[0])

// 	assert.Equal(24.0, new_read_data.Get([]int{12}))

// }

func TestWrite3DInt32Whole(t *testing.T) {
	assert := assert.New(t)
	test_fn := "_test_write_whole.h5"
	ds := "NESTED/int32_3d"

	ref := H5RefInt32{Filename: test_fn, Dataset: ds}

	the_data, err := data.ARangeInt32(1000).Reshape([]int{10, 20, 5})
	assert.Nil(err)

	indices := [][]int{
		[]int{0, 0, 0},
		[]int{5, 14, 2},
		[]int{7, 3, 4},
		[]int{9, 19, 4},
	}

	values := make([]int32, 4)
	for i, idx := range indices {
		values[i] = the_data.Get(idx)
	}

	err = ref.Write(the_data)
	assert.Nil(err)

	ref_read := H5RefInt32{Filename: test_fn, Dataset: ds}
	read_data, err := ref_read.Load()
	assert.Nil(err)

	shp := read_data.Shape()
	assert.Equal(3, len(shp))
	assert.Equal(10, shp[0])
	assert.Equal(20, shp[1])
	assert.Equal(5, shp[2])

	for i, idx := range indices {
		assert.Equal(values[i], read_data.Get(idx),
			"Error in test data point %d [%d,%d,%d]", i, idx[0], idx[1], idx[2])
	}
}

func TestWrite3DFloat64Partial(t *testing.T) {
	assert := assert.New(t)
	test_fn := "_test_write_partial.h5"
	ds := "float64_3d"

	ref := H5RefFloat64{Filename: test_fn, Dataset: ds}

	the_data, err := data.ARangeFloat64(1000).Reshape([]int{10, 20, 5})
	assert.Nil(err)

	slice := the_data.Slice([]int{0, 0, 0}, []int{10, 1, 1}, []int{1, 1, 1})

	indices := [][]int{
		[]int{0, 0, 0},
		[]int{5, 0, 0},
		[]int{7, 0, 0},
		[]int{9, 0, 0},
	}

	values := make([]float64, 4)
	for i, idx := range indices {
		values[i] = the_data.Get(idx)
	}

	err = ref.Create(the_data.Shape(), math.NaN())
	for d2 := 0; d2 < 20; d2++ {
		for d3 := 0; d3 < 5; d3++ {
			err = ref.WriteSlice(slice, []int{0, d2, d3})
			assert.Nil(err)
		}
	}

	ref_read := H5RefFloat64{Filename: test_fn, Dataset: ds}
	read_data, err := ref_read.Load()
	assert.Nil(err)

	shp := read_data.Shape()
	assert.Equal(3, len(shp))
	assert.Equal(10, shp[0])
	assert.Equal(20, shp[1])
	assert.Equal(5, shp[2])

	for d2 := 0; d2 < 20; d2++ {
		for d3 := 0; d3 < 5; d3++ {
			for i, idx := range indices {
				the_idx := []int{idx[0], d2, d3}
				assert.Equal(values[i], read_data.Get(the_idx),
					"Error in test data point %d [%d,%d,%d]", i, the_idx[0], the_idx[1], the_idx[2])
			}
		}
	}
}
