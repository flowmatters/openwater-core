package routing

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
InstreamFineSediment:
	inputs:
		incomingMass:
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
		durationInSeconds:
	outputs:
		loadDownstream:
		loadToFloodplain:
		floodplainDepositionFraction:
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

func instreamFineSediment(incomingMass, reachVolume, outflow data.ND1Float64,
	channelStoreFine, totalStoredMass float64,
	bankFullFlow, fineSedSettVelocityFlood, floodPlainArea,
	linkWidth, linkLength, linkSlope, bankHeight,
	propBankHeightForFineDep, sedBulkDensity, manningsN,
	fineSedSettVelocity, fineSedReMobVelocity, durationInSeconds float64,
	loadDownstream, loadToFloodplain, floodplainDepositionFraction data.ND1Float64) (float64, float64) {
	n := reachVolume.Len1()
	idx := []int{0}

	linkArea := linkWidth * linkLength
	maxStorage := propBankHeightForFineDep * bankHeight * linkArea * sedBulkDensity * units.TONNES_TO_KG

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
		incomingMassNow := incomingMass.Get(idx)
		outflowNow := outflow.Get(idx)
		reachVolumeNow := reachVolume.Get(idx)

		totalDailyConstsituentMass := totalStoredMass + incomingMassNow

		totalVolume := reachVolumeNow + outflowNow

		//Use this for assessing proportions deposited on floodplain and stream bed
		combinedConstituentStorageBeforeDeposition := totalDailyConstsituentMass
		outflowRate := outflowNow / durationInSeconds

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
			linkArea, linkWidth, linkSlope, bankHeight, sedBulkDensity, manningsN,
			fineSedSettVelocity, fineSedReMobVelocity, maxStorage)

		channelStoreFine += netStreamDepositionFineSed
		totalDailyConstsituentMass -= netStreamDepositionFineSed

		// //Stop division by zero
		// if combinedConstituentStorageBeforeDeposition > 0 {
		// 	//If this is negative, then we've picked up material
		// 	//proportionDepositedBed = (fineDailyDeposited - fineDailyReMobilised) / combinedConstituentStorageBeforeDeposition;
		// 	proportionDepositedBed = netStreamDepositionFineSed / combinedConstituentStorageBeforeDeposition
		// }

		//this gets apportioned to outflow & storage by conecntration in StorageRoutingConstituentProvider.ProcessLumped
		loadDownstream.Set(idx, totalDailyConstsituentMass)
		loadToFloodplain.Set(idx, floodPlainDepositionFine_Kg_per_Day)
		floodplainDepositionFraction.Set(idx, proportionDepositedFloodplain)
	}

	return channelStoreFine, totalStoredMass
}

func floodPlainDepositionEmperical(outflow, totalDailyConstsituentMass,
	bankFullFlow, fineSedSettVelocityFlood, floodPlainArea float64) float64 {
	FloodPlainDepositionFine_Kg_per_Day := 0.0
	if outflow > bankFullFlow {

		Qf := outflow - bankFullFlow
		FloodFlowProp := Qf / outflow
		expTerm := -1 * ((fineSedSettVelocityFlood * floodPlainArea) / Qf)

		//This will be zero if FloodPlainArea_M2 is zero
		FloodPlainDepositionFine_Kg_per_Day := totalDailyConstsituentMass * FloodFlowProp * (1.0 - math.Exp(expTerm))

		//safety net, shouldn't happen given the eqn above
		if FloodPlainDepositionFine_Kg_per_Day > totalDailyConstsituentMass {
			FloodPlainDepositionFine_Kg_per_Day = totalDailyConstsituentMass
		}

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
	linkArea, linkWidth, linkSlope, bankHeight, sedBulkDensity, manningsN,
	fineSedSettVelocity, fineSedReMobVelocity, maxStorage float64) float64 {
	fineDailyReMobilised := 0.0
	fineDailyDeposited := 0.0

	totalFootprintArea := linkArea

	DoStorage := true
	if DoStorage && totalVolume > 0 {
		loadInStreamBeforeDep_tons := totalDailyConstsituentMass / 1000

		//		inChannelCalcs(subStreamFootprintArea, loadInStreamBeforeDep_tons)
		mainFootprintArea := totalFootprintArea

		sedsetFine := fineSedSettVelocity
		sedsetReMob := fineSedReMobVelocity

		thisSegChannelFootprint := mainFootprintArea
		subStreamsFootprintArea := 0.0 // TODO

		//Will be 1 for sub-cats with no sub-streams
		propTotalStreamFootprint := thisSegChannelFootprint / (subStreamsFootprintArea + mainFootprintArea)

		outflowVal := outflow * propTotalStreamFootprint

		//multiply by seconds_in_one_day to get t/d, not t/s

		STC_Dep_t := (0.1 * (math.Pow(outflowVal, 1.4) * math.Pow(linkSlope, 1.3)) /
			(sedsetFine * math.Pow(linkWidth, 0.4) * math.Pow(manningsN, 0.6))) * units.SECONDS_PER_DAY
		STC_Mob_t := (0.1 * (math.Pow(outflowVal, 1.4) * math.Pow(linkSlope, 1.3)) /
			(sedsetReMob * math.Pow(linkWidth, 0.4) * math.Pow(manningsN, 0.6))) * units.SECONDS_PER_DAY

		loadInThisSegBeforeDep_tons := propTotalStreamFootprint * loadInStreamBeforeDep_tons
		if loadInThisSegBeforeDep_tons > STC_Dep_t {
			//Threshold of dep less than our load, so must be some deposition (and therefore no remobilisation)
			//Can only deposit as much as we have room for (thus checking how much 'room' is left)

			//Must do the daily deposition like this, adjusting TotalDailyLoadFine_Kg_per_DayOut later,
			//otherwise we get tiny rounding errors that can result in negative ConstituentStorage
			//Which compound through the system

			//the amount deposited actually is dependent on ConstituentStorage, not the ConstituentStorage + ChannelStorage....
			availDepFromStorage := (loadInThisSegBeforeDep_tons - STC_Dep_t) * 1000

			fineDailyDeposited += math.Min(availDepFromStorage, maxStorage-(propTotalStreamFootprint*initialChannelStore))

		} else if loadInThisSegBeforeDep_tons < STC_Mob_t {
			//Some remobilisation
			availReMob := (STC_Mob_t - loadInThisSegBeforeDep_tons) * 1000

			fineDailyReMobilised += math.Min(availReMob, (propTotalStreamFootprint * initialChannelStore))
		}

	}

	netStreamDepositionFineSed := fineDailyDeposited - fineDailyReMobilised

	//This is the mass in the stream right now, will be applied to Storage and Outflow
	//Moved to main runTimeStep to assist following process, April 2016
	//totalDailyConstsituentMass -= netStreamDepositionFineSed;//Should decrease if deposition, increase if remobilisation

	//totalStreamDepositionFineSed += netStreamDepositionFineSed //Should increase if deposition, decrease if remobilisation

	//TODO Update in caller channelStoreFine += netStreamDepositionFineSed //Should increase if deposition, decrease if remobilisation

	//ChannelSedimentStoreDepth_M = ChannelStore_Fine_kg / (LinkWidth_M * LinkLength_M * sedBulkDensity * conv.Tonnes_to_Kilograms);
	// TODO Update in caller (if required?) channelSedimentStoreDepth = (channelStoreFine / (sedBulkDensity * units.TONNES_TO_KG)) / totalFootprintArea

	return netStreamDepositionFineSed
}

/*
NOTES:

* Can this model be split up? Bank erosion, sediment deposition, etc?
* And related to coarse sediment model, how?

*/
