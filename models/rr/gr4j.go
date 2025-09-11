package rr

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
GR4J:
  inputs:
    rainfall: mm
    pet: mm
  states:
		s:
		r:
    n1:
    n2:
    q1:
    q9:
  parameters:
    X1: '[1,1500]mm Capacity of the production soil (SMA) store'
    X2: '[-10,5.0]mm Water exchange coefficient'
    X3: '[1,500]mm Capacity of the routing store'
    X4: '[0.5,4.0]days Time parameter for unit hydrographs'
	outputs:
		runoff: mm
	implementation:
		function: gr4j
		type: scalar
		lang: go
		outputs: params
	init:
		function: initGR4J
		type: scalar
		lang: go
	extractstates:
		function: extractGR4JStates
		packfunc: packGR4JStates
		type: scalar
		lang: go
	tags:
		rainfall runoff
*/

func initGR4J(x1 float64, x2 float64, x3 float64, x4 float64) data.ND2[float64] {
	// Initialise states
	// * S
	// * nUH1
	// * nUH2

	// * q1 (nUH2)
	// * q9 (nUH1)

	var n1 = int(math.Ceil(x4))
	var n2 = int(math.Ceil(2 * x4))
	q1 := make([]float64, n2)
	q9 := make([]float64, n1)

	result := packGR4JStates(0.0, 0.0, n1, n2, q1, q9)
	return result
}

func extractGR4JStates(states []float64) (float64, float64, int, int, []float64, []float64) {
	s := states[0]
	r := states[1]
	n1 := int(states[2])
	n2 := int(states[3])
	q1 := states[4:]
	q9 := states[4+n2:]
	return s, r, n1, n2, q1, q9
}

func packGR4JStates(s, r float64, n1, n2 int, q1, q9 []float64) data.ND2[float64] {
	result := data.NewArray2D[float64](1, 4+n1+n2)
	//result := make(sim.StateSet, 3+n1+n2)
	result.Set2(0, 0, s)
	result.Set2(0, 1, s)
	result.Set2(0, 2, float64(n1))
	result.Set2(0, 3, float64(n2))

	result.Apply([]int{0, 4}, 1, 1, q1)
	result.Apply([]int{0, 4 + n2}, 1, 1, q9)
	return result
}

func gr4j(rainfall []float64, pet []float64, s0 float64, r0 float64,
	n1 int, n2 int, q1State []float64, q9State []float64,
	x1 float64, x2 float64, x3 float64, x4 float64, runoff []float64) (float64, float64, int, int, []float64, []float64) {
	nDays := len(rainfall)
	//var runoff []float64 = data.NewArray1D(nDays)
	//var q9This []float64;
	//var q9Last []float64;
	//var q1This []float64;
	//var q1Last []float64;
	var S = s0
	var Ps float64
	var Es float64
	var Pr float64
	var R = r0

	var SH1 []float64 = make([]float64, n1)
	var i = 0
	for i := 0; i < n1; i++ {
		SH1[i] = math.Pow((float64)(i+1)/x4, 5.0/2.0)
	}
	SH1[n1-1] = 1.0

	var UH1 = make([]float64, n1)
	UH1[0] = SH1[0]
	for i := 1; i < n1; i++ {
		UH1[i] = SH1[i] - SH1[i-1]
	}

	var SH2 = make([]float64, n2)
	for i := 0; i <= int(x4-1); i++ {
		SH2[i] = 0.5 * math.Pow((float64)(i+1)/x4, 5.0/2.0)
	}
	i++
	for ; i < n2; i++ {
		SH2[i] = 1 - 0.5*math.Pow(2-(float64)(i+1)/x4, 5.0/2.0)
	}
	SH2[n2-1] = 1.0
	UH2 := make([]float64, n2)
	UH2[0] = SH2[0]
	for i := 1; i < n2; i++ {
		UH2[i] = SH2[i] - SH2[i-1]
	}

	// Used in mass balance
	//  var Sprev float64;
	//  var Rprev float64;
	//  var ech1 float64;
	//  var ech2 float64;
	var Perc float64
	//if x4 < 0.1 { // x4 set to be large number if optimisation brings it close to zero
	//  x4 = 10;
	//}

	//q9This = make([]float64, n1);
	//q9Last = make([]float64, n1);
	//
	//q1This = make([]float64, n2);
	//q1Last = make([]float64, n2);

	for day := 0; day < nDays; day++ {
		//    Sprev = S;
		//    Rprev = R;
		//    ech1 = 0.0;
		//    ech2 = 0.0;
		Ps = 0.0
		Es = 0.0
		Pr = 0.0
		var netRainfall float64 = 0.0
		var netET float64 = 0.0
		Perc = 0.0
		var Q1 float64 = 0.0
		var Q9 float64 = 0.0
		var Tp float64 = 0.0
		var Qd float64 = 0.0
		var Qr float64 = 0.0
		var ech float64 = 0.0
		var todaysRainfall float64 = rainfall[i]
		var todaysPET float64 = pet[i]
		//----------------Production-------------------------
		var ws float64 = 0
		if todaysRainfall > todaysPET {
			netRainfall = todaysRainfall - todaysPET
			ws = netRainfall / x1
			if ws > 13.0 {
				ws = 13.0
			}

			Ps = (x1 * (1 - math.Pow(S/x1, 2.0)) * math.Tanh(ws)) /
				(1.0 + (S/x1)*math.Tanh(ws))
			Pr = netRainfall - Ps
			netET = 0
		} else {
			netET = todaysPET - todaysRainfall
			//float64
			ws = netET / x1
			if ws > 13.0 {
				ws = 13.0
			}
			tws := math.Tanh(ws)
			Es = (S * (2 - S/x1) * tws) /
				(1 + (1-S/x1)*tws)

			//             Es = (S * (2 - S / x1) * Tanh(ws)) /
			//                  (1 + (1 - S / x1) * Tanh(ws));
			Pr = 0.0
		}
		S = S - Es + Ps
		Perc = S * (1 - math.Pow((1+
			math.Pow((4.0/9.0)*(S/x1), 4.0)), -0.25))
		S = S - Perc
		Pr = Perc + Pr
		for i := 0; i < n1; i++ {
			q9State[i] = q9State[i] + (Pr * 0.9 * UH1[i])
		}
		for i := 0; i < n2; i++ { // n2?
			q1State[i] = q1State[i] + (Pr * 0.1 * UH2[i])
		}
		Q9 = q9State[0]
		Q1 = q1State[0]
		for i := 1; i < n1; i++ {
			q9State[i-1] = q9State[i]
		}
		q9State[n1-1] = 0.0

		for i := 1; i < n2; i++ {
			q1State[i-1] = q1State[i]
		}
		q1State[n2-1] = 0.0

		ech = x2 * math.Pow(R/x3, 7.0/2.0)
		//Routing store calculation
		R = R + Q9 + ech

		if R < 0 {
			R = 0
		}
		//Case where reservoir content is not sufficient
		//    ech1 = -R - Q9;

		//outflow of routing reservoir
		Qr = (R - R/math.Pow(1+math.Pow(R/x3, 4.0), 0.25))
		R = R - Qr

		//Direct runoff calculation
		Qd = 0.0

		//Case where the UH cannot provide enough water
		Tp = Q1 + ech
		//    ech2 = -Q1;
		if Tp > 0 {
			Qd = Q1 + ech
			//      ech2 = ech;
		}

		qtot := Qr + Qd
		runoff[day] = qtot

	}

	return S, R, n1, n2, q1State, q9State
}
