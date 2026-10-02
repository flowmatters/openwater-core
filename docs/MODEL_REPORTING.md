# Model reporting

Status: complete (phases 0-6).

## The problem

Model engines log almost nothing beyond fatal errors. That is partly for
performance, but mostly because a kernel has no idea what it is modelling. A
Simhyd run doesn't know which subcatchment or functional unit it represents --
and those concepts are specific to how a user assembled their model. It doesn't
know the date of timestep 4,102 either.

So the kernel can't produce a message a user could act on. What it *can* say is
"this happened, at timestep N, with these numbers". Everything else -- which
place, which day, what wording -- is added afterwards by code that does have the
model file's metadata.

The models are already asking for this:

- `models/storage/storage.go:21` defines `MAX_SUBTIMESTEPS = 600000` and line
  202 increments `nSubtimeSteps`, but nothing ever checks it.
- `models/routing/storage_routing.go:229` is an empty `if math.Abs(delta) >
  massBalanceLimit { }` -- someone wanted to report a mass-balance failure and
  had nowhere to put it.
- `checkStorageConfiguration` failing (`storage.go:132`) logs once, without cell
  identity, then returns zeros for the entire run. A silent wrong answer.

## Design

### Opt in through OW-SPEC

A model participates by declaring a `reports:` block, in the bracketed style the
parameter grammar already uses:

```yaml
Storage:
  ...
  reports:
    badConfiguration: '[error] LVA table invalid -- {0} points, max volume {1} m^3'
    subtimestepFloor: '[error] sub-timestep floor ({0} s) reached with volume {1} m^3'
    releaseNotConverged: '[warning] release rate not converged -- est {0}, avg {1} m^3.s^-1'
    subtimestepBudget: '[warning] {0} sub-timesteps exceeds budget {1}'
```

The block declares the opt-in, the severity of each code, and the message
template, all in one place. Report templates must be **single-quoted**: they are
free-form English, and an unquoted `severity: message` form would require the
exact colon-space sequence YAML reserves. `ow-specgen` validates that each value
parsed as a string and fails the build if not.

Not every mistake reaches that check: `badConfig: error: LVA table invalid` and
`badConfig: {0} points seen` are both invalid *YAML*, so the spec block never
parses at all. That is only safe because phase 0 made `ow-specgen` exit
non-zero -- before it, such a file silently generated nothing and the build
carried on with a stale wrapper.

Codegen emits, for a model that opts in:

```go
const (
    StorageReportBadConfiguration    int32 = 0
    StorageReportSubtimestepFloor    int32 = 1
    StorageReportReleaseNotConverged int32 = 2
)

var storageReportCatalog = []sim.ReportDescription{...}

func (m *Storage) SetReportSink(s *sim.ReportSink)        { m.reports = s }
func (m *Storage) ReportCatalog() []sim.ReportDescription { return storageReportCatalog }
```

and inside each per-cell goroutine in `Run`, `rep := m.reports.Cell(i)` passed
as the kernel's final argument. Every one of the 42 existing wrappers
regenerates byte-identically, because none declares a `reports:` block yet.

Reporting is **not** part of `ModelSignature` (`util/version.go:121`), which
covers only inputs, outputs, states, parameters and dimensions. Adding reports to
a model does not change the signature hash, so existing model files stay
compatible.

### A code and three numbers, not a string

A report is a code plus up to three `float64`s. Message text lives in the
model's catalogue and is rendered only when someone reads the reports.

```go
type Report struct {
    Timestep int32
    Code     int32
    Severity Severity
    NArgs    int32
    Args     [MaxReportArgs]float64
}   // 40 bytes, pointer-free
```

Formatting in a timestep loop would dominate the model's own arithmetic. Fixed
arity rather than `...float64` matters too: a variadic call builds its slice at
the call site, *before* the nil check can short-circuit it.

Severity is limited to `warning` and `error`. At one report per cell per
timestep, anything more verbose floods faster than it informs.

### Threading it through: an optional interface

`sim.TimeSteppingModel` is unchanged. Models that report implement an optional
interface, following the existing `Hotstartable` pattern:

```go
type Reporting interface {
    SetReportSink(*ReportSink)
    ReportCatalog() []ReportDescription
}
```

Models that don't opt in are completely unaffected -- no generated code changes,
no cost.

### Lock-free per cell

Generated `Run` methods spawn one goroutine per cell. The sink pre-allocates one
`Reporter` per cell and each goroutine writes only to its own:

```go
rep := m.reports.Cell(i)   // nil when reporting is off
currentvolume, level, area = storageWaterBalance(..., rep)
```

Single-writer-per-slot, so nothing synchronises. The unbuffered `doneChan`
receive that joins the goroutines is what makes their writes visible to `Flush`.

