package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

func rootAttributes(t *testing.T, fn string) map[string]string {
	attrs, err := io.FileAttributes(fn,
		util.AttrVersion, util.AttrSignature, util.AttrCreatedBy, util.AttrCreatedTimestamp,
		util.AttrModelVersion, util.AttrModelSignature)
	require.NoError(t, err)
	return attrs
}

// stampedModelFile writes a model file as openwater's Python package stamps
// it, built for some other set of model interfaces.
func stampedModelFile(t *testing.T, dir string) string {
	fn := filepath.Join(dir, "model.h5")
	require.NoError(t, io.SetFileAttributes(fn, map[string]string{
		util.AttrVersion:   "0.9.0+0000000.deadbeef",
		util.AttrSignature: "deadbeef",
	}))
	return fn
}

func TestResultsFileIsStamped(t *testing.T) {
	dir := t.TempDir()
	modelFn := stampedModelFile(t, dir)
	resultsFn := filepath.Join(dir, "results.h5")

	newFileStamper(loadModelStamp(modelFn), modelFn).Ensure(resultsFn)

	attrs := rootAttributes(t, resultsFn)
	assert.Equal(t, util.FullVersion(), attrs[util.AttrVersion])
	assert.Equal(t, util.GetSignatureHash(), attrs[util.AttrSignature])
	assert.Equal(t, "ow-sim", attrs[util.AttrCreatedBy])
	assert.NotEmpty(t, attrs[util.AttrCreatedTimestamp])
	assert.Equal(t, "0.9.0+0000000.deadbeef", attrs[util.AttrModelVersion],
		"results record which model file build they came from")
	assert.Equal(t, "deadbeef", attrs[util.AttrModelSignature])
}

func TestResultsFromAnUnstampedModelFile(t *testing.T) {
	dir := t.TempDir()
	modelFn := filepath.Join(dir, "old_model.h5")
	resultsFn := filepath.Join(dir, "results.h5")

	// Model files written before stamping existed have no root attributes.
	ref := io.H5Ref[float64]{Filename: modelFn, Dataset: "/MODELS/Sum/states"}
	require.NoError(t, ref.Create([]int{2, 1}, 0, 0))

	newFileStamper(loadModelStamp(modelFn), modelFn).Ensure(resultsFn)

	attrs := rootAttributes(t, resultsFn)
	assert.Equal(t, util.GetSignatureHash(), attrs[util.AttrSignature], "ow-sim's own stamp is always written")
	assert.NotContains(t, attrs, util.AttrModelSignature, "nothing is invented for the model file")
	assert.NotContains(t, attrs, util.AttrModelVersion)
}

func TestAnUnknownModelStampIsNoStamp(t *testing.T) {
	// What the Python package writes when it can't load the library.
	modelFn := filepath.Join(t.TempDir(), "model.h5")
	require.NoError(t, io.SetFileAttributes(modelFn, map[string]string{
		util.AttrVersion:   "unknown",
		util.AttrSignature: "unknown",
	}))

	assert.Empty(t, loadModelStamp(modelFn))
}

func TestAMissingModelFileStillAllowsStamping(t *testing.T) {
	stamp := loadModelStamp(filepath.Join(t.TempDir(), "does_not_exist.h5"))
	assert.Empty(t, stamp)
}

func TestNothingIsStampedUntilItIsWritten(t *testing.T) {
	// A run that fails before producing anything must not leave a results
	// file behind: the next run would refuse to start without -overwrite.
	dir := t.TempDir()
	resultsFn := filepath.Join(dir, "results.h5")

	stamper := newFileStamper(loadModelStamp(stampedModelFile(t, dir)))

	_, err := os.Stat(resultsFn)
	assert.True(t, os.IsNotExist(err), "setting up the stamper creates nothing")
	assert.Equal(t, 0, stamper.stamped)
}

