package util

func MapStoF(series []string, f func(string) float64) []float64 {
  res := make([]float64, len(series))

  for i, row := range series{
    res[i] = f(row)
  }

  return res;
}

func Map(records [][]string, f func([]string) string) []string{
  res := make([]string,len(records))

  for i,row := range records {
    res[i] = f(row)
  }

  return res;
}

func Extract(i int) func([]string)string {
    return func(row []string) string{
      return row[i];
    };
  }
  