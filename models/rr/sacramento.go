package rr

import (
	"math"
)

/*OW-SPEC
Sacramento:
  inputs:
    rainfall: mm
    pet: mm
  states:
    UprTensionWater:
    UprFreeWater:
		LwrTensionWater:
		LwrPrimaryFreeWater:
		LwrSupplFreeWater:
		AdditionalImperviousStore:
	parameters:
		lzpk: '[0,1] Lower zone Primary Free water base flow ratio'
		lzsk: '[0,1] Lower zone Supplementary Free water base flow ratio'
		uzk: '[0,1] Upper zone free water interflow fraction'
		uztwm: '[0.1,125]mm Upper zone tension water maximum'
		uzfwm: '[0,75]mm Upper zone free water maximum'
		lztwm: '[0,300]mm Lower zone tension water maximum'
		lzfsm: '[0,300]mm Lower zone free water supplemental maximum'
		lzfpm: '[0,600]mm Lower zone free water primary maximum'
		pfree: '[0,1] Minimum proportion of percolation from upper zone to lower zone directly available for recharing lower zone free water stores.'
		rexp:  '[0,3] Exponent of rate of change of percolation rate with changing LZ storage'
		zperc: '[0,80] Proportional increase in Pbase defining maximum percolation rate'
		side:  '[0,1] Ratio of non-channel baseflow to channel baseflow'
		ssout: '[0,]mm Volume of flow that can be conveyed by porous material in streambed'
		pctim: '[0,1] fraction of catchment that is permanently, directly connected impervious'
		adimp: '[0,1] additional fraction of catchment that can act impervious under saturated soil conditions.'
		sarva: '[0,1] fraction of basin normally covered by streams,lakes and vegetation that can deplete streamflow by evapotranspiration.'
		rserv: '[0,1] Fraction lower zone free water that is not available for transpiration'
		uh1: '[0,1] Unit hydrograph - proportion runoff that is instantaneous'
		uh2: '[0,1] Unit hydrograph - proportion lagged by one timestep'
		uh3: '[0,1] Unit hydrograph - proportion lagged by two timesteps'
		uh4: '[0,1] Unit hydrograph - proportion lagged by three timesteps'
		uh5: '[0,1] Unit hydrograph - proportion lagged by four timesteps'
	outputs:
		runoff: mm
		imperviousRunoff: mm
		surfaceRunoff: mm
		interflow: mm
		baseflow: mm
	implementation:
		function: sacramento
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		rainfall runoff

*/

import (
	"github.com/flowmatters/openwater-core/data"
)

