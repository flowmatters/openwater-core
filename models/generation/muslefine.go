package generation

import (
	"fmt"
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
MUSLEFineSedimentGeneration:
  inputs:
    quickflow: m^3.s^-1
		baseflow: m^3.s^-
		rainfall: mm
		cover: '[0,1] Visual Cover'
		cFactor: '[0,1] Cover Factor'
		month: monthOfYear
		Rcm: '[] Runoff Coefficient Metric'
  states:
  parameters:
		gamma: unitless
		cr: 'threshold cover'
	  w: '[0,1] weighting fraction between RUSLE and MUSLE approaches (1==RUSLE, 0=MUSLE)'
		a: ''
		b1:
		b2:
		area: '[0,]m^2 Modelled area'
		latitude: ''
		elevation: ''
		avK: ''
		avLS: ''
		avFines: '% of fine sediment in soil'
		DWC: '[0.1,10000] Dry Weather Concentration'
		maxConc: '[0,10000]mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration'
		usleHSDRFine: '[0,100]% Hillslope Fine Sediment Delivery Ratio'
		usleHSDRCoarse: '[0,100]% Hillslope Coarse Sediment Delivery Ratio'
		timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds, default=86400'
	outputs:
		quickLoadFine: kg.s^-1
		slowLoadFine: kg.s^-1
		quickLoadCoarse: kg.s^-1
		slowLoadCoarse: kg.s^-1
		totalFineLoad: kg.s^-1
		totalCoarseLoad: kg.s^-1
		generatedLoadFine: kg
		generatedLoadCoarse: kg
	implementation:
		function: musleFine
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
		sediment
*/

func ei30(rainfall, monthOfYear data.ND1Float64, latitude, elevation float64) data.ND1Float64 {
	f := 1.0 / 12.0
	w := math.Pi / 6.0
	beta := 1.534
	alpha := 0.652 * (1.794 + 0.03063*latitude - 0.0002859*elevation) // PY: 0.75155
	eta := -0.238 - 0.0232*latitude                                   // PY: 0.29
	idx := []int{0}
	n := rainfall.Len1()
	result := data.NewArray1DFloat64(n)
	for day := 0; day < n; day++ {
		idx[0] = day
		p := rainfall.Get(idx)
		j := monthOfYear.Get(idx)
		ei30 := alpha * (1 + eta*math.Cos(2*math.Pi*f*j-w)) * math.Pow(p, beta)
		result.Set(idx, ei30)
	}
	return result
}

func rFactor(rainfallMM, runoffMM,
	monthOfYear data.ND1Float64,
	w,
	a, // PY: 89.45
	b1, // PY: 0.56
	b2, // PY: 0.56
	latitude, elevation, area float64) data.ND1Float64 {
	n := rainfallMM.Len1()
	result := data.NewArray1DFloat64(n)

	ei30 := ei30(rainfallMM, monthOfYear, latitude, elevation)

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day
		rain := rainfallMM.Get(idx)
		QmmPerDay := runoffMM.Get(idx) // mm
		// QmmPerDay := (Q / area) * units.METRES_TO_MILLIMETRES * units.SECONDS_PER_DAY
		areaHa := area * units.SQUARE_METRES_TO_HECTARES
		alpha := 0.7565 * math.Pow(areaHa, -0.256) // PY: 0.0609
		EI := ei30.Get(idx)
		eps := 0.3159 + 0.004461*latitude // Py: ED
		E := eps * rain
		I30 := EI / E
		qPeak := alpha * I30 * QmmPerDay / rain

		rf := w*EI + (1-w)*a*math.Pow(QmmPerDay, b1)*math.Pow(qPeak, b2) // PY: w=0?
		result.Set(idx, rf)
	}

	return result
}

func effectiveRunoff(rcm, p, cover data.ND1Float64, gamma, cr float64) data.ND1Float64 {
	n := rcm.Len1()
	q := data.NewArray1DFloat64(n)

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day
		cov := cover.Get(idx)
		cov = math.Min(cov, cr)

		rc := rcm.Get(idx) * math.Exp(-gamma*cov)

		q.Set(idx, rc*p.Get(idx))
	}
	return q
}

func musleFine(quickflow, slowflow, rainfall, visualCover, cFactor, monthOfYear, rcm data.ND1Float64,
	gamma, cr, // Runoff scaling
	w, a, b1, b2, // R factor
	area, latitude, elevation, // Characteristics
	avK, avLS, avFines, // USLE
	dwc, maxConc, // Concentrations
	usleHSDRFine, usleHSDRCoarse, // Delivery
	timeStepInSeconds float64,
	quickLoadFine, slowLoadFine,
	quickLoadCoarse, slowLoadCoarse,
	totalFineLoad, totalCoarseLoad, generatedLoadFine,
	generatedLoadCoarse data.ND1Float64) {

	runoff := effectiveRunoff(rcm, rainfall, visualCover, gamma, cr)
	rFactor := rFactor(rainfall, runoff, monthOfYear, w, a, b1, b2, latitude, elevation, area)

	coreUSLE(rFactor, quickflow, slowflow, cFactor, area, avK, avLS, avFines,
		dwc, maxConc, usleHSDRFine, usleHSDRCoarse, timeStepInSeconds,
		quickLoadFine, slowLoadFine, quickLoadCoarse, slowLoadCoarse,
		totalFineLoad, totalCoarseLoad, generatedLoadFine, generatedLoadCoarse)
}