`Reporter` is padded to two cache lines and per-cell counter blocks are padded to
a whole cache line. This is worth **19%** on the concurrent counter path -- see
Measurements.

### Two tiers: counts always, details capped

- **Counters** are dense, per (cell, code), and complete. One increment. They
  give unbiased coverage: *"link 5123 hit the mass-balance floor on 4,102 of
  10,957 days."*
- **Detail records** carry timesteps and arguments, capped per (cell, code)
  (default 3). They give diagnosable examples without unbounded memory.

The records kept are the *first* ones -- the onset of a condition, not its tail.
Without the cap, a condition firing every timestep costs +74% and 2.4 MB per
cell-run. With it, nothing measurable.

### Storage

Reports go in the **results** file, struct-of-arrays:

```
/REPORTS/signature                 string[1]             engine model signature
/REPORTS/<Model>/run               int32[n]              global run index
/REPORTS/<Model>/timestep          int32[n]              -1 for configuration faults
/REPORTS/<Model>/code              int32[n]
/REPORTS/<Model>/severity          int32[n]
/REPORTS/<Model>/args              float64[n,3]
/REPORTS/<Model>/counts            int32[nCells,nCodes]  complete, cell-major
/REPORTS/<Model>/names             string[nCodes]        the model's catalogue
/REPORTS/<Model>/messages          string[nCodes]
/REPORTS/<Model>/code_severity     int32[nCodes]
```

`counts` is written even for a model that reported nothing, so a reader can
tell "checked, all clear" from "never ran". The detail datasets are omitted
when there is nothing to put in them.

SoA because `io.H5Ref[T]` is already generic over `data.Number`; a compound type
would need new code in `io/hdf5.go`. It also compresses better.

**Buffer for the whole run and write once at the end, at exact final shape.**
There is no append: `io/hdf5_util.go:247` creates dataspaces with `nil` maxdims
(fixed extent), `openOrCreateDataset` refuses to resize, and the pinned
`gonum/hdf5` has **no `Resize`/`SetExtent` binding at all**. Growing a dataset
would mean vendoring a patched binding. Buffering avoids all of it, and makes
output deterministic as a side effect (flatten by cell index, never by goroutine
completion order).

The memory cost is trivial -- see Scale.

The sink buffer must live on `modelReference`, **not** `modelGeneration`:
`cmd/ow-sim/main.go:191,204,217` purge generations aggressively, some from the
writer goroutine, and would destroy reports before they are written.

### Translation to dates and places

Everything needed is already in the model file: `/META/timeperiod`,
`/DIMENSIONS/<dim>`, and `/MODELS/<m>/map`.

- **Python does the real work.** `res.reports()` returns a DataFrame with
  `date`, `timestep`, `model`, `severity`, `code`, `report`, `message`, plus one
  column per dimension of that model (`catchment`, `cu`, ...).
  `res.report_counts()` returns the complete occurrence counts, which is the
  accessor to use for "how often": detail records are capped during the run, so
  `reports()` shows examples and would understate a daily condition.

  Both degrade rather than fail when the model file can't answer: no run map
  means no dimension columns, no time period means no dates, and either way a
  warning is logged. The catalogue is read from the results file rather than the
  installed library, so an old results file stays readable against a newer
  build.
- **ow-sim prints a terse summary** at end of run -- counts by model/code/
  severity, no dimension decoding. A headless run should not be silent.
- **ow-single gets it free**: `singleModelResults.Log []string` already exists in
  `sim/single.go`. One cell, no perf concern, so render fully there.

Run and timestep indices are only meaningful against the originating model file.
`clip.py:291-310` renumbers run indices and `template.py:1526-1529` (`retime`)
rewrites `META/timeperiod`, either of which would silently re-locate or re-date
existing reports. Stamp the model file's `signature_hash` and a run identifier
into `/REPORTS`, and have the Python API refuse on mismatch.

### Defaults

On, at `warning`, detail cap 3. `-reports=off|warning|error` and
`-report-detail-limit=N` on ow-sim. The disabled path costs 0.10 ns/op; there is
no performance argument for shipping it off.

`-reports-fatal=error` (default on) makes ow-sim exit non-zero when any
error-severity report fires. Converting the existing `log.Panic()` sites to
reports otherwise turns a loud crash into a wrong-but-complete run with exit
code 0 -- which is how batch and CI runs currently detect breakage.

## Scale

Real Burdekin model (`BM_RC13_BASE.h5`, 13,514 timesteps):

| model | total cells | generations | max cells in one generation |
|---|---:|---:|---:|
| SednetDissolvedNutrientGeneration | 24,552 | 99 | 1,908 |
| DepthToRate | 22,506 | 99 | 1,626 |
| StorageRouting | 725 | 99 | 96 |
| Storage | 14 | 99 | 1 |

