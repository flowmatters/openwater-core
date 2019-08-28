package climate

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
ClimateVariables:
	inputs:
		dryBulb: degC
		humidity: '%'
  states:
  parameters:
		elevation: '[0,10000]m Elevation above sea level'
	outputs:
		vaporPressure: kPa
		dewPoint: degC
		wetBulb: degC
		deltaT: degC
	implementation:
		function: climateVariables
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		climate variable estimation
*/

func climateVariables(dryBulb, humidity data.ND1Float64,
	elevation float64,
	vaporPressure, dewPoint, wetBulb, deltaT data.ND1Float64) {
	nDays := dryBulb.Len1()
	idx := []int{0}

	pa := barometricPressure(elevation)

	for i := 0; i < nDays; i++ {
		idx[0] = i
		dryBulbTemp := dryBulb.Get(idx)
		relativeHumidity := humidity.Get(idx)

		vp := calcVaporPressure(dryBulbTemp)
		tdew := calcDewPoint(dryBulbTemp, relativeHumidity)

		e := calcEnthalpy(dryBulbTemp,
			calcHumidityRatioActual(dryBulbTemp, relativeHumidity, pa))

		twetBulbTemp := calcWetBulb(dryBulbTemp, tdew, e, pa)
		vaporPressure.Set(idx, vp)
		dewPoint.Set(idx, tdew)
		wetBulb.Set(idx, twetBulbTemp)
		deltaT.Set(idx, dryBulbTemp-twetBulbTemp)
	}
}

// -----------------------------------------------------------------------------
// Calculate vapor pressure at saturation given dry bulb temp
// where  Temperature = °C
// Result = kPa
//
//VaporPressure_GoffGratch
func calcVaporPressure(temperature float64) float64 {
	const a1 = -7.90298
	const a2 = 5.02808
	const a3 = -0.00000013816
	const a4 = 11.344
	const a5 = 0.0081328
	const a6 = -3.49149
	const b1 = -9.09718
	const b2 = -3.56654
	const b3 = 0.876793
	const b4 = 0.0060273
	var ta, z, p1, p2, p3, p4 float64

	ta = temperature + 273.16
	if temperature > 0 { // above freezing
		z = 373.16 / ta
		p1 = (z - 1) * a1
		p2 = math.Log10(z) * a2
		p3 = ((math.Pow(10, ((1 - (1 / z)) * a4))) - 1) * a3
		p4 = ((math.Pow(10, (a6 * (z - 1)))) - 1) * a5
	} else // below freezing
	{
		z = 273.16 / ta
		p1 = b1 * (z - 1)
		p2 = b2 * math.Log10(z)
		p3 = b3 * (1 - (1 / z))
		p4 = math.Log10(b4)
	}
	return 101.325 * math.Pow(10, (p1+p2+p3+p4))
}

func calcDewPoint(temperature, humidity float64) float64 {
	var ea float64

	if humidity <= 0 {
		humidity = 0.0001 // use 0.01% as minimum humidity
	}
	ea = calcVaporPressure(temperature) * humidity / 100 // actual vapour pressure
	if ea > 0 {
		Func := math.Log(ea / 0.6108)
		return 237.3 * Func / (17.27 - Func)
	}
	return math.NaN()
}

// // Calculate saturation vapor pressure given dry bulb temperature
// func calcVaporPressure(temperature float64) float64 {
// 	{
// 		if temperature < -20 {
// 			return math.NaN()
// 		}

// 		// this is the ASCE standardized equation
// 		return 0.6108 * math.Exp(17.27*temperature/(temperature+237.3))
// 	}
// }

// Estimate barometric pressure (kPa) given elevation (m)
func barometricPressure(elevation float64) float64 {
	return 101.3 * math.Pow((293-0.0065*elevation)/293, 5.26)
}

// -----------------------------------------------------------------------------
// General Humidity Ratio calculation given vapour pressure and atmospheric pressure
//
// where  VaporPressure and AtmPressure are the same units
// HumidityRatio is unitless (e.g. grams/gram, lb/lb)
//
func calcHumidityRatio(vaporPressure, atmPressure float64) float64 {
	return 0.62198 * vaporPressure / (atmPressure - vaporPressure)
}

// -----------------------------------------------------------------------------
// Calculate Actual Humidity Ratio given dry bulb temp, relative humidity, and
// Atmospheric Pressure
//
// where  TDryBulb = °C
// HumidityPC = % relative humidity (0 < HumidityPC <= 100)
// AtmPressure = kPA (e.g. 100)
// HumidityRatio = ratio (grams/gram, lb/lb, etc.)
//
func calcHumidityRatioActual(tDryBulb, humidityPC, atmPressure float64) float64 {
	var vp_sat float64
	vp_sat = calcVaporPressure(tDryBulb)             // saturated vapor pressure
	result := calcHumidityRatio(vp_sat, atmPressure) // saturated humidity ratio
	result = result * humidityPC / 100               // actual humidity ratio
	return result
}

// -----------------------------------------------------------------------------
// Calculate Enthalpy given dry bulb temp and humidity ratio
//
// where  TDryBulb = °C
// HumidityRatio = grams/gram (lb/lb, etc.)
// Result = Joules/gram
//
func calcEnthalpy(tDryBulb, humidityRatio float64) float64 {
	return 1.006*tDryBulb + (1.84*tDryBulb+2501)*humidityRatio
}

// -----------------------------------------------------------------------------
// Wet bulb temperature as a func of dry bulb temperature, dew point,
// enthalpy and atmospheric pressure
// http://hvac-talk.com/vbb/archive/index.php/t-73144.html
// contributor = osiyo: Bob G - Osiyo53@yahoo.com
// where  TDryBulb = °C
// TDewPoint = °C
// Enthalpy = Joules/gram
// AtmPress = kPa (around 100)
//
func calcWetBulb(tDryBulb, tDewPoint, hEnthalpy, pAtmosphere float64) float64 {
	var rtb, dx, xmid, psat, wstar, fmid float64
	// -----------------------------------------------------------------------------
	// Computes wet-bulb temperature iteratively from dry bulb, dew point
	// enthalpy and atm pressure using Bisection method.
	//
	const acc = 0.0001 // required accuracy
	rtb = tDewPoint    // initial guess
	dx = tDryBulb - tDewPoint
	for i := 0; i < 40; i++ { // max 40 attempts to resolve
		dx = dx * 0.5
		xmid = rtb + dx
		psat = calcVaporPressure(xmid)
		wstar = calcHumidityRatio(psat, pAtmosphere)
		fmid = calcEnthalpy(xmid, wstar)
		if (hEnthalpy - fmid) > 0.0 {
			rtb = xmid
		}
		if math.Abs(dx) < acc {
			break // convergence found
		}
	}
	return rtb
}
