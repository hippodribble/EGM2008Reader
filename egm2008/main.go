package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hippodribble/egmreader"
)

var egmfile string
var lat, lon float64
var locationfile string
var latcolumn, longcolumn int

var d = 1.0 / 60.0
var minlat, minlon, maxlat, maxlon, dlat, dlon float64

func main() {

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`
egm2008 - get geoidal height/s from the EGM2008 grid.

The EGM2008 data are assumed to be stored locally (by default in the user's home/.egm2008 folder). 
They are expected to be in a large 16-bit PGM file with "P5" in the first line.
The location of the file can be overridden if it is not in the default location.
The data are sampled every 1' of arc from -180° to +180° longitude, -90° to +90° latitude.

This command can return:

- the height for a single location
- a list of heights for a grid of locations with a range and spacing of longitude and latitude

If the spacing of both latitude and longitude is set to zero (default), then a single location is expected.

If a location file is specified, then lat/long values are drawn from that (be sure to specify the relevant columns)


- Glenn Reynolds 2026 -


Usage:
  egm2008 [options]

Options:
`)
		flag.PrintDefaults()
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("you don't have a home directory - you're weird")
	}

	defaultfile := filepath.Join(homedir, ".egm2008", "egm2008-1.pgm")

	flag.StringVar(&egmfile, "f", defaultfile, "location of 16-bit EGM2008 PGM file")

	flag.Float64Var(&lat, "lat", 0, "latitude (-90° to +90°)")
	flag.Float64Var(&lon, "lon", 0, "latitude (-180° to +180°)")
	flag.Float64Var(&minlon, "minlon", 0, "minimum longitude (-180° to +180°) for a grid of data")
	flag.Float64Var(&maxlon, "maxlon", 0, "maximum longitude (-180° to +180°) for a grid of data")
	flag.Float64Var(&dlon, "dlon", d, "longitude spacing (in degrees) for a grid of data")
	flag.Float64Var(&minlat, "minlat", 0, "minimum latitude (-90° to +90°) for a grid of data")
	flag.Float64Var(&maxlat, "maxlat", 0, "maximum latitude (-90° to +90°) for a grid of data")
	flag.Float64Var(&dlat, "dlat", d, "latitude spacing (in degrees) for a grid of data")

	flag.StringVar(&locationfile, "locationfile", "", "use a CSV file to provide locations to get their geoidal heights (specify the long/lat columns)")
	flag.IntVar(&latcolumn, "latcolumn", 2, "column in optional CSV file containing latitudes")
	flag.IntVar(&longcolumn, "longcolumn", 3, "column in optional CSV file containing longitudes")

	flag.Parse()

	r, err := egmreader.NewEGM2008Reader(egmfile)
	if err != nil {
		log.Fatalln(err)
	}

	if locationfile != "" {
		list, err := readCSV(locationfile, latcolumn, longcolumn)
		if err != nil {
			log.Fatalln(err)
		}
		newlist, err := r.List(list)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Longitude,Latitude,Geoidal Height")
		for _, row := range newlist {
			fmt.Printf("%9.4f,%9.4f,%6.2f\n", row[0], row[1], row[2])
		}
		os.Exit(0)
	}

	if dlat == 0 && dlon == 0 {
		h, err := r.At(lon, lat)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Longitude,Latitude,Geoidal Height")
		fmt.Printf("%9.4f,%9.4f,%6.2f\n", lon, lat, h)
		r.Close()
	} else {
		list, err := r.Grid(minlon, maxlon, minlat, maxlat, dlon, dlat)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Longitude,Latitude,Geoidal Height")
		for _, row := range list {
			fmt.Printf("%9.4f,%9.4f,%6.2f\n", row[0], row[1], row[2])
		}
	}
}
