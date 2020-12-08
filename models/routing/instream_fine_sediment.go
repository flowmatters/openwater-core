package routing

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
InstreamFineSediment:
	inputs:
		upstreamMass:
		lateralMass:
		reachLocalMass:
		reachVolume:
		outflow:
  states:
		channelStoreFine:
		totalStoredMass:
	parameters:
		bankFullFlow:
		fineSedSettVelocityFlood:
		floodPlainArea:
		linkWidth:
		linkLength:
		linkSlope:
		bankHeight:
		propBankHeightForFineDep:
		sedBulkDensity:
		manningsN:
		fineSedSettVelocity:
		fineSedReMobVelocity:
		durationInSeconds: '[1,86400] Timestep, default=86400'
	outputs:
		loadDownstream:
		loadToFloodplain:
		floodplainDepositionFraction:
		channelDepositionFraction:
	implementation:
		function: instreamFineSediment
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		sediment transport
*/

func instreamFineSediment(upstreamMass, lateralMass, reachLocalMass, reachVolume, outflow data.ND1Float64,
	channelStoreFine, totalStoredMass float64,
	bankFullFlow, fineSedSettVelocityFlood, floodPlainArea,
	linkWidth, linkLength, linkSlope, bankHeight,
	propBankHeightForFineDep, sedBulkDensity, manningsN,
	fineSedSettVelocity, fineSedReMobVelocity, durationInSeconds float64,
	loadDownstream, loadToFloodplain, floodplainDepositionFraction, channelDepositionFraction data.ND1Float64) (float64, float64) {
	n := reachVolume.Len1()
	idx := []int{0}

	linkArea := linkWidth * linkLength
	maxStorage := propBankHeightForFineDep * bankHeight * linkArea * sedBulkDensity * units.TONNES_TO_KG
	if channelStoreFine < 0.0 {
		// Treat initial value as proportion of maxStorage
		channelStoreFine = math.Abs(channelStoreFine) * maxStorage
	}

	for i := 0; i < n; i++ {
		idx[0] = i
		//note - Flow Storage is in cubic metres per day
		//note - ConstituentStorage is in kg per day

		///implemented such that all sediment in the link and the flood plain is fully mixed each day
		///

		// JR
		// TotalDailyLoadFine_Kg_per_DayIn = 0
		// fineDailyDeposited = 0
		// fineDailyReMobilised = 0
		// netStreamDepositionFineSed = 0
		// bankErosionFine_Kg_per_Day = 0
		// bankErosionCoarse_Kg_per_Day = 0
		// FloodPlainDepositionFine_Kg_per_Day = 0
		// proportionDepositedFloodplain = 0
		// proportionDepositedBed = 0
		// totalDailyConstsituentMass = 0

		// STC_Dep_t = 0
		// STC_Mob_t = 0

		//		model.step(incomingMass.Get(idx), volumeAtEndTimestep.Get(idx), outflow.Get(idx))
		//NOT USED: ChannelStoreAtStartOfTimeStep_Fine_kg := ism.channelStoreFine
		incomingMassNow := (upstreamMass.Get(idx)+lateralMass.Get(idx)+reachLocalMass.Get(idx)) * durationInSeconds
		outflowRate := outflow.Get(idx)
		outflowNow := outflowRate * durationInSeconds
		reachVolumeNow := reachVolume.Get(idx)

		totalDailyConstsituentMass := totalStoredMass + incomingMassNow
		totalVolume := reachVolumeNow + outflowNow

		//Use this for assessing proportions deposited on floodplain and stream bed
		combinedConstituentStorageBeforeDeposition := totalDailyConstsituentMass

		floodPlainDepositionFine_Kg_per_Day := floodPlainDepositionEmperical(outflowRate, totalDailyConstsituentMass,
			bankFullFlow, fineSedSettVelocityFlood, floodPlainArea)
		//Remove flood plain deposited material from Constituent Storage
		totalDailyConstsituentMass -= floodPlainDepositionFine_Kg_per_Day

		proportionDepositedFloodplain := 0.0
		//Stop division by zero
		if combinedConstituentStorageBeforeDeposition > 0 {
			proportionDepositedFloodplain = floodPlainDepositionFine_Kg_per_Day / combinedConstituentStorageBeforeDeposition
		}

		//This will modify netStreamDepositionFineSed as it processes
		//will only access the amoutn left for deposition that has remained post-flood assessment
		netStreamDepositionFineSed := inChannelStorage(outflowRate, totalVolume, totalDailyConstsituentMass, channelStoreFine,
			linkWidth, linkSlope, manningsN, fineSedSettVelocity, fineSedReMobVelocity, maxStorage)

		proportionDepositedChannel := 0.0
		//Stop division by zero
		if combinedConstituentStorageBeforeDeposition > 0 {
			proportionDepositedChannel = netStreamDepositionFineSed / combinedConstituentStorageBeforeDeposition
		}
	
		channelStoreFine += netStreamDepositionFineSed
		totalDailyConstsituentMass -= netStreamDepositionFineSed

		// //Stop division by zero
		// if combinedConstituentStorageBeforeDeposition > 0 {
		// 	//If this is negative, then we've picked up material
		// 	//proportionDepositedBed = (fineDailyDeposited - fineDailyReMobilised) / combinedConstituentStorageBeforeDeposition;
		// 	proportionDepositedBed = netStreamDepositionFineSed / combinedConstituentStorageBeforeDeposition
		// }
		outflowLoad := 0.0

		if totalVolume > 0 {
			concentration := totalDailyConstsituentMass / totalVolume
			totalStoredMass = concentration * reachVolumeNow
			outflowLoad = concentration * outflowRate
		} else {
			totalStoredMass = 0.0 // totalDailyConstsituentMass
		}

		//this gets apportioned to outflow & storage by concentration in StorageRoutingConstituentProvider.ProcessLumped
		loadDownstream.Set(idx, outflowLoad)
		loadToFloodplain.Set(idx, floodPlainDepositionFine_Kg_per_Day/durationInSeconds)
		floodplainDepositionFraction.Set(idx, proportionDepositedFloodplain)
		channelDepositionFraction.Set(idx,proportionDepositedChannel)
	}

	return channelStoreFine, totalStoredMass
}