StorageRouting counters: 96 x 8 x 4 = **3 KB**. Widest model: 61 KB. Whole-run
buffering: ~700 KB for StorageRouting, ~24 MB worst case for the widest model,
bounded by the global cap.

Each cell belongs to exactly one generation (`batches` is a cumsum,
`template.py:945`), so per-cell counts are complete when that generation
finishes.

## Measurements

Intel i9-14900K, Go 1.25, medians of 15 runs. Benchmarks live in
`sim/reporting_bench_test.go` and `models/routing/reporting_bench_test.go`.

Per call:

| | ns/op | allocs |
|---|---:|---:|
| Disabled (nil reporter) | 0.10 | 0 |
| Below severity threshold | 0.35 | 0 |
| Counting (past the detail cap) | 1.44 | 0 |
| Recording (uncapped) | 42.84 | 206 B/op |

Whole kernel, `storageRouting`, 13,514 timesteps, against the untouched kernel on
the same inputs:

| | overhead |
|---|---:|
| silent inputs, reporter nil | -4.4% |
| silent inputs, reporter enabled | -4.1% |
| firing every timestep, capped | -2.1% |
| 96 concurrent cells, firing every timestep | +0.7% |
| firing every timestep, uncapped | +74.1% |

The negative figures are code-layout noise, not a speedup: reporting overhead is
below the measurement floor. `-gcflags=-m` confirms no heap traffic in the kernel
region, and all report call sites show `inlining call to sim.(*Reporter).Warn2`.

Cache-line padding, 96 concurrent cells: 213 us packed vs 172 us padded, **-19%**.

Phase 3, the shipped instrumented kernels against a baseline measured on the
same machine immediately before instrumenting (1118.2 us single cell):

| | median | vs baseline |
|---|---:|---:|
| StorageRouting, reporting off | 1064.1 us | -4.8% |
| StorageRouting, reporting on | 1074.0 us | -3.9% |
| Storage, off -> on | 874.8 us | +0.4% |
| 96 concurrent cells, stiff case, off -> on | 236.0 ms | +2.3% |

Again the negative figures are code-layout noise. Both instrumented kernels
allocate nothing, `storage_routing.go` shows no heap escapes at all, and the 13
report call sites all inline to a nil check.

Per generation: 59 us / 369 KB / 3 allocs to build a sink for 1,908 cells x 8
codes; 10.6 us to flush.

## What to instrument

**Storage** (few instances, expensive each -- tests correctness and ergonomics):

- `checkStorageConfiguration` failure -> `error`, per cell. Currently a silent
  all-zeros run.
- Sub-timestep floor with `testVol < 0` -> `error`. Replaces a `log.Panic()` that
  kills the entire simulation for one bad cell.
- Release rate not converged (exit via `MIN_TIMESTEP_SECONDS_POSITIVE`) ->
  `warning`, **filtered**. Reporting every such break produced 2,374 reports
  over 50 timesteps on an ordinary storage: the solver routinely stops just
  outside its convergence tolerance, and that is not a problem. Two filters make
  it meaningful -- at most one report per timestep (the sub-timestep loop runs
  dozens of times per day), and only when the discrepancy exceeds
  `REPORTABLE_REL_ERROR_RELEASE_RATE` (1%, against a convergence tolerance of
  1e-5). Ordinary storages are now silent; a storage whose volume crosses a
  release discontinuity reports on the day it crosses; one pinned to a
  discontinuity reports daily.
- `nSubtimeSteps > MAX_SUBTIMESTEPS` -> `warning`. Activates the dead constant.
- NaN demand/inflow -> `error`, replacing three `log.Panic()` + `fmt.Printf`
  blocks.
- Volume clamped below zero -> `error` (`volumeClamped`).

**StorageRouting** (many instances, cheap each -- tests the performance envelope):

- `|delta| > massBalanceLimit` after `FindRoot` -> `warning`. Fills the empty
  `if` block at line 229.
- NaN inflow/lateral -> `error`.
- ~~Storage clamped at zero~~ -- dropped. There is no clean per-timestep place
  for it: the clamp lives inside `runRouting`, which the root finder calls many
  times per timestep, and the alternatives (a link running dry) are normal
  hydrology rather than something to warn about.

## Phases

0. **Done.** `ow-specgen` exits non-zero on failure; `build/build.sh` clears
   generated wrappers first; the discarded `WriteSubset` error in
   `io/hdf5.go:209`; `inputs.Len(20)`/`states.Len(1)` in `cmd/ows-ensemble`
   (which panicked on every run); two unused imports that stopped
   `models/routing` tests compiling.
1. **Done.** `sim/reporting.go` -- Severity, Report, ReportDescription, Reporter,
   ReportSink, ReportSet, the Reporting interface -- plus tests, race coverage,
   an inlining assertion and the benchmarks above.
