package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

func readCSV(file string, latcol, longcol int) ([][]float64, error) {

	data := [][]float64{}

	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	csv := csv.NewReader(r)
	latcol -= 1
	longcol -= 1

	for {
		fields, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(fields[longcol], 64)
		if err != nil {
			continue
		}
		lat, err := strconv.ParseFloat(fields[latcol], 64)
		if err != nil {
			continue
		}
		data = append(data, []float64{lon, lat})
	}
	return data, nil
}
