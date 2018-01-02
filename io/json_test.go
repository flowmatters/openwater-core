package io

import (
	"math"
	"github.com/flowmatters/openwater-core/data"
	"testing"
	"github.com/stretchr/testify/assert"
)
// import "fmt"

func Test1D(t *testing.T) {
	arr := data.NewArray1D(10)
	arr.Set1(0,1.0)
	arr.Set1(1,math.NaN())
	arr.Set1(2,math.Inf(0))
	arr.Set1(3,math.Inf(-1))
	arr.Set1(4,2.0)

	jsonSafe := JsonSafeArray(arr,0)

	assert := assert.New(t)
	
	assert.Equal(10,len(jsonSafe))

	// Numbers good
	assert.Equal(1.0,jsonSafe[0])
	assert.Equal(2.0,jsonSafe[4])

	assert.Equal("NaN",jsonSafe[1])
	assert.Equal("+Inf",jsonSafe[2])
	assert.Equal("-Inf",jsonSafe[3])
}

func Test2D(t *testing.T) {
	arr := data.NewArray2D(10,5)
	arr.Set2(0,0,1.0)
	arr.Set2(0,1,math.NaN())
	arr.Set2(1,0,math.Inf(0))
	arr.Set2(1,1,math.Inf(-1))
	arr.Set2(2,0,2.0)

	jsonSafe := JsonSafeArray(arr,0)

	assert := assert.New(t)
	
	assert.Equal(10,len(jsonSafe))

	row0,e := jsonSafe[0].([]interface{})
	if assert.True(e) {
		assert.Len(row0,5)
		assert.Equal(1.0,row0[0])
		assert.Equal("NaN",row0[1])
	}

	row1,e := jsonSafe[1].([]interface{})
	if assert.True(e) {
		assert.Len(row1,5)
		assert.Equal("+Inf",row1[0])
		assert.Equal("-Inf",row1[1])
	}

	row2,e := jsonSafe[2].([]interface{})
	if assert.True(e) {
		assert.Len(row2,5)
		assert.Equal(2.0,row2[0])
	}

	for i:=0; i<10; i++ {
		row,e := jsonSafe[i].([]interface{})
		if assert.True(e) {
			assert.Len(row,5)
		}
	}
}	