// Package reportingprobe is a codegen fixture, not a real model.
//
// It lives under testdata so the go tool ignores it: nothing imports it, so it
// never registers in sim.Catalog and the model signature hash is unaffected.
// TestGeneratedReportingCodeCompiles builds it explicitly.
package reportingprobe

import "github.com/flowmatters/openwater-core/sim"

/*OW-SPEC
ReportingProbe:
  symbol: RP
  inputs:
    inflow: m^3.s^-1
  states:
    total:
  parameters:
    threshold: '[0,1000] Value above which a report is raised, default=10'
  outputs:
    outflow: m^3.s^-1
  implementation:
    function: probeKernel
    type: scalar
    lang: go
    outputs: params
  init:
    zero: true
    lang: go
  reports:
    started: '[warning] kernel started'
    aboveThreshold: '[warning] inflow {0} exceeds threshold {1}'
    negativeInflow: '[error] inflow {0} at step, clamped to zero (was {0}, limit {1}, spare {2})'
*/

func probeKernel(inflow []float64, total float64, threshold float64,
	outflow []float64, rep *sim.Reporter) float64 {
	rep.Warn0(ReportingProbeReportStarted, 0)

	for i := range inflow {
		v := inflow[i]

		if v < 0 {
			rep.Err3(ReportingProbeReportNegativeInflow, int32(i), v, 0, 0)
			v = 0
		}

		if v > threshold {
			rep.Warn2(ReportingProbeReportAboveThreshold, int32(i), v, threshold)
		}

		total += v
		outflow[i] = v
	}

	return total
}
