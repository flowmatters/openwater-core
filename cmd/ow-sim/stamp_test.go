package main

import (
	"math"
	"os"
	"os/exec"
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

// unstartedCommand stands in for a -writer process. It is never run.
func unstartedCommand() *exec.Cmd {
	return exec.Command("true")
}

func TestResultsFileIsStamped(t *testing.T) {
	dir := t.TempDir()
	modelFn := filepath.Join(dir, "model.h5")
	resultsFn := filepath.Join(dir, "results.h5")

	// A model file as openwater's Python package stamps it, built for some
	// other set of model interfaces.
	require.NoError(t, io.SetFileAttributes(modelFn, map[string]string{
		util.AttrVersion:   "0.9.0+0000000.deadbeef",
		util.AttrSignature: "deadbeef",
	}))

	stampResultsFile(resultsFn, loadModelStamp(modelFn))

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

	stampResultsFile(resultsFn, loadModelStamp(modelFn))

	attrs := rootAttributes(t, resultsFn)
	assert.Equal(t, util.GetSignatureHash(), attrs[util.AttrSignature], "ow-sim's own stamp is always written")
	assert.NotContains(t, attrs, util.AttrModelSignature, "nothing is invented for the model file")
	assert.NotContains(t, attrs, util.AttrModelVersion)
}

func TestAMissingModelFileStillAllowsStamping(t *testing.T) {
	stamp := loadModelStamp(filepath.Join(t.TempDir(), "does_not_exist.h5"))
	assert.Empty(t, stamp)
}

func TestStampOutputFilesCoversEachFileOnce(t *testing.T) {
	dir := t.TempDir()
	outputs := filepath.Join(dir, "outputs.h5")
	states := filepath.Join(dir, "states.h5")
	split := filepath.Join(dir, "split.h5")

	models := map[string]*modelReference{
		"A": {ModelName: "A", OutputFilename: outputs, FinalStatesFilename: outputs},
		"B": {ModelName: "B", OutputFilename: outputs, FinalStatesFilename: states},
		// Handed to a -writer process, which stamps the file itself.
		"C": {ModelName: "C", OutputFilename: split, FinalStatesFilename: split, OutputProcess: unstartedCommand()},
	}

	stampOutputFiles(models, map[string]string{})

	assert.Equal(t, util.GetSignatureHash(), rootAttributes(t, outputs)[util.AttrSignature])
	assert.Equal(t, util.GetSignatureHash(), rootAttributes(t, states)[util.AttrSignature])
	_, err := os.Stat(split)
	assert.True(t, os.IsNotExist(err), "a file owned by a writer process must not be opened here")
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
