package conversion

/*OW-SPEC
DeliveryRatio:
  symbol: DR
  inputs:
		input:
	states:
	parameters:
		fraction: 'default=1'
	outputs:
		output:
	implementation:
		function: applyScaling
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		partition
*/