/*OW-SPEC
CoreUSLE:
	inputs:
		rFactor: unitless
		quickflow: m^3.s^-1
		baseflow: m^3.s^-
		cFactor: '[0,1] Cover Factor'
	states:
	parameters:
		area: '[0,]m^2 Modelled area'
		avK: ''
		avLS: ''
		avFines: '% of fine sediment in soil'
		DWC: '[0.1,10000] Dry Weather Concentration'
		maxConc: '[0,10000]mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration'
		usleHSDRFine: '[0,100]% Hillslope Fine Sediment Delivery Ratio'
		usleHSDRCoarse: '[0,100]% Hillslope Coarse Sediment Delivery Ratio'
		timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds, default=86400'
	outputs:
		quickLoadFine: kg.s^-1
		slowLoadFine: kg.s^-1
		quickLoadCoarse: kg.s^-1
		slowLoadCoarse: kg.s^-1
		totalFineLoad: kg.s^-1
		totalCoarseLoad: kg.s^-1
		generatedLoadFine: kg
		generatedLoadCoarse: kg
	implementation:
		function: coreUSLE
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
		sediment
*/

func coreUSLE(rFactor, quickflow, slowflow, cFactor data.ND1Float64,
	area, // Area characteristics
	avK, avLS, avFines, // USLE
	dwc, maxConc, // Concentrations
	usleHSDRFine, usleHSDRCoarse, // Delivery
	timeStepInSeconds float64,
	quickLoadFine, slowLoadFine,
	quickLoadCoarse, slowLoadCoarse,
	totalFineLoad, totalCoarseLoad, generatedLoadFine,
	generatedLoadCoarse data.ND1Float64) {

	t_per_ha_to_kg := func(t_per_ha float64) float64 {
		return t_per_ha * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG
	}

	n := quickflow.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day

		R := rFactor.Get(idx)
		cF := cFactor.Get(idx)
		klsc := avK * avLS * cF
		theKLSCClayval := klsc * (avFines / 100)

		USLE_soilEroded_Tons_per_Ha_per_Day_Total := R * klsc
		USLE_soilEroded_Tons_per_Ha_per_Day_Fine := R * theKLSCClayval
		USLE_soilEroded_Tons_per_Ha_per_Day_Coarse := USLE_soilEroded_Tons_per_Ha_per_Day_Total - USLE_soilEroded_Tons_per_Ha_per_Day_Fine

		theRateForAssignmentFine := USLE_soilEroded_Tons_per_Ha_per_Day_Fine
		theRateForAssignmentCoarse := USLE_soilEroded_Tons_per_Ha_per_Day_Coarse

		USLE_Daily_Load_kg_Fine := 0.
		USLE_Daily_Load_kg_Coarse := 0.

		USLE_Daily_Load_kg_after_HSDR_applied_Fine := 0.
		USLE_Daily_Load_kg_after_HSDR_applied_Coarse := 0.

		currentFineSedMassKg := t_per_ha_to_kg(USLE_soilEroded_Tons_per_Ha_per_Day_Fine)

		vol := quickflow.Get(idx) * timeStepInSeconds
		volL := vol * units.CUBIC_METRES_TO_LITRES
		Sediment_Conc_mg_per_L_Fine := (currentFineSedMassKg * units.KG_TO_MILLIGRAM) / volL
		if Sediment_Conc_mg_per_L_Fine > maxConc {
			allowedFineSedMassKg := maxConc * volL / units.KG_TO_MILLIGRAM

			concPropAdj := allowedFineSedMassKg / currentFineSedMassKg

			theRateForAssignmentFine *= concPropAdj
			theRateForAssignmentCoarse *= concPropAdj
		}

		USLE_Daily_Load_kg_Fine = t_per_ha_to_kg(theRateForAssignmentFine)
		USLE_Daily_Load_kg_Coarse = t_per_ha_to_kg(theRateForAssignmentCoarse)

		USLE_Daily_Load_kg_after_HSDR_applied_Fine = USLE_Daily_Load_kg_Fine * (usleHSDRFine * 0.01)
		USLE_Daily_Load_kg_after_HSDR_applied_Coarse = USLE_Daily_Load_kg_Coarse * (usleHSDRCoarse * 0.01)

		loadQ := USLE_Daily_Load_kg_after_HSDR_applied_Fine / timeStepInSeconds
		loadS := dwc * slowflow.Get(idx) * units.MG_PER_LITRE_TO_KG_PER_M3

		quickLoadFine.Set(idx, loadQ)
		slowLoadFine.Set(idx, loadS)
		totalFineLoad.Set(idx, loadQ+loadS)

		coarseQuick := USLE_Daily_Load_kg_after_HSDR_applied_Coarse / timeStepInSeconds

		quickLoadCoarse.Set(idx, coarseQuick)
		slowLoadCoarse.Set(idx, 0.0) // SURELY INCORRECT
		totalCoarseLoad.Set(idx, coarseQuick+0.0)

		generatedLoadFine.Set(idx, USLE_Daily_Load_kg_Fine/timeStepInSeconds)
		generatedLoadCoarse.Set(idx, USLE_Daily_Load_kg_Coarse/timeStepInSeconds)
	}
}

