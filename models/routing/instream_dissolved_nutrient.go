package routing

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/rough"
	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
InstreamDissolvedNutrientDecay:
	inputs:
		incomingMassUpstream:
		incomingMassLateral:
		reachVolume:
		outflow:
		floodplainDepositionFraction:
  states:
		totalStoredMass:
	parameters:
		doDecay:
		pointSourceLoad:
		linkHeight:
		linkWidth:
		linkLength:
		uptakeVelocity:
		durationInSeconds: '[1,86400] Timestep, default=86400'
	outputs:
		loadDownstream:
		loadToFloodplain:
	implementation:
		function: instreamDissolvedNutrient
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

func instreamDissolvedNutrient(incomingMassUpstream, incomingMassLateral, reachVolume, outflow, floodplainDepositionFraction data.ND1Float64,
	storedMass float64,
	doDecay, pointSourceLoad, linkHeight, linkWidth, linkLength, uptakeVelocity, durationInSeconds float64,
	loadDownstream, loadToFloodplain data.ND1Float64) float64 {
	n := incomingMassUpstream.Len1()
	idx := []int{0}
	prevVolume := reachVolume.Get(idx)

	timeStepInDays := units.SECONDS_PER_DAY / durationInSeconds

	pointSourcePerSecond := pointSourceLoad / (rough.DAYS_PER_YEAR * units.SECONDS_PER_DAY)

	if doDecay < 0.5 {
		storedMass = lumpedConstituents(
			incomingMassUpstream, incomingMassLateral, outflow, reachVolume,
			storedMass,
			0, pointSourcePerSecond, durationInSeconds,
			loadDownstream)
		return storedMass
	}

	for i := 0; i < n; i++ {
		idx[0] = i

		reachVolumeNow := reachVolume.Get(idx)
		incomingMassUpstreamNow := incomingMassUpstream.Get(idx)
		incomingMassLateralNow := incomingMassLateral.Get(idx)
		outflowNow := outflow.Get(idx)

		// RiverSystem.Link theLink = (Link)Link;

		//DivisionConstituentOutput divisionConstituents = divisionConstituents;
		//ConstituentOutput divisionConstituents = Division.ConstituentOutputs.Get(Constituent);
		//double totalConstsituentLoad = ConstituentOutput.DownstreamFlowMass + ConstituentOutput.StoredMass;
		incomingMassNow := incomingMassUpstreamNow + incomingMassLateralNow
		totalConstsituentLoad := storedMass + incomingMassNow

		pointSourceLoad_kg := 0.0
		//Convert annual load to daily, if there's a stream volume to put it in
		//This is important, or else internal E2 stuff will throw a double infinity
		//when creating a concentration from a non-zero mass in a zero volume
		//which is a NaN
		//Probably should work off Lateralinflow, but this safer for now

		//if (ConstituentOutput.StoredMass > 0)
		if reachVolumeNow > 0 {
			pointSourceLoad_kg = pointSourcePerSecond
		}

		//Need to grab this to separate existiing ConstituentStorage from included Inflows, as Inflows are added by Flow Routing
		//double constituentStoragePriorToInflows = ConstituentStorage - GetTotalIncomingLoad;

		constituentStoragePriorToInflows := totalConstsituentLoad - incomingMassNow

		totalConstsituentLoad += pointSourceLoad_kg

		////this.Link.Storage.Constituents[this.Constituent].Amount = totalConstsituentLoad;
		//ConstituentOutput.StoredMass = totalConstsituentLoad;
		////ConstituentStorage = totalConstsituentLoad;

		////this.Link.Storage.Constituents[this.Constituent].Amount += DailyPointSourceLoad_kg;
		//ConstituentOutput.StoredMass += DailyPointSourceLoad_kg;
		////ConstituentStorage += DailyPointSourceLoad_kg;

		////double loadOut = this.Link.Storage.Constituents[this.Constituent].Amount;
		//double loadOut = ConstituentOutput.StoredMass;
		loadOut := constituentStoragePriorToInflows
		////double loadOut = ConstituentStorage;

		waterDepth := 0.0

		dailyDecayedConstituentLoad := 0.0
		effectiveDecayCoefficient := 0.0
		decayCoefficient := 1000.0

		travelTimeInSeconds := 0.0

		//Get the average Storage for beginning and end of timestep
		//Use this to give our Depth (stage Height)
		avStorage := (reachVolumeNow + prevVolume) / 2

		//Average depth of water, limited to linkHeight
		waterDepth = math.Min(linkHeight, avStorage/(linkLength*linkWidth))
		crossAreaSection_m2 := waterDepth * linkWidth

		//Now we can calculate an average rate of flow (m/s), and a travel time
		//Assuming that geometries dont change and velocity is constant for the day

		flowVelocity := 0.0
		outflowRate := outflowNow // / durationInSeconds

		if crossAreaSection_m2 > 0 {
			if outflowRate > 0 {
				flowVelocity = outflowRate / crossAreaSection_m2
			}

			if flowVelocity > 0 {
				travelTimeInSeconds = linkLength / flowVelocity
			}
		}

		//Now decay coefficient
		if waterDepth > 0 {
			decayCoefficient = uptakeVelocity / waterDepth
		}

		//Should be zero if decayCoefficient has remained 1000
		effectiveDecayCoefficient = math.Exp(-1 * decayCoefficient * timeStepInDays)

		DailyLateralLoad_Kg_per_s := incomingMassLateralNow + pointSourceLoad_kg
		//DailyFromAboveLoad_Kg_per_Day := UpstreamFlowMass

		//Include amount in link prior to adding inflows
		//Changed as now ConstituentStorage is updated to include Inflows during water routing
		//double allAvailConstit = GetTotalLoad();

		//double allAvailConstit = this.Link.Storage.Constituents[this.Constituent].Amount;
		allAvailConstit := totalConstsituentLoad
		//double allAvailConstit = ConstituentStorage;

		if effectiveDecayCoefficient <= 0 {
			//loadOut = allAvailConstit;

			//ToolsModel.setLinkSourceSinkModelStorageAndOutflow(this, ConstituentOutput, this.Constituent, loadOut);
			////SetStorageAndOutflow(loadOut);

			loadDownstream.Set(idx, totalConstsituentLoad)

			dailyDecayedConstituentLoad = 0
			//ConstituentStorage = 0;
		} else if travelTimeInSeconds <= durationInSeconds {
			//All constituent is subject to decay AND leaves the link

			loadOut = allAvailConstit * effectiveDecayCoefficient
			dailyDecayedConstituentLoad = allAvailConstit - loadOut

			loadDownstream.Set(idx, allAvailConstit-dailyDecayedConstituentLoad)
			////this.Link.Storage.Constituents[this.Constituent].Amount = 0;
			//ConstituentOutput.StoredMass = 0;
			//ConstituentOutput.DownstreamFlowMass = loadOut / _timeStepInSeconds;
			////this.Link.Outflow.LoadPerSecond.Constituents[this.Constituent].Amount = loadOut / _timeStepInSeconds;

			////ConstituentOutflow = loadOut/_timeStepInSeconds;
			////ConstituentStorage = 0;

		} else { //travelTimeInSeconds > theTimeStepInSeconds
			//This calculates some load in ConstituentStorage and a specific ConstituentOutflow rate as well
			//which is acknowledged in the Dyn SedNet spec document
			//This is contradictory to all other Dyn SedNet stream models
			//where all ConstituentStorage and ConstituentOutflow is apportioned by concentration

			loadOut = (DailyLateralLoad_Kg_per_s + constituentStoragePriorToInflows) * (durationInSeconds / travelTimeInSeconds) * effectiveDecayCoefficient

			//timeFraction := (travelTimeInSeconds - durationInSeconds) / travelTimeInSeconds

			////ConstituentOutflow = loadOut / _timeStepInSeconds;
			////ConstituentStorage = ((DailyLateralLoad_Kg_per_Day + constituentStoragePriorToInflows) * timeFraction + DailyFromAboveLoad_Kg_per_Day) * effectiveDecayCoefficient;

			////this.Link.Storage.Constituents[this.Constituent].Amount = ((DailyLateralLoad_Kg_per_Day + constituentStoragePriorToInflows) * timeFraction + DailyFromAboveLoad_Kg_per_Day) * effectiveDecayCoefficient;
			////this.Link.Outflow.LoadPerSecond.Constituents[this.Constituent].Amount = loadOut / _timeStepInSeconds;
			//ConstituentOutput.StoredMass = ((DailyLateralLoad_Kg_per_Day + constituentStoragePriorToInflows) * timeFraction + DailyFromAboveLoad_Kg_per_Day) * effectiveDecayCoefficient;
			//ConstituentOutput.DownstreamFlowMass = loadOut / _timeStepInSeconds;

			//dailyDecayedConstituentLoad = allAvailConstit - ConstituentStorage - loadOut;

			dailyDecayedConstituentLoad = allAvailConstit - loadOut

			loadDownstream.Set(idx, allAvailConstit-dailyDecayedConstituentLoad) // Is this just loadOut?
		}

		prevVolume = reachVolumeNow

	}
	return storedMass
}