2. **Done.** `ow-specgen` parses the `reports:` block into `ReportSpec`s and the
   template emits report code constants, a catalogue, and `SetReportSink`/
   `ReportCatalog` for models that opt in. All 42 existing wrappers regenerate
   byte-identically. Also fixed: the package regex only matched at byte 0, so
   any model file with a licence header or package doc comment was silently
   skipped and generated nothing.
3. **Done.** Storage declares five reports, StorageRouting three. Both kernels
   take a trailing `*sim.Reporter`; reports are additive, and the existing
   `log.Panic()` calls are untouched until phase 6. Also fixed:
   `checkStorageConfiguration` handles `nLVA == 0`, but that branch was
   unreachable because the code indexed `volumes[nLVA-1]` first -- an empty LVA
   table panicked instead of being rejected.
4. **Done.** `ow-sim` builds a sink per generation, collects it onto the
   `modelReference` before the generation can be purged, and writes everything
   once at the end. Flags: `-reports`, `-report-detail-limit`, `-reports-fatal`.
   `ow-single` renders reports into the log it already returns (capped at 20 per
   code, with the total); `ows-ensemble` summarises per ensemble member. Also
   added `H5Ref.WriteText`, the write counterpart to `LoadText`, so the message
   catalogue travels with the results file.
5. **Done.** `OpenwaterResults.reports()` returns one row per detail record --
   date, place, severity, rendered message -- and `report_counts()` returns the
   complete per-cell occurrence counts. Both accept dimension constraints and a
   severity filter, and both refuse to run when the results' signature stamp
   doesn't match the model file (overridable with `check_signature=False`).
   `OpenwaterSplitResults` proxies both, summing counts across splits.
   `ModelGraph.run()` now raises a message pointing at `.reports()` instead of a
   bare `assert proc.returncode==0`.
6. **Done.** Every `log.Panic()` in Storage and StorageRouting is gone, along
   with the `fmt.Printf` blocks beside them, which dumped to stdout without
   saying which cell they belonged to. See Degradation policy below.

## Degradation policy

A model kernel that cannot compute a cell no longer takes down the run. What it
does instead depends on why:

- **Unusable inputs** -- a NaN inflow or demand, a malformed release curve.
  There is nothing to compute from, so the cell is reported at `error` severity
  and the rest of its outputs are filled with **NaN**. NaN rather than zero is
  the whole point: a zero is indistinguishable from a real answer, and a storage
  silently producing zeros for a whole run is the failure this work set out to
  make visible.
- **Bounded numerical failures** -- the sub-timestep solver hitting its floor
  with the volume still negative, or a volume that comes out below zero. The
  value is clamped, the cell keeps running, and an `error` report carries the
  detail.

Cells are independent, so a broken one leaves its neighbours untouched --
`TestOneBadCellDoesNotAffectTheOthers` and `TestOneBadLinkDoesNotAffectTheOthers`
run a whole generation through the generated wrapper to prove it.

The run still fails loudly overall: `-reports-fatal=error` (on by default) makes
ow-sim exit non-zero, and `ModelGraph.run()` raises with a message pointing at
`.reports()`. What changes is that the failure now names the catchment and the
date, the results file is still written, and one bad cell out of 24,000 no
longer costs the whole simulation.

Verified end to end: a single NaN inflow at day 12 of catchment sc1 produces

```
      date  timestep          model severity   report                                            message catchment
2020-01-13        12 StorageRouting    error nanInput inflow nan or lateral 0.5 m^3.s^-1 is not a number       sc1
```

with sc1 NaN from day 12 on, sc0 and sc2 untouched, and ow-sim exiting 1.

## Do not re-add the rm to build.sh

Phase 0 added both an `ow-specgen` exit code and an `rm -f generated_*.go` to
`build/build.sh`. The exit code is the actual guarantee -- with `set -e`, a
failed generation aborts the build before anything compiles. The `rm` was
belt-and-braces, and it turned out to break the build: deleting and recreating
the wrappers immediately before compiling them makes the Go build cache
intermittently compile against the pre-deletion state, and `build.sh` then fails
roughly every second run with spurious `undefined:` errors. Verified by
alternating pass/fail across six consecutive runs with the `rm`, and six clean
runs without it. Orphaned wrappers, from a model file that was renamed or
removed, are cleared by `pre_build.sh`, which does not compile straight after.

## Known gaps

- **Split outputs.** `WriteData` diverts to `writeProtobuf` when
  `mr.OutputProcess != nil`, and `io/protobuf/ow-outputs.proto` carries no
  reports (it already silently drops states). Either extend the proto and
  `cmd/ow-sim/writer.go`, or warn at startup that reports are unavailable for
  split-output models.
- ~~**String datasets from Go.**~~ Done: `H5Ref.WriteText` mirrors `LoadText`
  (fixed-length, NUL-padded), with `openOrCreateParents` in `io/hdf5_util.go`
  creating intermediate groups.
