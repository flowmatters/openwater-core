package rr

// TODO: init states

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
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
		lzpk: '[0,1] Lower zone Primary Free water base flow ratio, default=0.01'
		lzsk: '[0,1] Lower zone Supplementary Free water base flow ratio, default=0.05'
		uzk: '[0,1] Upper zone free water interflow fraction, default=0.3'
		uztwm: '[0.1,125]mm Upper zone tension water maximum, default=50'
		uzfwm: '[0,75]mm Upper zone free water maximum, default=40'
		lztwm: '[0,300]mm Lower zone tension water maximum, default=130'
		lzfsm: '[0,300]mm Lower zone free water supplemental maximum, default=25'
		lzfpm: '[0,600]mm Lower zone free water primary maximum, default=60'
		pfree: '[0,1] Minimum proportion of percolation from upper zone to lower zone directly available for recharing lower zone free water stores., default=0.06'
		rexp:  '[0,3] Exponent of rate of change of percolation rate with changing LZ storage, default=1.0'
		zperc: '[0,80] Proportional increase in Pbase defining maximum percolation rate, default=40'
		side:  '[0,1] Ratio of non-channel baseflow to channel baseflow, default=0'
		ssout: '[0,]mm Volume of flow that can be conveyed by porous material in streambed, default=0'
		pctim: '[0,1] fraction of catchment that is permanently and directly connected impervious, default=0.01'
		adimp: '[0,1] additional fraction of catchment that can act impervious under saturated soil conditions.,default=0'
		sarva: '[0,1] fraction of basin normally covered by streams,lakes and vegetation that can deplete streamflow by evapotranspiration.,default=0'
		rserv: '[0,1] Fraction lower zone free water that is not available for transpiration,default=0.3'
		uh1: '[0,1] Unit hydrograph - proportion runoff that is instantaneous,default=0.8'
		uh2: '[0,1] Unit hydrograph - proportion lagged by one timestep,default=0.1'
		uh3: '[0,1] Unit hydrograph - proportion lagged by two timesteps,default=0.05'
		uh4: '[0,1] Unit hydrograph - proportion lagged by three timesteps,default=0.03'
		uh5: '[0,1] Unit hydrograph - proportion lagged by four timesteps,default=0.02'
	outputs:
		actualET: mm
		runoff: mm
		imperviousRunoff: mm
		surfaceRunoff: mm
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

const pdn20 = 5.08
const pdnor = 25.4
const nunit = 5
const VERY_SMALL = 0.0 // 1.0e-10

func sumSlice(s []float64) (sum float64) {
	sum = 0.0
	for _, v := range s {
		sum += v
	}

	return
}

