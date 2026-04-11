package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gonum.org/v1/hdf5"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/util"
)

const (
	LINK_SRC_GENERATION  = 0
	LINK_SRC_MODEL       = 1
	LINK_SRC_NODE        = 2
	LINK_SRC_GEN_NODE    = 3
	LINK_SRC_VAR         = 4
	LINK_DEST_GENERATION = 5
	LINK_DEST_MODEL      = 6
	LINK_DEST_NODE       = 7
	LINK_DEST_GEN_NODE   = 8
	LINK_DEST_VAR        = 9
)

func main() {
	flag.Parse()
	setupLogger()

	if *showVersion {
		fmt.Printf("ow-sim %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		fmt.Printf("HDF5 thread-safe: %v\n", io.IsHDF5ThreadSafe())
		os.Exit(0)
	}

	hdf5.DisplayErrors(false)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()

	if *writerMode {
		run_writer(args)
	} else {
		run_simulation(args)
	}
}

func run_simulation(args []string) {
	if len(args) == 0 {
		log.Fatal().Msg("Must specify an input file")
	}
	fn := args[0]
	var outputFn string = ""
	if len(args) > 1 {
		outputFn = args[1]

		if _, err := os.Stat(outputFn); err == nil {
			if *overwrite {
				os.Remove(outputFn)
			} else {
				log.Fatal().Str("Output Filename", outputFn).Msg("Output file exists and overwrite not set. Delete file or use -overwrite")
			}
		}
	}

	modelsRef := io.H5Ref[float64]{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5Ref[float64]{Filename: fn, Dataset: "/DIMENSIONS"}
	//	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

	modelNames, err := modelsRef.LoadText()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Couldn't read model metadata")
	}
	log.Debug().Any("Models", modelNames).Msg("")

	dims, err := dimsRef.GetDatasets()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Couldn't read model dimensions")
	}
	log.Debug().Any("Dimensions", dims).Msg("")

	linksRef := io.H5Ref[uint32]{Filename: fn, Dataset: "/LINKS"}
	linksND, err := linksRef.Load()
	links := linksND.(data.ND2[uint32])
	// Raw flat view of the LINKS table: row r column c is at linksRaw[r*linkStride + c].
	// Hand-indexing this avoids allocating a fresh ND1 slice per link in the hot loop.
	const linkStride = LINK_DEST_VAR + 1
	linksRaw := links.Unroll()
	nLinks := links.Len(0)
	nextLink := 0
	simStart := time.Now()

	totalTimeSimulation := 0.0
	totalTimeFinalWrite := 0.0
	totalTimeLinks := 0.0

	models, genCount, nodeCount := makeModelRefs(modelNames, fn, outputFn)
	nodesCompleted := 0

	// Flat index of model references by model number (the LINKS table uses
	// numeric model indices into modelNames). Avoids a string + map lookup per
	// link in the hot loop.
	modelList := make([]*modelReference, len(modelNames))
	for idx, name := range modelNames {
		modelList[idx] = models[name]
	}

	// Link-loop worker pool configuration. The link phase is memory-bandwidth
	// bound (dest[k] += src[k] across tens of GB of data on realistic models),
	// so parallelism helps mostly by driving multiple memory channels. Default
	// caps workers below NumCPU to leave headroom for the writer goroutine and
	// the Go runtime; override with -link-workers for benchmarking.
	numLinkWorkers := *linkWorkers
	if numLinkWorkers <= 0 {
		numLinkWorkers = runtime.NumCPU()
		if numLinkWorkers > 8 {
			numLinkWorkers = 8
		}
	}
	if numLinkWorkers < 1 {
		numLinkWorkers = 1
	}
	log.Debug().Int("Link Workers", numLinkWorkers).Msg("Link-loop parallelism")
	// Per-worker buckets of link row indices. Allocated once here and reset
	// (length = 0, capacity retained) at the start of each generation's link
	// phase to avoid allocating in the hot path.
	linkBuckets := make([][]int, numLinkWorkers)
	for b := range linkBuckets {
		linkBuckets[b] = make([]int, 0, 16)
	}
	var linkWG sync.WaitGroup

	// Per-generation destination pre-warm scratch map. Reused across the
	// generation loop (cleared each iteration) so we don't allocate per link.
	// Sized for a typical per-gen destination count.
	destInitSet := make(map[uint64]struct{}, 64)

	// One done-channel per generation. Writer for gen g waits on writingDone[g-1],
	// writes, then closes writingDone[g]. This gives strictly-ordered, sleep-free
	// hand-off between generation writers.
	writingDone := make([]chan struct{}, genCount)
	for i := range writingDone {
		writingDone[i] = make(chan struct{})
	}

	for i := 0; i < genCount; i++ {
		pcComplete := 100.0 * float64(nodesCompleted) / float64(nodeCount)
		log.Info().Float64("Percent Complete", pcComplete).Int("Generation", i+1).Int("Total Generations", genCount).Msg("Generation progress")

		// === RUN GENERATION ===
		genSimulationTime, nodesInGeneration := runGeneration(i, models, modelNames) // synchronous
		nodesCompleted += nodesInGeneration
		totalTimeSimulation += genSimulationTime
		// === /RUN GENERATION ===

		// === WRITE GENERATION OUTPUTS ===
		// asynchronous: one goroutine per generation, strictly ordered via
		// writingDone[g-1] -> writingDone[g].
		if outputFn != "" {
			go func(g int) {
				if g > 0 {
					<-writingDone[g-1]
					for _, modelName := range modelNames {
						models[modelName].PurgeGeneration(g - 1)
					}
				}

				writeGeneration(g, models, modelNames)
				close(writingDone[g])
			}(i)
		}
		// === /WRITE GENERATION OUTPUTS ===

		// === PROCESS LINKS ===
		// Parallel fan-out. The LINKS table is flat uint32[nLinks*linkStride]
		// and generation Outputs/Inputs are contiguous ND3[float64] laid out
		// C-order as [nNodes, nVars, nTimesteps]. For each link we compute
		// the linear row of src/dest directly and accumulate with a tight
		// scalar loop.
		//
		// Correctness of parallelism: multiple links may target the same
		// destination row (that is how accumulation works). To avoid write
		// conflicts without locking, we hash-partition links by the full
		// destination identity (destModel, destGen, destIdx, destVar) into
		// numLinkWorkers buckets. Two links that target the same dest row
		// always land in the same bucket and are processed sequentially by a
		// single worker in the original link-table order, so both the result
		// and the floating-point summation order are preserved and stable
		// across runs.
		genLinkStart := time.Now()
		currentLink := nextLink
		var partitionElapsed, workersElapsed time.Duration

		// Determine [currentLink, endLink) = all links with srcGen <= i.
		// linksRaw is sorted by srcGen (the serial version used this too via
		// the break condition), so a forward scan suffices.
		endLink := currentLink
		for endLink < nLinks {
			if linksRaw[endLink*linkStride+LINK_SRC_GENERATION] > uint32(i) {
				break
			}
			endLink++
		}
		nextLink = endLink

		if endLink > currentLink {
			partitionStart := time.Now()
			// Lazy per-generation destination init (serial on main). We walk
			// this generation's links once and initialize any destination
			// (model, gen) pair that hasn't been initialized yet, so that the
			// parallel link workers that follow can run pure-math loops
			// without ever touching HDF5 themselves. Doing this on main keeps
			// memory bounded to the "link reach window" rather than allocating
			// every destination generation up front, which would peak memory
			// at O(all generations) — infeasible on larger models.
			//
			// In the same pass we hash-partition links by full destination
			// identity into numLinkWorkers buckets so that all links writing
			// to the same dest row land in the same bucket and are processed
			// sequentially by one worker (preserves determinism and summation
			// order).
			for b := range linkBuckets {
				linkBuckets[b] = linkBuckets[b][:0]
			}
			for k := range destInitSet {
				delete(destInitSet, k)
			}
			nw := uint32(numLinkWorkers)
			for r := currentLink; r < endLink; r++ {
				base := r * linkStride

				destModelNum := linksRaw[base+LINK_DEST_MODEL]
				destGenNum := linksRaw[base+LINK_DEST_GENERATION]
				key := uint64(destModelNum)<<32 | uint64(destGenNum)
				if _, ok := destInitSet[key]; !ok {
					destInitSet[key] = struct{}{}
					destModelRef := modelList[destModelNum]
					if destModelRef.Generations[int(destGenNum)] == nil {
						if _, err := destModelRef.GetGeneration(int(destGenNum)); err != nil {
							log.Fatal().Stack().Err(err).Uint32("Model", destModelNum).Uint32("Generation", destGenNum).Msg("Lazy destination init failed")
						}
					}
				}

				// FNV-ish mix of the 4 dest identity fields. Cheap and
				// well-distributed across modest bucket counts.
				h := destModelNum
				h = h*16777619 ^ destGenNum
				h = h*16777619 ^ linksRaw[base+LINK_DEST_GEN_NODE]
				h = h*16777619 ^ linksRaw[base+LINK_DEST_VAR]
				linkBuckets[h%nw] = append(linkBuckets[h%nw], r)
			}
			partitionElapsed = time.Since(partitionStart)

			workersStart := time.Now()
			// Fan out. Each worker processes its bucket sequentially.
			//
			// Workers read modelList[X].Generations[Y] directly rather than
			// calling GetGeneration(), which would acquire the per-model genMu
			// mutex once per link. In the hash-partitioned setup, source
			// models are shared across worker buckets — all buckets contain
			// links coming from gen i, and gen i's source models are
			// identical across buckets — so those GetGeneration calls would
			// serialize every worker behind the same few mutexes.
			//
			// Direct reads are safe because:
			//   (a) source gens for srcGen == i are populated by the
			//       runGeneration(i) call on main just above;
			//   (b) destination gens are populated by the partition-pass
			//       GetGeneration calls on main immediately above;
			//   (c) the `go func()` fan-out is a happens-before synchronization
			//       point in Go's memory model, so workers observe all writes
			//       main made before spawning them.
			linkWG.Add(numLinkWorkers)
			for b := 0; b < numLinkWorkers; b++ {
				go func(bucket []int) {
					defer linkWG.Done()
					for _, r := range bucket {
						base := r * linkStride
						linkGen := linksRaw[base+LINK_SRC_GENERATION]

						srcGen := modelList[linksRaw[base+LINK_SRC_MODEL]].Generations[int(linkGen)]
						destGen := modelList[linksRaw[base+LINK_DEST_MODEL]].Generations[int(linksRaw[base+LINK_DEST_GENERATION])]

						srcShape := srcGen.Outputs.Shape() // [nodes, nVars, nTimesteps]
						nVarsSrc := srcShape[1]
						nTimesteps := srcShape[2]
						srcBuf := srcGen.Outputs.Unroll()
						srcStart := (int(linksRaw[base+LINK_SRC_GEN_NODE])*nVarsSrc + int(linksRaw[base+LINK_SRC_VAR])) * nTimesteps

						destShape := destGen.Inputs.Shape()
						nVarsDest := destShape[1]
						destBuf := destGen.Inputs.Unroll()
						destStart := (int(linksRaw[base+LINK_DEST_GEN_NODE])*nVarsDest + int(linksRaw[base+LINK_DEST_VAR])) * nTimesteps

						destRow := destBuf[destStart : destStart+nTimesteps]
						srcRow := srcBuf[srcStart : srcStart+nTimesteps]
						for k := range destRow {
							destRow[k] += srcRow[k]
						}
					}
				}(linkBuckets[b])
			}
			linkWG.Wait()
			workersElapsed = time.Since(workersStart)
		}

		genLinkEnd := time.Now()
		genLinkElapsed := genLinkEnd.Sub(genLinkStart).Seconds()
		totalTimeLinks += genLinkElapsed
		log.Debug().
			Int("Generation", i).
			Int("Links", nextLink-currentLink).
			Float64("Partition (s)", partitionElapsed.Seconds()).
			Float64("Workers (s)", workersElapsed.Seconds()).
			Float64("Total (s)", genLinkElapsed).
			Msg("Link phase timing")
		// === /PROCESS LINKS ===

		genElapsed := genLinkElapsed + genSimulationTime
		log.Debug().Float64("Elapsed Seconds", genElapsed).Msg("Generation completed")
	}

	log.Debug().Msg("Simulation finished. Waiting for results to be written")
	generationsEnd := time.Now()

	if outputFn != "" {
		<-writingDone[genCount-1]
		log.Debug().Int("Generation", genCount-1).Msg("Final generation finished writing")
	}

	simEnd := time.Now()
	finalWriteElapsed := simEnd.Sub(generationsEnd)
	totalTimeFinalWrite = finalWriteElapsed.Seconds()
	simElapsed := simEnd.Sub(simStart)
	log.Info().
		Float64("Total Run Time", simElapsed.Seconds()).
		Float64("Total Simulation Time", totalTimeSimulation).
		Float64("Total Link Time", totalTimeLinks).
		Float64("Total Final Write Time", totalTimeFinalWrite).
		Msg("Simulation complete")
}
