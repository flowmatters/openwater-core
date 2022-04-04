package routing

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
InstreamParticulateNutrient:
	inputs:
		incomingMassUpstream:
		incomingMassLateral:
		reachVolume:
		outflow:
		streambankErosion:
		lateralSediment:
		floodplainDepositionFraction:
		channelDepositionFraction:
  states:
		instreamStoredMass:
		channelStoredMass:
	parameters:
		particulateNutrientConcentration: '[0,1] Proportion of sediment mass, default=0'
		soilPercentFine:
		durationInSeconds: '[1,86400] Timestep, default=86400'
	outputs:
		loadDownstream:
		loadToFloodplain:
	implementation:
		function: instreamParticulateNutrient
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		sediment transport
*/

// Ideally we'd separate out parameters from flags.

// Can I parameterise this WITHOUT the boolean DoDecay?
// Is the theLinkIsLumpedFlowRouting necessary??? assumed if we are using this model?
// Can we take out the point source logic?

func instreamParticulateNutrient(incomingMassUpstream, incomingMassLateral, reachVolume, outflow, 
	streamBankErosion, lateralSediment, floodplainDepositionFraction, channelDepositionFraction data.ND1Float64,
	initialInstreamStoredMass, initialChannelStoredMass float64,
	particulateNutrientConcentration, soilPercentFine, durationInSeconds float64,
	loadDownstream, loadToFloodplain data.ND1Float64) (instreamStoredMass, channelStoredMass float64) {
	n := incomingMassUpstream.Len1()
	idx := []int{0}

	instreamStoredMass = initialInstreamStoredMass
	channelStoredMass = initialChannelStoredMass

	for i:=0; i < n; i++ {
		idx[0]=i
		incomingUpstream := incomingMassUpstream.Get(idx) * durationInSeconds
		incomingLateral := incomingMassLateral.Get(idx) * durationInSeconds

		totalDailyConstsituentMass := instreamStoredMass + incomingUpstream + incomingLateral // + AdditionalInflowMass

		//Do some adjustments to try to overcome the issue where FUs might've provided a Nutrient load (DWC?) but ont a sediment load
		//This ultimately makes very little difference
		totalDailyConstsituentMassForDepositionProcesses := instreamStoredMass + incomingUpstream /*+ AdditionalInflowMass*/
		if (lateralSediment.Get(idx) > 0.0) {
			totalDailyConstsituentMassForDepositionProcesses += incomingLateral
		}

		if totalDailyConstsituentMassForDepositionProcesses < 0.0 {
			totalDailyConstsituentMassForDepositionProcesses = 0.0
		}

		//stream bank generation
		streamBankParticulate := streamBankErosion.Get(idx) * particulateNutrientConcentration * durationInSeconds
		totalDailyConstsituentMass += streamBankParticulate
		totalDailyConstsituentMassForDepositionProcesses += streamBankParticulate * (soilPercentFine / 100)

		//Deposition on floodplain
		fpDepositionFraction := math.Min(math.Max(floodplainDepositionFraction.Get(idx),0.0),1.0)
		nutrientDailyDepositedFloodPlain := fpDepositionFraction * totalDailyConstsituentMassForDepositionProcesses
		loadToFloodplain.Set(idx,nutrientDailyDepositedFloodPlain/durationInSeconds)

		//Intentionally haven't adjusted ConstituentStorage to remove the floodplain deposited material
		//as the proportion of channel storage in fine sediment model is also relevant to pre-floodplain total.

		//Deposition/remobilisation in stream bed (negative value is remobilisation)
		//double bedDeposit = SedMod.proportionDepositedBed * totalDailyConstsituentMass;
		//Potentially the sed model could have re-mobilised with 'zero' existing constituent mass...
		//But there's not much we can do about that
		bedDepositSignal := channelDepositionFraction.Get(idx)
		bedExchange := 0.0

		if bedDepositSignal >= 0 { // Fraction of working mass is deposited
			bedExchange = math.Min(bedDepositSignal * totalDailyConstsituentMassForDepositionProcesses,
														 totalDailyConstsituentMassForDepositionProcesses - nutrientDailyDepositedFloodPlain)
			channelStoredMass += bedExchange
		} else {
			// resuspension := - bedDepositFraction * channelStoredMass // Signal is a fraction of stored sediment
			// resuspension := - bedDepositSignal * particulateNutrientConcentration // Signal is a kg/s of sediment remobilised
			resuspension := - bedDepositSignal * totalDailyConstsituentMassForDepositionProcesses
			channelStoredMass -= resuspension
			bedExchange = - resuspension
		// } else {
		// 	//Remobilisation
		// 	nutrientDailyRemobilisedBed = bedDeposit * -1.0 //This will make remobilisation a positive number
		// 	nutrientDailyDepositedBed = 0.0
		}

		netLoss := nutrientDailyDepositedFloodPlain + bedExchange
		amountLeft := totalDailyConstsituentMass - netLoss

		// copied from lumped constituent - should refactor
		outflowRate := outflow.Get(idx)
		outflowV :=  outflowRate * durationInSeconds
		storedV := reachVolume.Get(idx)

		workingVol := outflowV + storedV
		if workingVol < MINIMUM_VOLUME {
			instreamStoredMass = 0.0 // workingMass
			loadDownstream.Set(idx, 0.0)
			continue
		}

		concentration := amountLeft / workingVol
		instreamStoredMass = concentration * storedV

		outflowLoad := concentration * outflowRate

		loadDownstream.Set(idx,outflowLoad)
	}

	return
}


/* From sediment model

* SoilPercentFine (param)

* CatchmentInflowMass (? from generation models?) 
* BankErosionTotal_kg_per_Day (bank erosion model)
* proportionDepositedFloodplain (existing)
* proportionDepositedBed (not existing?)

*/