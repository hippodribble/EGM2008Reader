package egmreader

import (
	"fmt"
	"testing"
)

func TestFileOpen(t *testing.T) {
	Q, err := NewEGM2008Reader("/Users/glenn/Downloads/geoids/egm2008-1.pgm")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	defer Q.Close()
}
func TestLLtoIndex(t *testing.T) {
	Q, err := NewEGM2008Reader("/Users/glenn/Downloads/geoids/egm2008-1.pgm")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	defer Q.Close()

	fmt.Println(Q.FilePath, "dimensions:", Q.nx, Q.ny)
	lats := []float64{90}
	longs := []float64{0}

	for _, lat := range lats {
		for _, long := range longs {
			x, y, err := Q.llToIndex(long, lat)
			if err != nil {
				fmt.Printf("%7.1f,%7.1f: %8.2f,%8.2f %v\n", long, lat, x, y, err)
				continue
			}
			fmt.Printf("%7.1f,%7.1f: %8.2f,%8.2f\n", long, lat, x, y)
		}
	}
}

func TestHeightFromLatLong(t *testing.T) {
	Q, err := NewEGM2008Reader("/Users/glenn/Downloads/geoids/egm2008-1.pgm")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	defer Q.Close()

	points := make(map[string][]float64) // lat, long, height from GIS
	points["p1"] = []float64{38.31, 154.37, 5.955}
	points["p2"] = []float64{-48.7, 47, 45.105}
	points["p3"] = []float64{5.4, 67.6, -86.595}
	points["p4"] = []float64{-2.2, -148.3, 10.947}
	points["p5"] = []float64{41.996, 142.311, 20.184}
	points["0-0"] = []float64{0, 0, 17.226}

	for k, p := range points {
		h, err := Q.At(p[1], p[0])
		if err != nil {
			fmt.Println(k, err)
			continue
		}
		fmt.Println(k, p, h, "m")
	}
}

func TestProbeFile(t *testing.T) {
	Q, err := NewEGM2008Reader("/Users/glenn/Downloads/geoids/egm2008-1.pgm")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	defer Q.Close()
	indices := []int{0, 1, 2, 3, 4, 5, 6}
	for i := range 7 {
		indices = append(indices, i*10*Q.nx)
	}
	for _, index := range indices {
		f, err := Q.probe(index)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%3d: %9.2f\n", index, f)
	}
}
