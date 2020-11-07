package generation

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
USLEFineSedimentGeneration:
  inputs:
    quickflow: m^3.s^-1
		baseflow: m^3.s^-
		rainfall: mm
		KLSC: '[0,100000000] KLSC'
		KLSC_Fine : '[0,100000000] KLSC'
		CovOrCFact: '[] Average C Factor'
		dayOfYear: dayOfYear
  states:
  parameters:
    S: '[0,5000]mm Mean Summer Rainfall'
    P: '[0,5000]mm Mean Annual Rainfall'
    RainThreshold: '[0,12.7]mm R Factor Rainfall Threshold'
		Alpha: ''
		Beta: '[0.1,10] Monthly EI30 Parameter'
		Eta: '[0.1,10] Monthly EI30 Parameter'
		A1: '[0.001,10] Alpha term 1'
		A2: '[0.001,10] Alpha term 2'
		A3: '[0.001,100] Alpha term 3'
		DWC: '[0.1,10000] Dry Weather Concentration'
		avK: ''
		avLS: ''
		avFines: ''
		area: '[0,]m^2 Modelled area'
		maxConc: '[0,10000]mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration'
		usleHSDRFine: '[0,100]% Hillslope Fine Sediment Delivery Ratio'
		usleHSDRCoarse: '[0,100]% Hillslope Coarse Sediment Delivery Ratio'
		timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds, default=86400'
	outputs:
		quickLoadFine: kg
		slowLoadFine: kg
		quickLoadCoarse: kg
		slowLoadCoarse: kg
		totalLoad: kg
		generatedLoadFine: kg
		generatedLoadCoarse: kg
	implementation:
		function: usleFine
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
		sediment
*/

