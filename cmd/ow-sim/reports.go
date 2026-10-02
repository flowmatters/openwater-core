package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/sim"
)

// Model reports are written once, at the end of the run, at their exact final
// shape. There is deliberately no incremental append: the HDF5 layer creates
// dataspaces with fixed extents and the pinned gonum/hdf5 binding has no
// H5Dset_extent at all, so a growing dataset isn't expressible. Buffering for
// the whole run costs well under a megabyte on realistic models (see
// docs/MODEL_REPORTING.md) and makes the output deterministic besides.

const reportsGroup = "/REPORTS"

// writeReports persists every model's accumulated reports to the results file.
func writeReports(models map[string]*modelReference, modelNames []string, outputFn string) error {
	if outputFn == "" || reportSeverity() == sim.SevNone {
		return nil
	}

	for _, modelName := range modelNames {
		mr := models[modelName]
		if len(mr.reportCatalog) == 0 {
			continue
		}

		set := mr.Reports()
		if set.NCells == 0 {
			continue
		}

		if err := writeModelReports(outputFn, modelName, mr.reportCatalog, set); err != nil {
			return prefix("writing reports for "+modelName+": ", err)
		}
		// Reports can be the only thing in this file, when every model's
		// outputs went elsewhere. They still need to say what they came from.
		outputStamper.Ensure(outputFn)
	}

	return nil
}

func writeModelReports(outputFn, modelName string, catalog []sim.ReportDescription, set sim.ReportSet) error {
	base := fmt.Sprintf("%s/%s", reportsGroup, modelName)

	write := func(name string, values data.ND[int32]) error {
		ref := io.H5Ref[int32]{Filename: outputFn, Dataset: base + "/" + name}
		return ref.Write(values)
	}

	// Counts are dense and complete -- they cover every event, including those
	// whose detail record was dropped by the per-code cap -- so they are always
	// written, even when no detail record survived.
	counts := make([]int32, len(set.Counts))
	for i, c := range set.Counts {
		counts[i] = int32(c)
	}
	if err := write("counts", data.ArrayFromSlice(counts, []int{set.NCells, set.NCodes})); err != nil {
		return prefix("counts: ", err)
	}

	names := make([]string, len(catalog))
	messages := make([]string, len(catalog))
	severities := make([]int32, len(catalog))
	for i, entry := range catalog {
		names[i] = entry.Name
		messages[i] = entry.Template
		severities[i] = int32(entry.Severity)
	}

	// The catalogue travels with the reports so a results file can be read
	// without the exact engine build that produced it.
	nameRef := io.H5Ref[int32]{Filename: outputFn, Dataset: base + "/names"}
	if err := nameRef.WriteText(names); err != nil {
		return prefix("names: ", err)
	}
	messageRef := io.H5Ref[int32]{Filename: outputFn, Dataset: base + "/messages"}
	if err := messageRef.WriteText(messages); err != nil {
		return prefix("messages: ", err)
	}
	if err := write("code_severity", data.ArrayFromSlice(severities, []int{len(severities)})); err != nil {
		return prefix("code_severity: ", err)
	}

	if set.Len() == 0 {
		return nil
	}

	if err := write("run", data.ArrayFromSlice(set.Runs, []int{set.Len()})); err != nil {
		return prefix("run: ", err)
	}
	if err := write("timestep", data.ArrayFromSlice(set.Timesteps, []int{set.Len()})); err != nil {
		return prefix("timestep: ", err)
	}
	if err := write("code", data.ArrayFromSlice(set.Codes, []int{set.Len()})); err != nil {
		return prefix("code: ", err)
	}
	if err := write("severity", data.ArrayFromSlice(set.Severities, []int{set.Len()})); err != nil {
		return prefix("severity: ", err)
	}

	argsRef := io.H5Ref[float64]{Filename: outputFn, Dataset: base + "/args"}
	return argsRef.Write(data.ArrayFromSlice(set.Args, []int{set.Len(), sim.MaxReportArgs}))
}

// reportSummary is one line of the end-of-run summary.
type reportSummary struct {
	Model     string
	Code      int32
	Name      string
	Severity  sim.Severity
	Total     uint32
	Cells     int
	FirstStep int32
}

func summariseReports(models map[string]*modelReference, modelNames []string) []reportSummary {
	var summaries []reportSummary

	for _, modelName := range modelNames {
		mr := models[modelName]
		if len(mr.reportCatalog) == 0 {
			continue
		}

		set := mr.Reports()
		if set.NCells == 0 || set.NCodes == 0 {
			continue
		}

		for code := 0; code < set.NCodes; code++ {
			s := reportSummary{
				Model:     modelName,
				Code:      int32(code),
				Name:      mr.reportCatalog[code].Name,
				Severity:  mr.reportCatalog[code].Severity,
				FirstStep: -1,
			}

			for cell := 0; cell < set.NCells; cell++ {
				if n := set.Counts[cell*set.NCodes+code]; n > 0 {
					s.Total += n
					s.Cells++
				}
			}
			if s.Total == 0 {
				continue
			}

			// Earliest timestep among the surviving detail records. Config
			// faults carry -1, so only real timesteps update it.
			for i := 0; i < set.Len(); i++ {
				if set.Codes[i] != int32(code) {
					continue
				}
				if ts := set.Timesteps[i]; ts >= 0 && (s.FirstStep < 0 || ts < s.FirstStep) {
					s.FirstStep = ts
				}
			}

			summaries = append(summaries, s)
		}
	}

	// Most severe first, then noisiest.
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].Severity != summaries[j].Severity {
			return summaries[i].Severity > summaries[j].Severity
		}
		return summaries[i].Total > summaries[j].Total
	})

	return summaries
}

// logReportSummary prints what the models reported. Timesteps are left as
// indices: ow-sim has the model file but not the dimension decoding needed to
// name a place, so translating to dates and locations is left to the Python
// results API, which does.
func logReportSummary(summaries []reportSummary) {
	if len(summaries) == 0 {
		return
	}

	for _, s := range summaries {
		event := log.Warn()
		if s.Severity >= sim.SevError {
			event = log.Error()
		}

		event.
			Str("Model", s.Model).
			Str("Report", s.Name).
			Uint32("Occurrences", s.Total).
			Int("Cells Affected", s.Cells)

		if s.FirstStep >= 0 {
			event = event.Int32("First Timestep", s.FirstStep)
		}

		event.Msg("Model report")
	}
}

// exitCodeForReports returns 1 when a report at or above the configured fatal
// severity was raised.
func exitCodeForReports(summaries []reportSummary) int {
	threshold := reportsFatalSeverity()
	if threshold == sim.SevNone {
		return 0
	}

	for _, s := range summaries {
		if s.Severity >= threshold {
			return 1
		}
	}
	return 0
}

func finaliseReports(models map[string]*modelReference, modelNames []string, outputFn string) {
	if reportSeverity() == sim.SevNone {
		return
	}

	if err := writeReports(models, modelNames, outputFn); err != nil {
		log.Error().Stack().Err(err).Msg("Couldn't write model reports")
	}

	summaries := summariseReports(models, modelNames)
	logReportSummary(summaries)

	if code := exitCodeForReports(summaries); code != 0 {
		log.Error().Msg("Model reports include errors; exiting non-zero (disable with -reports-fatal=off)")
		os.Exit(code)
	}
}