func sacramento(rainfall data.ND1Float64, pet data.ND1Float64,
	uprTensionWater, uprFreeWater, lwrTensionWater,
	lwrPrimaryFreeWater, lwrSupplFreeWater, additionalImperviousStore float64,
	lzpk, lzsk, uzk, uztwm, uzfwm, lztwm, lzfsm, lzfpm, pfree, rexp,
	zperc, side, ssout, pctim, adimp, sarva, rserv,
	uh1, uh2, uh3, uh4, uh5 float64,
	runoff, imperviousRunoff, surfaceRunoff, interflow, baseflow data.ND1Float64) (
	float64, // final uprTensionWater,
	float64, // final uprFreeWater,
	float64, // final lwrTensionWater,
	float64, // final lwrPrimaryFreeWater
	float64, // final lwrSupplFreeWater
	float64) { // final additionalImperviousStore
	nDays := rainfall.Len1()

	pbase := (lzfsm*lzsk + lzfpm*lzpk) * (1 + side)
	percMax := pbase * (1 + zperc)
	lowerMax := (1+side)*(lzfpm+lzfsm) + lztwm
	uhTotal := uh1 + uh2 + uh3 + uh4 + uh5
	uh1 /= uhTotal
	uh2 /= uhTotal
	uh3 /= uhTotal
	uh4 /= uhTotal
	uh5 /= uhTotal

	for i := 0; i < nDays; i++ {
		ed := pet.Get1(i)
		rain := rainfall.Get1(i)

		e1 := ed * uprTensionWater / uztwm
		uprTensionWater -= e1

		e2 := 0.0
		if e1 < ed {
			e2 = math.Min(ed-e1, uprFreeWater)
		}
		uprFreeWater -= e2

		e3 := lwrTensionWater * math.Min((ed-e1-e2)/(uztwm+lztwm), 1)
		lwrTensionWater -= e3

		e5 := math.Min(e1+(ed-e1-e2)*(additionalImperviousStore-e1-uprTensionWater)/(uztwm+lztwm), additionalImperviousStore)
		additionalImperviousStore -= e5
		if additionalImperviousStore < 0 {
			e5 += additionalImperviousStore
			additionalImperviousStore = 0.0
		}
		e5 *= adimp

		//e4 := ed * sarva

		roImp := rain * pctim

		rainExcess := 0.0

		// After updating uztw
		uprTensionWater += rain

		if uprTensionWater > uztwm {
			rainExcess = uprTensionWater - uztwm
			uprTensionWater = uztwm
		}

		uprFreeWater += rainExcess
		if uprFreeWater > uzfwm {
			rainExcess = uprFreeWater - uzfwm
			uprFreeWater = uzfwm
		}

		lowerCurrent := lwrPrimaryFreeWater + lwrSupplFreeWater + lwrTensionWater
		LZrs := math.Pow(1-lowerCurrent/lowerMax, rexp)
		UZrs := uprFreeWater / uzfwm

		percPotential := pbase * (1 + zperc*LZrs) * UZrs
		lowerDeficit := lowerMax - lowerCurrent
		perc := math.Min(math.Min(math.Min(percPotential, uprFreeWater), lowerDeficit), percMax)
		percLZTW := perc * (1 - pfree)
		perc -= percLZTW

		lwrTensionWater += percLZTW
		if lwrTensionWater > lztwm {
			perc += (lwrTensionWater - lztwm)
			lwrTensionWater = lztwm
		}

		hpl := lzfpm / (lzfpm + lzfsm)
		rs := 1 - lwrSupplFreeWater/lzfsm
		rp := 1 - lwrPrimaryFreeWater/lzfpm

		percLZFWS := math.Min(lzfsm-lwrSupplFreeWater, perc*(1-hpl*(2.0*rp)/(rp+rs)))
		lwrSupplFreeWater += percLZFWS
		lwrPrimaryFreeWater += (perc - percLZFWS)
		if lwrPrimaryFreeWater > lzfpm {
			lwrSupplFreeWater += lwrPrimaryFreeWater - lzfpm
			lwrPrimaryFreeWater = lzfpm
		}

		adjSurfaceRunoff := rainExcess * (1 - pctim - adimp)

		// if uztw is full
		addRo := rainExcess * math.Pow((additionalImperviousStore-uprTensionWater)/lztwm, 2.0)

		// if rainExcess > capacity in lztw
		addRo += (rainExcess - uzfwm + uprFreeWater) * (1 - addRo/rainExcess)

		additionalImperviousStore += rainExcess - addRo
		roImp += addRo * adimp

		iflow := uprFreeWater * uzk * (1 - pctim - adimp) // adimp? really? shouldn't it be the active portion?
		uprFreeWater -= iflow

		//instantaneousRunoff := roImp + adjSurfaceRunoff + iflow

		flobfs := lwrSupplFreeWater * lzsk
		lwrSupplFreeWater -= flobfs

		flopbfp := lwrPrimaryFreeWater * lzpk
		lwrPrimaryFreeWater -= flopbfp

		bflow := flobfs + flopbfp - (flobfs+flopbfp)/(1+side) // WTF?
		// channelLoss := ssout

		imperviousRunoff.Set1(i, roImp)
		surfaceRunoff.Set1(i, adjSurfaceRunoff)
		interflow.Set1(i, iflow)
		baseflow.Set1(i, bflow)
		runoff.Set1(i, roImp+adjSurfaceRunoff+iflow+bflow)
	}

	return uprTensionWater, uprFreeWater, lwrTensionWater,
		lwrPrimaryFreeWater, lwrSupplFreeWater,
		additionalImperviousStore
}
