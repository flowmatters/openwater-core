package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

var showVersion = flag.Bool("version", false, "display version information")

// logEnsembleReports summarises what the models reported across the sweep.
// Members are identified by index: ows-ensemble is handed bare arrays, with no
// model file and so no way to name a place or a date.
func logEnsembleReports(model sim.TimeSteppingModel, sink *sim.ReportSink) {
	if sink == nil {
		return
	}

	reporting, ok := model.(sim.Reporting)
	if !ok {
		return
	}
	catalog := reporting.ReportCatalog()
	set := sink.Flush()

	for code := 0; code < set.NCodes && code < len(catalog); code++ {
		var total uint32
		members := 0
		for cell := 0; cell < set.NCells; cell++ {
			if n := set.Counts[cell*set.NCodes+code]; n > 0 {
				total += n
				members++
			}
		}
		if total == 0 {
			continue
		}

		event := log.Warn()
		if catalog[code].Severity >= sim.SevError {
			event = log.Error()
		}
		event.
			Str("Report", catalog[code].Name).
			Uint32("Occurrences", total).
			Int("Ensemble Members Affected", members).
			Int("Total Members", set.NCells).
			Msg("Model report")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ows-ensemble %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) < 4 {
		log.Fatal().Msg("Insufficient arguments.\nUsage: ows-ensemble <model> <hdf5pathtoinputs> <hdf5pathtoparameters> <hdf5pathtooutputs>")
	}

	modelName := args[0]
	inputPath := io.ParseH5Ref[float64](args[1])
	paramPath := io.ParseH5Ref[float64](args[2])
	outputPath := io.ParseH5Ref[float64](args[3])

	factory := sim.Catalog[modelName]
	if factory == nil {
		log.Fatal().Str("Model Name", modelName).Msg("Unknown model")
	}
	model := factory()

	inputs, err := inputPath.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	params, err := paramPath.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	model.ApplyParameters(params.(data.ND2[float64]))
	states := model.InitialiseStates(max(params.Len(1), inputs.Len(0)))
	cells := states.Len(sim.DIMS_CELL)
	outputs := sim.InitialiseOutputs(model, inputs.Len(sim.DIMI_TIMESTEP), cells)

	// An ensemble sweep pushes models into parameter regions the modeller has
	// not looked at, which is where warnings pay off most.
	var sink *sim.ReportSink
	if reporting, ok := model.(sim.Reporting); ok {
		sink = sim.NewReportSink(cells, len(reporting.ReportCatalog()), sim.SevWarning, 3)
		reporting.SetReportSink(sink)
	}

	model.Run(inputs.(data.ND3[float64]), states, outputs)
	logEnsembleReports(model, sink)

	err = outputPath.Write(outputs)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error writing outputs")
	}

	// err = outputPath.Write(outputs.States)

}