func sacramento(rainfall, pet data.ND1Float64,
	uprTensionWater, uprFreeWater, lwrTensionWater,
	lwrPrimaryFreeWater, lwrSupplFreeWater, additionalImperviousStore float64,
	lzpk, lzsk, uzk, uztwm, uzfwm, lztwm, lzfsm, lzfpm, pfree, rexp,
	zperc, side, ssout, pctim, adimp, sarva, rserv,
	uh1, uh2, uh3, uh4, uh5 float64,
	actualET, runoff, imperviousRunoff, surfaceRunoff, baseflow data.ND1Float64) (
	float64, // final uprTensionWater,
	float64, // final uprFreeWater,
	float64, // final lwrTensionWater,
	float64, // final lwrPrimaryFreeWater
	float64, // final lwrSupplFreeWater
	float64) { // final additionalImperviousStore
	nDays := rainfall.Len1()

	// percMax := pbase * (1 + zperc)
	// lowerMax := (1+side)*(lzfpm+lzfsm) + lztwm
	// uhTotal := uh1 + uh2 + uh3 + uh4 + uh5
	// uh1 /= uhTotal
	// uh2 /= uhTotal
	// uh3 /= uhTotal
	// uh4 /= uhTotal
	// uh5 /= uhTotal
	qq := make([]float64, nunit)
	dro := make([]float64, nunit)
	dro[0] = uh1
	dro[1] = uh2
	dro[2] = uh3
	dro[3] = uh4
	dro[4] = uh5

	saved := rserv * (lzfpm + lzfsm)
	alzfsm := lzfsm * (1. + side)
	alzfpm := lzfpm * (1. + side)
	pbase := (alzfsm*lzsk + alzfpm*lzpk) // * (1 + side)

	alzfsc := lwrSupplFreeWater * (1. + side)
	alzfpc := lwrPrimaryFreeWater * (1. + side)
	idx := []int{0}
	for timestep := 0; timestep < nDays; timestep++ {
		idx[0] = timestep
		// prevUprTensionWater := uprTensionWater
		// prevUprFreeWater := uprFreeWater
		// prevLwrTensionWater := lwrTensionWater
		// prevLwrPrimaryFreeWater := lwrPrimaryFreeWater
		// prevLwrSuppFreeWater := lwrSupplFreeWater
		// prevAddImpStore := additionalImperviousStore
		// prevHydrographStore := sumSlice(qq)
		evapt := pet.Get(idx)
		pliq := rainfall.Get(idx)

		//     Determine evaporation from upper zone tension water store
		e1 := 0.0
		if uztwm > VERY_SMALL {
			e1 = evapt * uprTensionWater / uztwm
		}

		//     Determine evaporation from free water surface
		e2 := 0.0
		if uprTensionWater < e1 {
			e1 = uprTensionWater
			uprTensionWater = 0.
			e2 = math.Min(evapt-e1, uprFreeWater)
			uprFreeWater = uprFreeWater - e2
		} else {
			uprTensionWater = uprTensionWater - e1
		}

		//     If the upper zone free water ratio exceeded the upper tension zone
		//     content ratio, then transfer the free water into tension
		a := 1.0
		if uztwm > VERY_SMALL {
			//      if( uztwm > tiny(uztwm) ) then
			a = uprTensionWater / uztwm
		}

		b := 1.0
		if uzfwm > VERY_SMALL {
			//  //!REB  This should be > 0.0 as it is the
			//                                      Upper zone free water capacity
			b = uprFreeWater / uzfwm
		}

		if a < b {
			a = (uprTensionWater + uprFreeWater) / (uztwm + uzfwm)
			uprTensionWater = uztwm * a
			uprFreeWater = uzfwm * a
		}

		//     Evaporation from ADIMP area and Lower zone tension water
		e3 := 0.0
		e5 := 0.0
		if uztwm+lztwm > VERY_SMALL {
			//      if( uztwm+lztwm > tiny(uztwm) ) then
			e3 = math.Min((evapt-e1-e2)*lwrTensionWater/(uztwm+lztwm), lwrTensionWater)
			e5 = math.Min(e1+(evapt-e1-e2)*(additionalImperviousStore-e1-uprTensionWater)/(uztwm+lztwm), additionalImperviousStore)
		}

		//     Compute the transpiration loss from the lower zone tension
		lwrTensionWater = lwrTensionWater - e3
		//     Adjust the impervious area store
		additionalImperviousStore = additionalImperviousStore - e5
		e1 = e1 * (1 - adimp - pctim)
		e2 = e2 * (1 - adimp - pctim)
		e3 = e3 * (1 - adimp - pctim)
		e5 = e5 * adimp

		//     Resupply the lower zone tension with water from the lower zone
		//     free, if more water is available there.
		if lztwm > VERY_SMALL {
			//      if( lztwm > tiny(lztwm) ) then
			a = lwrTensionWater / lztwm
		} else {
			a = 1.
		}

		if alzfpm+alzfsm-saved+lztwm > VERY_SMALL {
			//      if( alzfpm+alzfsm-saved+lztwm > tiny(lztwm) ) then
			b = (alzfpc + alzfsc - saved + lwrTensionWater) / (alzfpm + alzfsm - saved + lztwm)
		} else {
			b = 1.
		}
		if a < b {
			del := (b - a) * lztwm
			//       Transfer water from the lower zone secondary free water to lower zone
			//       tension water store
			lwrTensionWater = lwrTensionWater + del
			alzfsc = alzfsc - del
			if alzfsc < 0 {
				//         Transfer primary free water if secondary free water is inadequate
				alzfpc = alzfpc + alzfsc
				alzfsc = 0.
			}
		}

		//     Runoff from the impervious or water covered area
		roimp := pliq * pctim

		//     Reduce the rain by the amount of upper zone tension water deficiency
		pav := pliq + uprTensionWater - uztwm
		if pav < 0 {
			//       Fill the upper zone tension water as much as rain permits
			additionalImperviousStore = additionalImperviousStore + pliq
			uprTensionWater = uprTensionWater + pliq
			pav = 0.
		} else {

			additionalImperviousStore = additionalImperviousStore + uztwm - uprTensionWater
			uprTensionWater = uztwm
		}

		//     Determine the number of increments
		var adj float64
		var itime int

		if pav <= pdn20 {
			adj = 1.
			itime = 2
		} else {
			if pav < pdnor {
				//         Effective rainfall in a period is assumed to be half of the
				//         period length for rain equal to the normal rainy period
				adj = 0.5 * math.Sqrt(pav/pdnor)
			} else {
				adj = 1.0 - 0.5*pdnor/pav
			}
			itime = 1
		}

		var duz float64

		flobf := 0.
		flosf := 0.
		floin := 0.
		hpl := alzfpm / (alzfpm + alzfsm)
		for ii := itime; ii <= 2; ii++ {
			ninc := int(math.Floor((uprFreeWater*adj+pav)*0.2)) + 1
			dinc := 1. / float64(ninc)
			pinc := pav * dinc
			dinc = dinc * adj
			dlzp := 0.0
			dlzs := 0.0

			if ninc == 1 && adj >= 1.0 {
				duz = uzk
				dlzp = lzpk // TODO confirm local variable
				dlzs = lzsk
			} else {
				if uzk < 1. {
					duz = 1. - math.Pow(1.-uzk, dinc)
				} else {
					duz = 1.0
				}

				if lzpk < 1. {
					dlzp = 1. - math.Pow(1.-lzpk, dinc)
				} else {
					dlzp = 1.0
				}

				if lzsk < 1. {
					dlzs = 1. - math.Pow(1.-lzsk, dinc)
				} else {
					dlzs = 1.0
				}
			}

			//       Drainage and percolation loop
			for inc := 1; inc <= ninc; inc++ {
				ratio := (additionalImperviousStore - uprTensionWater) / lztwm
				addro := pinc * ratio * ratio

				//         Compute the baseflow from the lower zone
				//          bf= alzfpc*dlzp
				bf := 0.0
				if alzfpc > VERY_SMALL {
					//          if( alzfpc > tiny(alzfpc) ) then //!REB Epsilon*Real should= 0.0
					bf = alzfpc * dlzp //!REB this is a strange problem
				} else { //!REB
					alzfpc = 0. //!REB
					bf = 0.     //!REB
				} //!REB
				flobf = flobf + bf
				alzfpc = alzfpc - bf
				//          bf= alzfsc*dlzs
				if alzfsc > VERY_SMALL { //!REB
					//          if( alzfsc > tiny(alzfsc) ) then //!REB Epsilon*Real should= 0.0
					bf = alzfsc * dlzs //!REB this is also a strange problem
				} else { //!REB
					alzfsc = 0. //!REB
					bf = 0.     //!REB
				} //!REB
				alzfsc = alzfsc - bf
				flobf = flobf + bf

				//         Adjust the upper zone for percolation and interflow
				if uprFreeWater > VERY_SMALL { //!REB
					//           Determine percolation from the upper zone free water
					//           limited to available water and lower zone air space
					lzair := lztwm - lwrTensionWater + alzfsm - alzfsc + alzfpm - alzfpc
					perc := 0.0
					if lzair > VERY_SMALL {
						//            if( lzair > tiny(lzair) ) then
						perc = (pbase * dinc * uprFreeWater) / uzfwm
						perc = math.Min(lzair, math.Min(uprFreeWater,
							perc*(1.+(zperc*math.Pow(1.-(alzfpc+alzfsc+lwrTensionWater)/(alzfpm+alzfsm+lztwm), rexp)))))
						uprFreeWater = uprFreeWater - perc
					}

					//           Compute the interflow
					del := duz * uprFreeWater
					floin = floin + del
					uprFreeWater = uprFreeWater - del

					//           Distribute water to lower zone tension and free water stores
					perctw := math.Min(perc*(1.-pfree), lztwm-lwrTensionWater)
					percfw := perc - perctw
					//           Shift any excess lower zone free water percolation to the
					//           lower zone tension water store
					lzair = alzfsm - alzfsc + alzfpm - alzfpc
					if percfw > lzair {
						perctw = perctw + percfw - lzair
						percfw = lzair
					}
					lwrTensionWater = lwrTensionWater + perctw

					//           Distribute water between LZ free water supplemental and primary
					if percfw > VERY_SMALL {
						//            if( percfw > tiny(percfw) ) then
						ratlp := 1. - alzfpc/alzfpm
						ratls := 1. - alzfsc/alzfsm
						percs := math.Min(alzfsm-alzfsc,
							percfw*(1.-hpl*(ratlp+ratlp)/(ratlp+ratls)))
						alzfsc = alzfsc + percs
						//             Check for spill from supplemental to primary
						if alzfsc > alzfsm {
							percs = percs - alzfsc + alzfsm
							alzfsc = alzfsm
						}
						alzfpc = alzfpc + percfw - percs
						//             Check for spill from primary to supplemental
						if alzfpc > alzfpm {
							alzfsc = alzfsc + alzfpc - alzfpm
							alzfpc = alzfpm
						}
					}
				}

				//         Fill upper zone free water with tension water spill
				if pinc > VERY_SMALL {
					//          if( pinc > tiny(pinc) ) then
					pav = pinc
					if pav-uzfwm+uprFreeWater <= 0 {
						uprFreeWater = uprFreeWater + pav
					} else {
						pav = pav - uzfwm + uprFreeWater
						uprFreeWater = uzfwm
						flosf = flosf + pav
						addro = addro + pav*(1.-addro/pinc)
					}
				}
				additionalImperviousStore = additionalImperviousStore + pinc - addro
				roimp = roimp + addro*adimp
			}
			adj = 1. - adj
			pav = 0.
		}

		//     Compute the storage volumes, runoff components and evaporation
		//     Note evapotranspiration losses from the water surface and
		//     riparian vegetation areas are computed in stn7a
		flosf = flosf * (1. - pctim - adimp)
		floin = floin * (1. - pctim - adimp)
		flobf = flobf * (1. - pctim - adimp)

		lwrSupplFreeWater = alzfsc / (1. + side)
		lwrPrimaryFreeWater = alzfpc / (1. + side)
		qq[0] = flosf + roimp + floin

		//           Adjust flow for unit hydrograph
		flwsf := 0.
		for j := 0; j < nunit; j++ {
			flwsf = flwsf + qq[j]*dro[j]
		}
		for k := (nunit - 1); k > 0; k-- {
			qq[k] = qq[k-1]
		}

		flwbf := flobf / (1. + side)
		if flwbf < 0. {
			flwbf = 0.
		}
		baseflowFraction := 0.0

		//           Subtract losses from the total channel flow
		qf := flwbf + flwsf

		if qf > 0.0 {
			baseflowFraction = flwbf / qf
		}

		qf = math.Max(0., qf-ssout)

		e4 := math.Min(evapt*sarva, qf)
		qf = qf - e4

		//           Route the flows if required

		// if NrOut > 0 {
		// 	volsum := qf * SqMi
		// 	for j := 0; j < Nrout; j++ {
		// 		qnow[j] = math.Min(Volsum, Volum(j))
		// 		volsum = math.Max(0., volsum-Volum(j))
		// 	}
		// 	qnow(Nrout + 1) = volsum

		// 	//             Set qold to qnow on the first time step
		// 	if qold(1) < 0.0 {
		// 		for j := 0; j <= Nrout; j++ {
		// 			qold(j) = qnow(j)
		// 		}
		// 	}

		// 	//             Route the flows
		// 	qf = 0.
		// 	for j = 0; j <= Nrout; j++ {
		// 		qold(j) = qold(j)*(1.0-RCoef(j)) + qnow(j)*RCoef(j)
		// 		qf = qf + qold(j)/SqMi
		// 	}
		// }

		// if qf*SqMi < 0.001 {
		// 	qf = 0.
		// }
		// flwch = qf
		// if Imprt > 0 {
		// 	qf = qf + rryy(jt, 3)/SqMi
		// }

		bf := baseflowFraction * qf
		imperviousRunoff.Set(idx, roimp)
		surfaceRunoff.Set(idx, qf-bf)
		baseflow.Set(idx, bf)
		runoff.Set(idx, qf)
		actualET.Set(idx, e1+e2+e3+e4+e5)
		//hydrographStore := sumSlice(qq)

		// deltaS := ((uprTensionWater-prevUprTensionWater)+
		// 	(uprFreeWater-prevUprFreeWater)+
		// 	(lwrTensionWater-prevLwrTensionWater)+
		// 	(lwrPrimaryFreeWater-prevLwrPrimaryFreeWater)+
		// 	(lwrSupplFreeWater-prevLwrSuppFreeWater)+
		// 	(additionalImperviousStore-prevAddImpStore))*(1.0-pctim) +
		// 	(hydrographStore - prevHydrographStore)
		//aet := e1 + e2 + e3 + e4 + e5

		//baseFlowLoss := ((alzfsc - lwrSupplFreeWater) + (alzfpc - lwrPrimaryFreeWater) + (flobf - flwbf)) * (1.0 - pctim)
		//massBalance := pliq - aet - qf - deltaS - baseFlowLoss - math.Min(ssout, flwbf+flwsf)
		// if math.Abs(massBalance) > VERY_SMALL {
		// 	if !mbError {
		// 		mbError = true
		// 	}
		// }
	}

	return uprTensionWater, uprFreeWater, lwrTensionWater,
		lwrPrimaryFreeWater, lwrSupplFreeWater,
		additionalImperviousStore
}
