package io

import (
	"os"
	"strings"
	"testing"

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