func TestEachFileIsStampedOnce(t *testing.T) {
	dir := t.TempDir()
	outputs := filepath.Join(dir, "outputs.h5")
	states := filepath.Join(dir, "states.h5")

	stamper := newFileStamper(map[string]string{})
	for i := 0; i < 3; i++ {
		stamper.Ensure(outputs)
		stamper.Ensure(states)
	}

	assert.Equal(t, 2, stamper.stamped)
	assert.Equal(t, util.GetSignatureHash(), rootAttributes(t, outputs)[util.AttrSignature])
	assert.Equal(t, util.GetSignatureHash(), rootAttributes(t, states)[util.AttrSignature])
}

func TestTheModelFileIsNeverRestamped(t *testing.T) {
	// -final-states can name the model file, to write states back into it.
	// Its stamp says what it was built for and has to survive that, or the
	// next run would see no mismatch and the results would misreport it.
	dir := t.TempDir()
	modelFn := stampedModelFile(t, dir)

	// Reached by another name, as the model file often is.
	link := filepath.Join(dir, "link_to_model.h5")
	require.NoError(t, os.Symlink(modelFn, link))

	stamper := newFileStamper(loadModelStamp(modelFn), modelFn)
	stamper.Ensure(modelFn)
	stamper.Ensure(link)

	attrs := rootAttributes(t, modelFn)
	assert.Equal(t, "deadbeef", attrs[util.AttrSignature])
	assert.Equal(t, "0.9.0+0000000.deadbeef", attrs[util.AttrVersion])
	assert.NotContains(t, attrs, util.AttrCreatedBy)
	assert.Equal(t, 0, stamper.stamped)
}

func TestAReusedFileDoesNotKeepThePreviousRunsModelStamp(t *testing.T) {
	dir := t.TempDir()
	statesFn := filepath.Join(dir, "states.h5")

	newFileStamper(loadModelStamp(stampedModelFile(t, dir))).Ensure(statesFn)
	require.Equal(t, "deadbeef", rootAttributes(t, statesFn)[util.AttrModelSignature])

	// The same states file, written again by a run from an unstamped model.
	newFileStamper(map[string]string{}).Ensure(statesFn)

	attrs := rootAttributes(t, statesFn)
	assert.NotContains(t, attrs, util.AttrModelSignature)
	assert.NotContains(t, attrs, util.AttrModelVersion)
	assert.Equal(t, util.GetSignatureHash(), attrs[util.AttrSignature])
}

func TestANilStamperDoesNothing(t *testing.T) {
	var stamper *fileStamper
	fn := filepath.Join(t.TempDir(), "results.h5")
	stamper.Ensure(fn)
	_, err := os.Stat(fn)
	assert.True(t, os.IsNotExist(err))
}

func TestAFileThatCannotBeStampedIsNotFatal(t *testing.T) {
	// Not an HDF5 file. This used to dereference nil.
	fn := filepath.Join(t.TempDir(), "not_hdf5.h5")
	require.NoError(t, os.WriteFile(fn, []byte("this is not an HDF5 file"), 0o644))

	stamper := newFileStamper(map[string]string{})
	stamper.Ensure(fn)
	assert.Equal(t, 0, stamper.stamped)
}

func TestVariableNamesAreRecorded(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "results.h5")
	ref := io.H5Ref[float64]{Filename: fn, Dataset: "/MODELS/StorageRouting/outputs"}
	require.NoError(t, ref.Create([]int{4, 2, 10}, math.NaN(), 0))

	names := sim.VariableNames([]sim.VariableDescription{{Name: "outflow"}, {Name: "storage"}})
	nameVariables(ref, names, 2)

	back, found, err := ref.TextAttribute(util.AttrVariables)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, []string{"outflow", "storage"}, back)
}

func TestVariableNamesAreNotGuessed(t *testing.T) {
	// Three columns but two names: there is no telling which column is which,
	// and a wrong name is worse than none.
	fn := filepath.Join(t.TempDir(), "results.h5")
	ref := io.H5Ref[float64]{Filename: fn, Dataset: "/MODELS/StorageRouting/outputs"}
	require.NoError(t, ref.Create([]int{4, 3, 10}, math.NaN(), 0))

	nameVariables(ref, []string{"outflow", "storage"}, 3)

	_, found, err := ref.TextAttribute(util.AttrVariables)
	require.NoError(t, err)
	assert.False(t, found)
}