func usleFine(quickflow, slowflow, rainfall, klsc, klscFine, covOrCFact, dayOfYear data.ND1Float64,
	s, p, rainThreshold, alpha, beta, eta,
	a1, a2, a3, dwc, avK, avLS, avFines, area, maxConc,
	usleHSDRFine, usleHSDRCoarse, timeStepInSeconds float64,
	quickLoadFine, slowLoadFine,
	quickLoadCoarse, slowLoadCoarse,
	totalLoad, generatedLoadFine,
	generatedLoadCoarse data.ND1Float64) {
	n := quickflow.Len1()

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day
		doy := dayOfYear.Get(idx)
		rain := rainfall.Get(idx)
		qf := quickflow.Get(idx)
		sf := slowflow.Get(idx)
		cFactor := covOrCFact.Get(idx)

		loadS := dwc * sf * units.MG_PER_LITRE_TO_KG_PER_M3

		loadQ := 0.0

		theKLSCval := klsc.Get(idx)
		theKLSCClayval := klscFine.Get(idx)
		useAvModel := false
		//Use averaged data first if needed
		if useAvModel {
			//double cFactor = CFact;
			//if (cFactorDataType != Tools.ListBoxOptions.CoverAnalysisMethod.CFACTOR)
			//{
			//    cFactor = Tools.ToolsModel.cFactorFromCover(cFactorDataType, CFact);
			//}

			theKLSCval = avK * avLS * cFactor // cFactor;
			theKLSCClayval = theKLSCval * (avFines / 100)
		}
		//Must both have valid entries, indicating that they are being read from a valid TimeSeries in reflection
		//else if (KLSC > 0 && KLSC_Clay > 0)
		//{
		//    theKLSCval = KLSC;
		//    theKLSCClayval = KLSC_Clay;
		//}

		//This scanlon ToY term is used in GRASP, but they use DoY + 30 to 'peak' in mid Dec
		//This implementation peaks in mid Jan
		//Scanlan et al 1996. Run-off and soil movement on mid-slopes... Rangelands journal.
		scanlon_ToY_Term := math.Cos(2 * math.Pi * (doy - 15) / 365)

		//Latest calibrations have yielded baseflow ratios of 95%+, so we need to consider all flows as 'runoff'
		//		combinedRunoffM3perSec := qf + sf
		// baseflowRat := 0.
		// if combinedRunoffM3perSec > 0 {
		// 	baseflowRat = sf / combinedRunoffM3perSec
		// }

		//Rob removed the check on quickflow here, as Cameron wants to access the
		//timeseries of R factor, and rightly it should get updated regardless of runoff
		R := 0.
		if rain > rainThreshold {

			R = alpha * (1 + eta*scanlon_ToY_Term) * math.Pow(rain, beta)
		}

		USLE_soilEroded_Tons_per_Ha_per_Day_Total := R * theKLSCval

		//getting rid of 'max load' concept, May 2016, reverting to fines concentration in runoff (pre-delivery)
		////Need to alter the KLSC * clay val as well, maybe we shouldn't have created 2 timeseries...
		//double origVal = R * theKLSCval;
		//double maxRateProp = 1;

		//double checkerVal = Max_Conc / 365.25;

		//if (USLE_soilEroded_Tons_per_Ha_per_Day_Total > checkerVal)
		//{
		//    USLE_soilEroded_Tons_per_Ha_per_Day_Total = checkerVal;

		//    if (origVal > 0)
		//    {
		//        maxRateProp = USLE_soilEroded_Tons_per_Ha_per_Day_Total / origVal;
		//    }
		//}

		//USLE_soilEroded_Tons_per_Ha_per_Day_Fine = R * theKLSCClayval * maxRateProp;
		//USLE_soilEroded_Tons_per_Ha_per_Day_Coarse = USLE_soilEroded_Tons_per_Ha_per_Day_Total - USLE_soilEroded_Tons_per_Ha_per_Day_Fine;

		USLE_soilEroded_Tons_per_Ha_per_Day_Fine := R * theKLSCClayval
		USLE_soilEroded_Tons_per_Ha_per_Day_Coarse := USLE_soilEroded_Tons_per_Ha_per_Day_Total - USLE_soilEroded_Tons_per_Ha_per_Day_Fine

		//Create alternatives, we want to keep the USLE-derived values for checking
		theRateForAssignmentTotal := USLE_soilEroded_Tons_per_Ha_per_Day_Total
		theRateForAssignmentFine := USLE_soilEroded_Tons_per_Ha_per_Day_Fine
		theRateForAssignmentCoarse := USLE_soilEroded_Tons_per_Ha_per_Day_Coarse

		//USLE_Daily_Load_kg_Total := 0
		USLE_Daily_Load_kg_Fine := 0.
		USLE_Daily_Load_kg_Coarse := 0.

		USLE_Daily_Load_kg_after_HSDR_applied_Fine := 0.
		USLE_Daily_Load_kg_after_HSDR_applied_Coarse := 0.

		//Call this an 'event checker'
		if qf > 0 && USLE_soilEroded_Tons_per_Ha_per_Day_Total > 0 {

			//this sets a Maximum sediment Concentration value - fudge for when you get large load with a small flow
			// In reality probably shouldn't happen but it does on the rare occasion - due to sediment generation
			//being based on rainfall not runoff

			//currentTotalSedMassKg := USLE_soilEroded_Tons_per_Ha_per_Day_Total * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG
			currentFineSedMassKg := USLE_soilEroded_Tons_per_Ha_per_Day_Fine * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG
			//currentCoarseSedMassKg := USLE_soilEroded_Tons_per_Ha_per_Day_Coarse * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG

			Sediment_Conc_mg_per_L_Fine := (currentFineSedMassKg * units.KG_TO_MILLIGRAM) / (qf * units.CUBIC_METRES_PER_SECOND_TO_MEGA_LITRES_PER_DAY * units.MEGA_LITRES_TO_LITRES)
			if Sediment_Conc_mg_per_L_Fine > maxConc {
				allowedFineSedMassKg := maxConc * (qf * units.CUBIC_METRES_PER_SECOND_TO_MEGA_LITRES_PER_DAY * units.MEGA_LITRES_TO_LITRES) / units.KG_TO_MILLIGRAM

				concPropAdj := allowedFineSedMassKg / currentFineSedMassKg

				theRateForAssignmentTotal *= concPropAdj
				theRateForAssignmentFine *= concPropAdj
				theRateForAssignmentCoarse *= concPropAdj

			}

			//*************Rob moved this stuff here from outside of the 'event checker' above
			//USLE_Daily_Load_kg_Total = theRateForAssignmentTotal * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG // Was TON_TO_KG
			USLE_Daily_Load_kg_Fine = theRateForAssignmentFine * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG
			USLE_Daily_Load_kg_Coarse = theRateForAssignmentCoarse * area * units.SQUARE_METRES_TO_HECTARES * units.TONNES_TO_KG

			USLE_Daily_Load_kg_after_HSDR_applied_Fine = USLE_Daily_Load_kg_Fine * (usleHSDRFine * 0.01)
			USLE_Daily_Load_kg_after_HSDR_applied_Coarse = USLE_Daily_Load_kg_Coarse * (usleHSDRCoarse * 0.01)

			//how do we handle this concentration in 'combined' runoff?
			//USLE_Daily_Sediment_Conc_mg_per_L := (USLE_Daily_Load_kg_after_HSDR_applied_Fine * units.KG_TO_MILLIGRAM) / (qf * units.CUBIC_METRES_PER_SECOND_TO_MEGA_LITRES_PER_DAY * units.MEGA_LITRES_TO_LITRES)
			//USLE_Daily_Sediment_Conc_mg_per_L = (USLE_Daily_Load_kg_after_HSDR_applied_Fine * KG_TO_MILLIGRAM) / (combinedRunoffM3perSec * units.CUBIC_METRES_PER_SECOND_TO_MEGA_LITRES_PER_DAY * MEGA_LITRES_TO_LITRES);

			//ultimately the quickflowConstituent and slowflowConstituent will be re-assigned if this model belongs to a 'wrapper'
			loadQ = USLE_Daily_Load_kg_after_HSDR_applied_Fine / timeStepInSeconds
			//quickflowConstituent = (1 - baseflowRat) * (USLE_Daily_Load_kg_after_HSDR_applied_Fine / timeStepInSeconds);

			//Add USLE derived material to DWC derived
			//slowflowConstituent += baseflowRat * (USLE_Daily_Load_kg_after_HSDR_applied_Fine / timeStepInSeconds);
		} else {
			//quickflowConstituent never used in this context, so Rob added the load objects below
			loadQ = 0

			//Is this ever used???
			//USLE_Daily_Sediment_Conc_mg_per_L := 0
		}

		quickLoadFine.Set(idx, loadQ)
		slowLoadFine.Set(idx, loadS)

		quickLoadCoarse.Set(idx, USLE_Daily_Load_kg_after_HSDR_applied_Coarse/timeStepInSeconds)
		slowLoadCoarse.Set(idx, loadS)
		totalLoad.Set(idx, loadQ+loadS)

		generatedLoadFine.Set(idx,USLE_Daily_Load_kg_Fine/timeStepInSeconds)
		generatedLoadCoarse.Set(idx,USLE_Daily_Load_kg_Coarse/timeStepInSeconds)
	}
}