/*OW-SPEC
MUSLECoverMetric:
	  inputs:
			P: mm
			cover: '[0,1]unitless'
	  states:
	  parameters:
			area: '[0,]m^2 Modelled area'
			gamma: unitless
			cr: 'threshold cover'
		outputs:
			cover_metric: unitless
		implementation:
			function: musleCover
			type: scalar
			lang: go
			outputs: params
		init:
			zero: true
		tags:
			constituent generation
			sediment
*/

func musleCover(p, cover data.ND1Float64, area, gamma, cr float64, cover_metric data.ND1Float64) {
	// 1 per FU
	n := p.Len1()

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day
		precip := p.Get(idx)
		cov := cover.Get(idx)
		cov = math.Min(cov, cr)

		metric := precip * math.Exp(-gamma*cov) * area
		cover_metric.Set(idx, metric)
	}
}

/*OW-SPEC
MUSLEEventBasedRFactor:
	  inputs:
			rainfall: mm.day^-1
			totalFlow: mm.day^-1
			month:
	  states:
	  parameters:
		  alpha: 'unitless, default=0.75155'
			eta: 'unitless, default=0.29'
			scalingFactor: 'unitless, default=0.0609'
			a: 'unitless, default=89.45'
			b1: 'unitless, default=0.56'
			b2: 'unitless, default=0.56'
			timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds, default=86400'
		outputs:
			R: unitless
			debugInEvent: unitless
		implementation:
			function: musleEventBasedRFactor
			type: scalar
			lang: go
			outputs: params
		init:
			zero: true
		tags:
			constituent generation
			sediment
*/

func musleEventBasedRFactor(rainfall, totalFlow, monthOfYear data.ND1Float64,
	alpha, eta, scalingFactor, a, b1, b2, timeStepInSeconds float64,
	rFactor, debugInEvent data.ND1Float64) {
	const ED = 0.205
	const beta_o = 1.49

	n := rainfall.Len1()

	idx := []int{0}
	inEvent := func() bool {
		return (rainfall.Get(idx) > 0) || (totalFlow.Get(idx) > 0)
	}

	idx[0] = 0
	if inEvent() {
		fmt.Println("MUSLEEventBasedRFactor: First day has non-zero values")
		panic("First day has non-zero values")
	}

	idx[0] = n - 1
	if inEvent() {
		fmt.Println("MUSLEEventBasedRFactor: Last day has non-zero values")
		panic("Last day has non-zero values")
	}

	midEvent := false
	startEvent := -1
	totalEventFlow := 0.0
	totalEventRain := 0.0
	maxEventRain := 0.0
	maxEventFlow := 0.0
	// eventEI30 := 0.0
	eventI30Max := 0.0

	resetEvent := func() {
		totalEventFlow = 0.0
		totalEventRain = 0.0
		maxEventRain = 0.0
		maxEventFlow = 0.0
		eventI30Max = 0.0
		midEvent = false
	}

	for day := 0; day < n; day++ {
		idx[0] = day
		currentlyInEvent := inEvent()
		if currentlyInEvent {
			debugInEvent.Set(idx, 1.0)
		} else {
			debugInEvent.Set(idx, 0.0)
		}

		pm := rainfall.Get(idx)
		// flowPerSecond := totalFlow.Get(idx)
		// vol := flowPerSecond * timeStepInSeconds
		vol := totalFlow.Get(idx)

		if pm > maxEventRain {
			mon := monthOfYear.Get(idx)
			fac := alpha * (1 + eta*math.Cos(math.Pi*(mon-1)/6.))
			eventI30Max = fac * math.Pow(pm, beta_o-1) / ED
		}
		maxEventRain = math.Max(maxEventRain, pm)
		maxEventFlow = math.Max(maxEventFlow, vol)
		totalEventRain += pm
		totalEventFlow += vol

		rFactor.Set(idx, 0)

		eventJustEnded := midEvent && !currentlyInEvent
		eventJustStarted := currentlyInEvent && !midEvent

		if eventJustEnded {
			endEvent := day - 1
			if totalEventFlow == 0.0 {
				resetEvent()
				continue
			}

			Qp := scalingFactor * eventI30Max * (maxEventFlow / maxEventRain)
			eventQQp := a * math.Pow(totalEventFlow, b1) * math.Pow(Qp, b2)
			unitR := eventQQp / totalEventFlow

			for replayDay := startEvent; replayDay <= endEvent; replayDay++ {
				idx[0] = replayDay
				dayVol := totalFlow.Get(idx)
				rFactor.Set(idx, dayVol*unitR)
			}

			resetEvent()
		} else if eventJustStarted {
			startEvent = day
		}

		midEvent = currentlyInEvent

	}
}
