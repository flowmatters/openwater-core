package generation

// import (
// 	"github.com/flowmatters/openwater-core/conv"
// 	"github.com/flowmatters/openwater-core/data"
// )

// /*OW-NOT-SPEC
// GullyGeneration:
//   inputs:
//     quickFlow: m^3.s^-1
// 		slowFlow: m^3.s^-1
// 		jday: jDay
// 		rainfall: mm
//   states:
// 	parameters:
// 		YearDisturbance: '[0,3000]year Year of disturbance'
// 		YearActivityEnd: '[0,3000]year Year of end of gully activity'
//     S: '[0,5000]mm Mean Summer Rainfall'
//     P: '[0,5000]mm Mean Annual Rainfall'
//     RainThreshold: '[0,12.7]mm R Factor Rainfall Threshold'
// 		Alpha: ''
// 		Beta: '[0.1,10] Monthly EI30 Parameter'
// 		Eta: '[0.1,10] Monthly EI30 Parameter'
// 		A1: '[0.001,10] Alpha term 1'
// 		A2: '[0.001,10] Alpha term 2'
// 		A3: '[0.001,100] Alpha term 3'
// 		DWC: '[0.1,10000] Dry Weather Concentration'
// 		KLSC: '[0,100000000] KLSC'
// 		KLSC_Fine : '[0,100000000] KLSC'
// 		CovOrCFact: '[] Average C Factor'
// 		avK: ''
// 		avLS: ''
// 		avFines: ''
// 		area: '[0,]m^2 Modelled area'
// 		maxConc: '[0,10000]mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration'
// 		usleHSDRFine: '[0,100]% Hillslope Fine Sediment Delivery Ratio'
// 		usleHSDRCoarse: '[0,100]% Hillslope Coarse Sediment Delivery Ratio'
// 		timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds'
// 	outputs:
// 		quickLoadFine: kg
// 		slowLoadFine: kg
// 		quickLoadCoarse: kg
// 		slowLoadCoarse: kg
// 		totalLoad: kg
// 	implementation:
// 		function: usleFine
// 		type: scalar
// 		lang: go
// 		outputs: params
// 	init:
// 		zero: true
// 	tags:
// 		constituent generation
// 		sediment
// */

// func gully(quickflow, slowflow, jday data.ND1Float64,
// 	timeStepInSeconds, area, yearDisturbance, yearActivityEnd float64) {
// 	nDays := quickflow.Len1()

// 	for i := 0; i < nDays; i++ {
// 		now := conv.JtoDate(jday.Get1(i))
// 		year := float64(now.Year())
// 		Daily_Runoff := quickflow.Get1(i) //these in cumecs
// 		Daily_Baseflow := slowflow.Get1(i)

// 		Gully_Daily_Load_kg_Fine = 0
// 		Gully_Daily_Load_kg_Coarse = 0
// 		Gully_Daily_Load_kg_After_SDR_Applied_Fine = 0
// 		Gully_Daily_Load_kg_After_SDR_Applied_Coarse = 0

// 		//Rob added the division of SECONDS_PER_DAY, as the Gully_Load time series is kg/d
// 		//this.quickflowConstituent = Gully_Load / Constants.SECONDS_PER_DAY;

// 		//double Event_Runoff = Math.Max(0, Daily_Runoff - Daily_Baseflow);
// 		Event_Runoff := Daily_Runoff

// 		//Need to convert to MM for comparison with annual vals
// 		Event_Runoff_MM := ((Event_Runoff * timeStepInSeconds) / area) / conv.MILLIMETRES_TO_METRES

// 		// System.Diagnostics.Debug.WriteLine(Now.ToShortDateString()+ " Catchment: " + ((LumpedSimHyd)this.RainfallRunoffModel).CatchmentIamIn + " RRm Runoff: " + ((LumpedSimHyd)this.RainfallRunoffModel).runoff + " Gully Runoff: " + Daily_Runoff + " Annual Runoff: " + Annual_Runoff);

// 		//if (Now.Year <= Gully_End_Year && Now.Year >= Gully_Year_Disturb)
// 		if year >= yearDisturbance {
// 			//if (Annual_Gully_Load == 0 && (Gully_Annual_Fine_Load > 0 || Gully_Annual_Coarse_Load > 0))
// 			//{
// 			//    Annual_Gully_Load = Gully_Annual_Fine_Load + Gully_Annual_Coarse_Load;
// 			//}

// 			activityFactor = 1
// 			//Apply the selected gully activity factor after the 'end' ot 'maturity' date
// 			if year > yearActivityEnd {
// 				activityFactor = Average_Gully_Activity_Factor
// 			}

