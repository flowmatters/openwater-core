package main

import (
	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/util"
)

// Every file ow-sim writes is stamped with the build that wrote it and, where
// the model file records one, the build the model file was made for. Results
// are read by position against a model's variable lists, and those lists
// change between builds, so a reader needs to know which build it is looking
// at. Datasets also carry their variable names, which lets a reader do without
// the matching build altogether.

// loadModelStamp reads the version stamp from the model file and warns when
// the model file was built for a different set of model interfaces. A model
// file with no stamp (it predates stamping, or wasn't written by openwater's
// Python package) yields an empty map and no warning.
func loadModelStamp(modelFn string) map[string]string {
	stamp, err := io.FileAttributes(modelFn, util.AttrVersion, util.AttrSignature)
	if err != nil {
		log.Debug().Err(err).Str("Model File", modelFn).Msg("Couldn't read version stamp from model file")
		return map[string]string{}
	}

	if signature, ok := stamp[util.AttrSignature]; ok && signature != util.GetSignatureHash() {
		log.Warn().
			Str("Model File", modelFn).
			Str("Model File Version", stamp[util.AttrVersion]).
			Str("Model File Signature", signature).
			Str("ow-sim Version", util.FullVersion()).
			Str("ow-sim Signature", util.GetSignatureHash()).
			Msg("Model file was built for a different set of model interfaces. Parameters, inputs and states may be misread")
	}
	return stamp
}

// resultsStamp is the set of root attributes for a file ow-sim writes.
func resultsStamp(modelStamp map[string]string) map[string]string {
	stamp := util.VersionStamp("ow-sim")
	if version, ok := modelStamp[util.AttrVersion]; ok {
		stamp[util.AttrModelVersion] = version
	}
	if signature, ok := modelStamp[util.AttrSignature]; ok {
		stamp[util.AttrModelSignature] = signature
	}
	return stamp
}

// stampResultsFile stamps one file. A failure is logged rather than fatal:
// the stamp is metadata, and not worth losing a run over.
func stampResultsFile(fn string, modelStamp map[string]string) {
	if err := io.SetFileAttributes(fn, resultsStamp(modelStamp)); err != nil {
		log.Warn().Err(err).Str("Filename", fn).Msg("Couldn't write version stamp")
	}
}

// stampOutputFiles stamps every file this process writes itself. Files handed
// to a -writer process are stamped by that process, which is the only one
// that should have them open.
func stampOutputFiles(models map[string]*modelReference, modelStamp map[string]string) {
	stamped := make(map[string]bool)
	for _, mr := range models {
		if mr.OutputProcess != nil {
			continue
		}
		for _, fn := range []string{mr.OutputFilename, mr.FinalStatesFilename} {
			if fn == "" || stamped[fn] {
				continue
			}
			stamped[fn] = true
			stampResultsFile(fn, modelStamp)
		}
	}
}

// nameVariables records the names of the variables along a dataset's variable
// axis. columns is the size of that axis: when it doesn't match the model's
// description there is no telling which name belongs to which column, and
// wrong names would be worse than none, so nothing is written.
func nameVariables(ref io.H5Ref[float64], names []string, columns int) {
	if len(names) != columns {
		log.Debug().
			Str("Dataset", ref.Dataset).
			Int("Columns", columns).
			Int("Names", len(names)).
			Msg("Variable count doesn't match the model description; not recording variable names")
		return
	}
	if err := ref.SetTextAttribute(util.AttrVariables, names); err != nil {
		log.Warn().Err(err).Str("Dataset", ref.Dataset).Msg("Couldn't record variable names")
	}
}
