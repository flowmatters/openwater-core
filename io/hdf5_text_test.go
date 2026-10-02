package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/data"
)

func arrayOf[T data.Number](values []T) data.ND[T] {
	return data.ArrayFromSlice(values, []int{len(values)})
}

func TestWriteTextRoundTrip(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "text.h5")

	values := []string{"massBalanceNotConverged", "nanInput", "a"}
	ref := H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/StorageRouting/names"}
	require.NoError(t, ref.WriteText(values))

	back, err := ref.LoadText()
	require.NoError(t, err)
	assert.Equal(t, values, back, "strings must survive the round trip, padding and all")
}

func TestWriteTextCreatesNestedGroups(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "nested.h5")
	ref := H5Ref[int32]{Filename: fn, Dataset: "/a/b/c/values"}
	require.NoError(t, ref.WriteText([]string{"one", "two"}))

	back, err := ref.LoadText()
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, back)
}

func TestWriteTextAlongsideNumericData(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "mixed.h5")

	nums := H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Model/code"}
	require.NoError(t, nums.Write(arrayOf([]int32{0, 1, 2, 1})))

	names := H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Model/names"}
	require.NoError(t, names.WriteText([]string{"first", "second", "third"}))

	loaded, err := nums.Load()
	require.NoError(t, err)
	assert.Equal(t, []int32{0, 1, 2, 1}, loaded.Unroll())

	back, err := names.LoadText()
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second", "third"}, back)
}

func TestWriteTextRejectsEmpty(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "empty.h5")
	ref := H5Ref[int32]{Filename: fn, Dataset: "/nothing"}
	assert.Error(t, ref.WriteText(nil))

	_, err := os.Stat(fn)
	assert.True(t, os.IsNotExist(err), "a rejected write should not create the file")
}
