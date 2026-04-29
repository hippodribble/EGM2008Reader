package egmreader

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/mmap"
)

type EGM2008PGMReader struct {
	FilePath      string
	fileoffset    int64
	lat0, long0   float64
	dLat, dLon    float64
	scale, offset float64
	nx, ny        int
	mappedFile    *mmap.ReaderAt
}

func NewEGM2008Reader(filename string) (*EGM2008PGMReader, error) {
	reader := &EGM2008PGMReader{
		FilePath: filename,
		long0:    0.0,
		lat0:     90.0,
		dLon:     1.0 / 60.0,
		dLat:     -1.0 / 60.0,
		scale:    0.003,
		offset:   -108,
	}
	err := reader.openPGMFile()
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// opens the underlying PGM file of EGM2008 data
func (R *EGM2008PGMReader) openPGMFile() error {
	r, err := os.Open(R.FilePath)
	if err != nil {
		return err
	}
	defer r.Close()
	sc := bufio.NewScanner(r)

	sc.Scan()
	if !strings.HasPrefix(sc.Text(), "P5") {
		return errors.New("Not a 16-bit EGM PGM file")
	}
	for {
		sc.Scan()
		if !strings.HasPrefix(sc.Text(), "#") {
			break
		}
	}
	f := strings.Fields(sc.Text())

	R.nx, err = strconv.Atoi(f[0])
	if err != nil {
		return err
	}
	R.ny, err = strconv.Atoi(f[1])
	if err != nil {
		return err
	}

	info, err := r.Stat()
	// fmt.Println("File size:", info.Size(), nx*ny*2)
	R.fileoffset = info.Size() - int64(2*R.nx*R.ny)

	r.Close()

	mappedFile, err := mmap.Open(R.FilePath)
	if err != nil {
		return err
	}
	R.mappedFile = mappedFile

	return nil
}

// closes the internal memory-mapped reader linked to the EGM2008 dataset
func (R *EGM2008PGMReader) Close() {
	R.mappedFile.Close()
}

// returns the grid indices of the requested lat/long as floats
// We can then interpolate these float positions to grid cell
// positions in order to get a value at the location.
//
//	Note the order of parameters - longitude, then latitude
//
// If the lat long values are out of range an error is returned.
// That is, the position is not wrapped around the sphere.
func (R EGM2008PGMReader) llToIndex(long, lat float64) (x float64, y float64, err error) {
	if lat < -90 || lat > 90 {
		return 0, 0, errors.New("latitude is out of range")
	}
	if long < -180 || long > 180 {
		return 0, 0, errors.New("longitude is out of range")
	}

	// long += 360
	// long = math.Mod(long, 180)

	x = (long - R.long0) / R.dLon
	y = (lat - R.lat0) / R.dLat
	err = nil
	return
}

// Converts a grid location in the EGM file to a lat/long location
// func (R EGM2008PGMReader) IndexToLongLat(i, j float64) (long, lat float64, err error) {
// 	if i < 0 || j < 0 {
// 		return 0, 0, errors.New("index is out of bounds")
// 	}
// 	if i > float64(R.nx-1) || j > float64(R.ny-1) {
// 		return float64(R.nx) - 1, float64(R.ny - 1), errors.New("is max - index is out of bounds")
// 	}
// 	long = i*R.dLon + R.long0
// 	lat = j * R.dLat * R.lat0
// 	return long, lat, nil
// }

// At gets the EGM height value at the specified linear index
//   - it returns an error if the index is out of global bounds
//   - the value is offset and scaled to a height in metres. This is the value to subtract from the GPS ellipsoid height to get an orthometric height
func (R EGM2008PGMReader) height(i int) (float64, error) {
	if i < 0 || i > R.nx*R.ny-1 {
		return 0, errors.New("index out of range")
	}
	b2 := make([]byte, 2)
	n, err := R.mappedFile.ReadAt(b2, int64(2*i)+R.fileoffset)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, fmt.Errorf("failed to read 16-bit value for index %d", i)
	}

	ub16 := float64(binary.BigEndian.Uint16(b2))
	return ub16*R.scale + R.offset, nil

}

// returns the value of the EGM2008 geoidal height at te requested location
// - lat/long - location of request
//
//	Returns an error if the location is outside -180 to +180° longitude, -90 to +90° latitude
func (R EGM2008PGMReader) At(long, lat float64) (float64, error) {
	i, j, err := R.llToIndex(long, lat)
	if err != nil {
		return 0, err
	}
	I := math.Floor(i)
	J := math.Floor(j)

	// fmt.Println(I, J, lat, long, " <")

	f := (I - i)
	g := (J - j)
	a, err := R.height(int(J)*R.nx + int(I))
	if err != nil {
		return 0, err
	}
	b, err := R.height(int(J+1)*R.nx + int(I))
	if err != nil {
		return 0, err
	}
	c, err := R.height(int(J)*R.nx + int(I+1))
	if err != nil {
		return 0, err
	}
	d, err := R.height(int(J+1)*R.nx + int(I+1))
	if err != nil {
		return 0, err
	}

	k := (1-f)*a + f*b
	l := (1-g)*c + f*d

	return k*(1-g) + l*g, nil
}

// Grid returns an X,Y,Z list of longitude,latitude,height  values corresponding to the requested grid
// - min/max long/lat define the limits
// - dlong/dlat define the spacing
//
//	If the location of any of the values is outside the global range, an error will be returned
//	i.e., outside -180 to +180° longitude, -90 to +90° latitude
func (R EGM2008PGMReader) Grid(minlong, maxlong, minlat, maxlat, dlong, dlat float64) ([][]float64, error) {

	dataout := [][]float64{}
	for i := minlong; i <= maxlong; i += dlong {
		for j := minlat; j <= maxlat; j += dlat {
			h, err := R.At(i, j)
			if err != nil {
				return nil, err
			}
			dataout = append(dataout, []float64{i, j, h})
		}
	}
	return dataout, nil
}

// Grid returns an X,Y,Z list of longitude,latitude,height  values corresponding to the requested list
// the list should contain [longitude,latitude] pairs in the first two columns
//
//	If the location of any of the values is outside the global range, an error will be returned
//	i.e., outside -180 to +180° longitude, -90 to +90° latitude
func (R EGM2008PGMReader) List(list [][]float64) ([][]float64, error) {

	dataout := [][]float64{}
	for _,row := range list {
		h, err := R.At(row[0], row[1])
		if err != nil {
			return nil, err
		}
		dataout = append(dataout, []float64{row[0], row[1], h})
	}
	return dataout, nil
}

func (R EGM2008PGMReader) probe(i int) (float64, error) {
	if i < 0 {
		return 0, errors.New("index out of range")
	}
	if i > R.nx*R.ny {
		return 0, errors.New("index out of range")
	}
	b2 := make([]byte, 2)
	n, err := R.mappedFile.ReadAt(b2, int64(2*i)+R.fileoffset)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, fmt.Errorf("failed to read 16-bit value for index %d", i)
	}
	bb := float64(binary.BigEndian.Uint16(b2))

	return bb*R.scale + R.offset, nil
}
