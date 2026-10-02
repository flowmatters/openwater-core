package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/data"
)

func TestFileAttributesRoundTrip(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "stamped.h5")

	written := map[string]string{
		"openwater_version": "1.0.3+abc1234.6d3bb266",
		"signature_hash":    "6d3bb266",
		"created_by":        "ow-sim",
	}
	require.NoError(t, SetFileAttributes(fn, written), "stamping creates the file when it doesn't exist")

	back, err := FileAttributes(fn, "openwater_version", "signature_hash", "created_by")
	require.NoError(t, err)
	assert.Equal(t, written, back)
}

func TestFileAttributesAreReplacedNotDuplicated(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "restamped.h5")

	require.NoError(t, SetFileAttributes(fn, map[string]string{"signature_hash": "aaaaaaaa", "created_by": "ow-sim"}))
	// A states file can outlive the run that first wrote it, so a stamp has to
	// be replaceable -- including by a value of a different length.
	require.NoError(t, SetFileAttributes(fn, map[string]string{"signature_hash": "a-much-longer-signature"}))

	back, err := FileAttributes(fn, "signature_hash", "created_by")
	require.NoError(t, err)
	assert.Equal(t, "a-much-longer-signature", back["signature_hash"])
	assert.Equal(t, "ow-sim", back["created_by"], "attributes not named in the second call are left alone")
}

func TestMissingFileAttributesAreOmitted(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "unstamped.h5")
	ref := H5Ref[float64]{Filename: fn, Dataset: "data"}
	require.NoError(t, ref.Write(arrayOf([]float64{1, 2, 3})))

	back, err := FileAttributes(fn, "openwater_version", "signature_hash")
	require.NoError(t, err, "an unstamped file is not an error")
	assert.Empty(t, back)
}

func TestFileAttributesSurviveAlongsideData(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "mixed.h5")
	ref := H5Ref[float64]{Filename: fn, Dataset: "/MODELS/Sum/outputs"}

	require.NoError(t, SetFileAttributes(fn, map[string]string{"signature_hash": "6d3bb266"}))
	values, err := data.ARange[float64](24).Reshape([]int{2, 3, 4})
	require.NoError(t, err)
	require.NoError(t, ref.Write(values))

	back, err := FileAttributes(fn, "signature_hash")
	require.NoError(t, err)
	assert.Equal(t, "6d3bb266", back["signature_hash"])

	loaded, err := ref.Load()
	require.NoError(t, err)
	assert.Equal(t, []int{2, 3, 4}, loaded.Shape())
}

func TestTextAttributeRoundTrip(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "named.h5")
	ref := H5Ref[float64]{Filename: fn, Dataset: "/MODELS/StorageRouting/outputs"}
	require.NoError(t, ref.Write(arrayOf([]float64{1, 2})))

	_, found, err := ref.TextAttribute("VARIABLES")
	require.NoError(t, err)
	assert.False(t, found, "no attribute until one is written")

	names := []string{"outflow", "storage"}
	require.NoError(t, ref.SetTextAttribute("VARIABLES", names))

	back, found, err := ref.TextAttribute("VARIABLES")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, names, back)

	// Replacing with a list of a different length and width.
	longer := []string{"a", "considerablyLongerVariableName", "c"}
	require.NoError(t, ref.SetTextAttribute("VARIABLES", longer))
	back, _, err = ref.TextAttribute("VARIABLES")
	require.NoError(t, err)
	assert.Equal(t, longer, back)
}

func TestEmptyTextAttribute(t *testing.T) {
	// A model with no states still gets a states dataset, [N, 0]. Its name
	// list is empty, which is different from not knowing the names.
	fn := filepath.Join(t.TempDir(), "empty.h5")
	ref := H5Ref[float64]{Filename: fn, Dataset: "states"}
	require.NoError(t, ref.Write(arrayOf([]float64{0})))

	require.NoError(t, ref.SetTextAttribute("VARIABLES", []string{}))

	back, found, err := ref.TextAttribute("VARIABLES")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Empty(t, back)
}

func TestTextAttributeNeedsTheDataset(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "nodataset.h5")
	require.NoError(t, SetFileAttributes(fn, map[string]string{"created_by": "test"}))

	ref := H5Ref[float64]{Filename: fn, Dataset: "/MODELS/Missing/outputs"}
	assert.Error(t, ref.SetTextAttribute("VARIABLES", []string{"a"}))
}

func TestStampingAFileThatIsNotHDF5IsAnError(t *testing.T) {
	// openWriteOrCreate used to return (nil, nil) for a file that exists but
	// can't be opened read-write, and every caller then dereferenced nil.
	fn := filepath.Join(t.TempDir(), "not_hdf5.h5")
	require.NoError(t, os.WriteFile(fn, []byte("this is not an HDF5 file"), 0o644))

	assert.Error(t, SetFileAttributes(fn, map[string]string{"created_by": "test"}))

	ref := H5Ref[float64]{Filename: fn, Dataset: "data"}
	assert.Error(t, ref.Create([]int{2, 2}, 0, 0), "the same goes for creating a dataset in it")
}

func TestRemoveFileAttributes(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "stamped.h5")
	require.NoError(t, SetFileAttributes(fn, map[string]string{
		"signature_hash":       "6d3bb266",
		"model_signature_hash": "deadbeef",
	}))

	require.NoError(t, RemoveFileAttributes(fn, "model_signature_hash", "never_there"))

	back, err := FileAttributes(fn, "signature_hash", "model_signature_hash")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"signature_hash": "6d3bb266"}, back)
}
