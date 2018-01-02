package io

import (
	csv "encoding/csv"
	"os"
	"github.com/flowmatters/openwater-core/util"
)

type TimeSeriesTable map[string] []float64

func ReadTimeSeriesCSVFile(f *os.File) (TimeSeriesTable,error){
  reader := csv.NewReader(f);
  records,_ := reader.ReadAll();

	result := make(TimeSeriesTable)

	keys := records[0][1:]
	records = records[1:]
	for i,key := range(keys){
		values := util.MapStoF(util.Map(records,util.Extract(i+1)),util.ParseFloatNaN);
		result[key] = values
	}

	return result,nil
}

func ReadTimeSeriesCSV(fn string) (TimeSeriesTable,error){
	file,err := os.Open(fn);
	defer file.Close()

	if err != nil {
		return nil,err
	}

	return ReadTimeSeriesCSVFile(file)
}