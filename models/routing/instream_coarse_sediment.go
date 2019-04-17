package routing

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
InstreamCoarseSediment:
	inputs:
		incomingMass:
  states:
		totalStoredMass:
	parameters:
	outputs:
		loadDownstream:
	implementation:
		function: instreamCoarseSediment
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		sediment transport
*/

func instreamCoarseSediment(incomingMass data.ND1Float64,
	storedMass float64,
	loadDownstream data.ND1Float64) float64 {
	n := incomingMass.Len1()
	idx := []int{0}

	for i := 0; i < n; i++ {
		idx[0] = i
		//Robs standard approach at working out what constituent load we can manipulate this timestep
		//ToolsModel.determineOuflowAndPreProcessingLoads(this);

		//double totalConstsituentLoad = GetTotalConstituentLoadWithOutflow;
		//ConstituentStorage = totalConstsituentLoad;

		//This does not yet deal with deposition, but at least we are now bringing in the bank eroded coarse sediment
		//Previously it was never added to the catchment provided coarse sediment (lateralinflow)

		dailyCoarseSedDeposited_Kg := 0.0
		//TotalDailyLoadCoarse_Kg_per_DayOut = 0;
		totalDailyConstituentMass := 0.0

		//SedNet_InStream_Fine_Sediment_Model relatedFineSedModel = (SedNet_InStream_Fine_Sediment_Model)ToolsModel.getStreamProcessingModel((Link)Link, fineSed);
		//SedNet_InStream_Fine_Sediment_Model relatedFineSedModel = null;

		//ConstituentOutput divisionConstituents = Division.ConstituentOutputs.Get(Constituent);
		//DivisionConstituentOutput divisionConstituents = divisionConstituents;
		//totalDailyConstituentMass = ConstituentOutput.DownstreamFlowMass + ConstituentOutput.StoredMass;
		totalDailyConstituentMass = storedMass + incomingMass.Get(idx)

		// NOT NEEDED		combinedCoarseSedInFlows_Kg_per_Day = (CatchmentInflowMass) + (UpstreamFlowMass)

		//totalVolumeofWaterInLink := reachVolume.Get(idx) + outflow.Get(idx)

		//Use this to implement the same deposition model as Fine sediment
		//Currently just drop everything

		//dailyCoarseSedDeposited_Kg = ConstituentStorage;
		dailyCoarseSedDeposited_Kg = totalDailyConstituentMass
		//ConstituentStorage -= dailyCoarseSedDeposited_Kg;
		totalDailyConstituentMass = 0

		storedMass += dailyCoarseSedDeposited_Kg

		//TotalDailyLoadCoarse_Kg_per_DayOut = 0;
		//InChannelStorage();

		//ToolsModel.setLinkSourceSinkModelStorageAndOutflow(this, ConstituentOutput, this.Constituent, totalDailyConstituentMass);
		//SetStorageAndOutflow(totalDailyConstsituentMass);

		////LoadOut = ConstituentOutflow * timeStepInSeconds;
		//residualCoarseSedInLink = ConstituentStorage;
		//TotalDailyLoadCoarse_Kg_per_DayOut = ConstituentOutflow * timeStepInSeconds;

		// NOT NEEDED ChannelSedimentStoreDepth_M := storedMass / (linkWidth * linkLength * sedBulkDensity * units.TONNES_TO_KG)

		//Update the total deposited for reporting

		//This one now updates here, and ResultsGopher uses from here

		loadDownstream.Set(idx, totalDailyConstituentMass)
	}
	return storedMass
}
