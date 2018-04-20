package io

import (
	"os"
	"strings"
	"testing"

	"github.com/flowmatters/openwater-core/data"

	"github.com/stretchr/testify/assert"
)

const TEST_PATH = "src/github.com/flowmatters/openwater-core/test/files"
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
