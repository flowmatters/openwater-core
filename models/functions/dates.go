package functions

/*OW-SPEC
DateGenerator:
	inputs:
		tick:
	states:
	parameters:
		startDate:
		startMonth:
		startYear:
	outputs:
		date:
		month:
		year:
		dayOfYear:
	implementation:
		function: dateGenerator
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		dates function
*/

var DAYS_IN_MONTH = [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func leapYear(y int) bool {
	if y%4 != 0 {
		return false
	}

	if y%100 != 0 {
		return true
	}

	if y%400 == 0 {
		return true
	}

	return false
}

func daysInMonth(month, year int) int {
	if month == 2 {
		if leapYear(year) {
			return 29
		}
	}
	return DAYS_IN_MONTH[month-1]
}

func _dayOfYear(d, m, y int) int {
	doy := 0
	for mi := 1; mi < m; mi++ {
		doy += daysInMonth(mi, y)
	}
	doy += d
	return doy
}

func dateGenerator(tick []float64,
	startDate, startMonth, startYear float64,
	date, month, year, dayOfYear []float64) {

	d := int(startDate)
	m := int(startMonth)
	y := int(startYear)

	n := len(tick)

	for i := 0; i < n; i++ {
		dayOfYear[i] = float64(_dayOfYear(d, m, y))
		date[i] = float64(d)
		month[i] = float64(m)
		year[i] = float64(y)

		d++

		if d > daysInMonth(m, y) {
			d = 1
			m++
		}

		if m > 12 {
			m = 1
			y++
		}
	}
}