func floodPlainDepositionEmperical(outflow, totalDailyConstsituentMass,
	bankFullFlow, fineSedSettVelocityFlood, floodPlainArea float64) float64 {

	if (outflow < bankFullFlow) || (bankFullFlow==0.0) {
		return 0.0
	}

	FloodPlainDepositionFine_Kg_per_Day := 0.0

	Qf := outflow - bankFullFlow
	FloodFlowProp := Qf / outflow
	expTerm := -1 * ((fineSedSettVelocityFlood * floodPlainArea) / Qf)

	//This will be zero if FloodPlainArea_M2 is zero
	FloodPlainDepositionFine_Kg_per_Day = totalDailyConstsituentMass * FloodFlowProp * (1.0 - math.Exp(expTerm))

	//safety net, shouldn't happen given the eqn above
	if FloodPlainDepositionFine_Kg_per_Day > totalDailyConstsituentMass {
		FloodPlainDepositionFine_Kg_per_Day = totalDailyConstsituentMass
	}

	return FloodPlainDepositionFine_Kg_per_Day
}

// func floodPlainDepositionPhysical() {
// 	if (TotalAddedVolume / timeStepInSeconds) > BankFullFlow {
// 		conc_kg_per_m3 = TotalDailyLoadFine_Kg_per_DayOut / totalVolumeofWaterInLink_m3_per_day
// 		VolumeSliceM3 = FloodPlainArea_M2 * 0.0864 //fineSedSettVelocity*conv.seconds_in_one_day;
// 		masstoremoveKg = VolumeSliceM3 * conc_kg_per_m3
// 		FloodPlainDepositionFine_Kg_per_Day = masstoremoveKg
// 	} else {
// 		FloodPlainDepositionFine_Kg_per_Day = 0
// 	}

// 	totalFloodPlainDepositionFineSed += FloodPlainDepositionFine_Kg_per_Day

// 	//Really should try to prevent division by zero here
// 	divisor = sedBulkDensity * units.TONNES_TO_KG * FloodPlainArea_M2

// 	if divisor > 0 {
// 		totalFloodplainDepositionFineDepth_M += FloodPlainDepositionFine_Kg_per_Day / divisor
// 	}

// }

///<summary>
///This is the long term channel storage component. Not to be confused with the much more transient Flow Routing Storage
///<//summary>
func inChannelStorage(outflow, totalVolume, totalDailyConstsituentMass, initialChannelStore,
	linkWidth, linkSlope, manningsN,
	fineSedSettVelocity, fineSedReMobVelocity, maxStorage float64) float64 {

	// fineDailyReMobilised := 0.0
	// fineDailyDeposited := 0.0

	// DoStorage := true
	// if !DoStorage || (totalVolume <= 0) {
	if totalVolume <= 0 {
		return 0.0
	}

	// thisSegChannelFootprint := linkArea
	// subStreamsFootprintArea := 0.0 // TODO

	//Will be 1 for sub-cats with no sub-streams
	propTotalStreamFootprint := 1.0 //thisSegChannelFootprint / (subStreamsFootprintArea + linkArea)

	loadInStreamBeforeDep_tons := totalDailyConstsituentMass * units.KG_TO_TONNES
	loadInThisSegBeforeDep_tons := propTotalStreamFootprint * loadInStreamBeforeDep_tons

	outflowVal := outflow * propTotalStreamFootprint

	//multiply by seconds_in_one_day to get t/d, not t/s

	STC_Dep_t := (0.1 * (math.Pow(outflowVal, 1.4) * math.Pow(linkSlope, 1.3)) /
		(fineSedSettVelocity * math.Pow(linkWidth, 0.4) * math.Pow(manningsN, 0.6))) * units.SECONDS_PER_DAY
	STC_Mob_t := (0.1 * (math.Pow(outflowVal, 1.4) * math.Pow(linkSlope, 1.3)) /
		(fineSedReMobVelocity * math.Pow(linkWidth, 0.4) * math.Pow(manningsN, 0.6))) * units.SECONDS_PER_DAY

	if loadInThisSegBeforeDep_tons > STC_Dep_t {
		//Threshold of dep less than our load, so must be some deposition (and therefore no remobilisation)
		//Can only deposit as much as we have room for (thus checking how much 'room' is left)

		//Must do the daily deposition like this, adjusting TotalDailyLoadFine_Kg_per_DayOut later,
		//otherwise we get tiny rounding errors that can result in negative ConstituentStorage
		//Which compound through the system

		//the amount deposited actually is dependent on ConstituentStorage, not the ConstituentStorage + ChannelStorage....
		availDepFromStorage := (loadInThisSegBeforeDep_tons - STC_Dep_t) * units.TONNES_TO_KG

		return math.Min(availDepFromStorage, maxStorage-(propTotalStreamFootprint*initialChannelStore))

	} else if loadInThisSegBeforeDep_tons < STC_Mob_t {

		//Some remobilisation
		availReMob := (STC_Mob_t - loadInThisSegBeforeDep_tons) * units.TONNES_TO_KG

		return -math.Min(availReMob, (propTotalStreamFootprint * initialChannelStore))
	}

	return 0.0
}

/*
NOTES:

* Can this model be split up? Bank erosion, sediment deposition, etc?
* And related to coarse sediment model, how?

*/