// 			if Event_Runoff == 0 || Annual_Runoff == 0 || Gully_Annual_Average_Sediment_Supply == 0 { // || (Gully_Annual_Fine_Load == 0)

// 				Gully_Daily_Load_kg_Fine = 0
// 				Gully_Daily_Load_kg_Coarse = 0
// 				Gully_Daily_Load_kg_After_SDR_Applied_Fine = 0
// 				Gully_Daily_Load_kg_After_SDR_Applied_Coarse = 0

// 			} else {
// 				//Use total runoff, not event runoff, for this ratio calculation

// 				//Need to do some IFs here to work out when to apply average/management factors to Fine sediment (last 20 years???)
// 				//Gully_Daily_Load_kg_Fine = (Daily_Runoff/Annual_Runoff)* Gully_Annual_Fine_Load*Average_Gully_Activity_Factor * Gully_Management_Practice_Factor;

// 				propFine = Gully_Percent_Fine / 100

// 				if gullyModelType == ListBoxOptions.GullyModelType.DERMRATIO || gullyModelType == ListBoxOptions.GullyModelType.DERM { // && Annual_Gully_Load > 0, no need for this checker anymore, as our projects have transitioned
// 					//Annual_Gully_Load has already had the 'yearly proportion' taken into account (could be a non-linear calculation
// 					//and has also had the 'annual runoff magnitude compared to average annual runoff' adjustment made during parameterisation

// 					//Gully_Daily_Load_kg_Fine = (Daily_Runoff / Annual_Runoff) * Gully_Annual_Fine_Load;
// 					//Gully_Daily_Load_kg_Coarse = (Daily_Runoff / Annual_Runoff) * Gully_Annual_Coarse_Load;

// 					//double fact = (Event_Runoff / Annual_Runoff);

// 					//DateTime checker = new DateTime(1987, 1, 15);
// 					////if (quickflow > 0)
// 					//if (Now == checker && this.areaInSquareMeters > 19729119.6 && this.areaInSquareMeters < 19729119.7)
// 					//{
// 					//    string stopper = "";
// 					//}

// 					Gully_Daily_Load_kg_Fine = (Event_Runoff_MM / Annual_Runoff) * propFine * activityFactor * Gully_Management_Practice_Factor * Annual_Gully_Load

// 					Gully_Daily_Load_kg_Coarse = (Event_Runoff_MM / Annual_Runoff) * (1 - propFine) * Annual_Gully_Load * Gully_Management_Practice_Factor

// 				} else {
// 					//Scott's simplified factor to break annual load into daily
// 					annualToDailyAdjustmentFactor = 1 / 365.25

// 					thisYearsSedimentSupply = Gully_Annual_Average_Sediment_Supply

// 					dailyRunoffFactor = 1
// 					//Stop NaN's on models that don't have the required longterm flow analysis
// 					if Gully_Long_Term_Runoff_Factor > 0 {
// 						if Gully_Daily_Runoff_Power_Factor <= 0 {
// 							Gully_Daily_Runoff_Power_Factor = 1
// 						}

// 						//Swap these over if reverting to Scott's power-based event-to-annual adjustment

// 						//Scott's complex version with stuffed raised to to a power
// 						//all cumecs
// 						dailyRunoffFactor = Math.Pow(Event_Runoff, Gully_Daily_Runoff_Power_Factor) / Gully_Long_Term_Runoff_Factor
// 					}

// 					Gully_Daily_Load_kg_Fine = annualToDailyAdjustmentFactor * dailyRunoffFactor * propFine * activityFactor * Gully_Management_Practice_Factor * thisYearsSedimentSupply * ConversionConst.Tonnes_to_Kilograms

// 					Gully_Daily_Load_kg_Coarse = annualToDailyAdjustmentFactor * dailyRunoffFactor * (1 - propFine) * thisYearsSedimentSupply * ConversionConst.Tonnes_to_Kilograms * Gully_Management_Practice_Factor

// 				}

// 				Gully_Daily_Load_kg_After_SDR_Applied_Fine = Gully_Daily_Load_kg_Fine * (Gully_SDR_Fine * 0.01)
// 				Gully_Daily_Load_kg_After_SDR_Applied_Coarse = Gully_Daily_Load_kg_Coarse * (Gully_SDR_Coarse * 0.01)

// 			}
// 		}

// 		//Set this to zero for now, as until such time as Annual_Gully_Load is populated as a TimeSeries via Reflection,
// 		//we need to be able to use the Fine and Coarse loads that ARE in reflection.
// 		//Annual_Gully_Load = 0;
// 	}
// }
