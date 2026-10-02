package main

import (
	"os"
	"sync"

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

	// The Python package writes "unknown" when it can't load the library it
	// is stamping for. That is the absence of a stamp, not a stamp.
	for name, value := range stamp {
		if value == "" || value == "unknown" {
			delete(stamp, name)
		}
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

// fileStamper stamps each file ow-sim writes, once, at the point the file is
// first written to.
//
// Not earlier: stamping at startup would create the results file before the
// run has produced anything, so a run that failed would leave a stamped,
// empty file behind -- and the next run would refuse to start without
// -overwrite.
//
// And never a file the run reads from. -final-states can name the model file
// itself, to write the final states back into it; that file's stamp says what
// it was built for and must survive.
type fileStamper struct {
	mu         sync.Mutex
	modelStamp map[string]string
	inputs     []os.FileInfo
	done       map[string]bool
	stamped    int
}

func newFileStamper(modelStamp map[string]string, inputFiles ...string) *fileStamper {
	s := &fileStamper{modelStamp: modelStamp, done: make(map[string]bool)}
	for _, fn := range inputFiles {
		if info, err := os.Stat(fn); err == nil {
			s.inputs = append(s.inputs, info)
		}
	}
	return s
}

// Ensure stamps fn if this stamper hasn't already dealt with it. Safe to call
// on a nil stamper, and from several goroutines. A failure is logged rather
// than fatal: the stamp is metadata, and not worth losing a run over.
func (s *fileStamper) Ensure(fn string) {
	if s == nil || fn == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done[fn] {
		return
	}
	s.done[fn] = true

	if info, err := os.Stat(fn); err == nil {
		for _, input := range s.inputs {
			if os.SameFile(info, input) {
				log.Debug().Str("Filename", fn).Msg("Not stamping a file this run reads its model from")
				return
			}
		}
	}

	stamp := resultsStamp(s.modelStamp)
	if err := io.SetFileAttributes(fn, stamp); err != nil {
		log.Warn().Err(err).Str("Filename", fn).Msg("Couldn't write version stamp")
		return
	}
	s.stamped++

	// A states or split file can be left over from an earlier run. Don't
	// let it go on claiming a model file this run didn't use.
	var stale []string
	for _, name := range []string{util.AttrModelVersion, util.AttrModelSignature} {
		if _, ok := stamp[name]; !ok {
			stale = append(stale, name)
		}
	}
	if len(stale) > 0 {
		if err := io.RemoveFileAttributes(fn, stale...); err != nil {
			log.Warn().Err(err).Str("Filename", fn).Msg("Couldn't clear a previous run's model stamp")
		}
	}
}

// outputStamper stamps the files this process writes. Nil until a simulation
// or writer sets it up, in which case nothing is stamped.
var outputStamper *fileStamper

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
