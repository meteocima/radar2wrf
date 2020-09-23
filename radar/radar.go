package radar

import (
	"fmt"
	"io"
	"time"

	"github.com/fhs/go-netcdf/netcdf"
)

// CappiDataset ...
type CappiDataset struct {
	ds  *netcdf.Dataset
	err error
}

// Error ...
func (data *CappiDataset) Error() error {
	return data.err
}

// GetDimensionLen ...
func (data *CappiDataset) GetDimensionLen(name string) uint64 {
	if data.err != nil {
		return 0
	}

	dim, err := data.ds.Dim(name)
	if data.err != nil {
		data.err = err
		return 0
	}

	dimlen, err := dim.Len()
	if data.err != nil {
		data.err = err
		return 0
	}

	return dimlen
}

// Close ...
func (data *CappiDataset) Close() {
	if data.err != nil {
		return
	}
	if data.ds == nil {
		panic("File closed")
	}

	data.err = data.ds.Close()
	data.ds = nil
}

// ReadFloatVar ...
func (data *CappiDataset) ReadFloatVar(name string) []float32 {
	if data.err != nil {
		return nil
	}
	res, err := data.ds.Var(name)
	if err != nil {
		data.err = err
		return nil
	}

	varlen, err := res.Len()
	if err != nil {
		data.err = err
		return nil
	}

	varval := make([]float32, varlen)

	err = res.ReadFloat32s(varval)
	if err != nil {
		data.err = err
		return nil
	}

	return varval
}

// ReadTimeVar ...
func (data *CappiDataset) ReadTimeVar(name string) []time.Time {
	if data.err != nil {
		return nil
	}
	varDs, err := data.ds.Var(name)
	if err != nil {
		data.err = err
		return nil
	}

	varlen, err := varDs.Len()
	if err != nil {
		data.err = err
		return nil
	}

	varval := make([]int32, varlen)

	err = varDs.ReadInt32s(varval)
	if err != nil {
		data.err = err
		return nil
	}

	res := make([]time.Time, varlen)
	for i, inst := range varval {
		res[i] = time.Unix(int64(inst), 0).UTC()
	}
	return res
}

// Open ...
func (data *CappiDataset) Open(filename string) {
	if data.err != nil {
		return
	}
	if data.ds != nil {
		panic("Already open")
	}

	ds, err := netcdf.OpenFile(filename, netcdf.NOWRITE)
	data.ds, data.err = &ds, err
}

// FlattenLanLot converts the original format of lat & lon in radar file
// in two `lat` and `lon' slices.
// In original netcdf radar file, lat & lon are specified for each point in the grid.
func (dims *Dimensions) FlattenLanLot(data *CappiDataset) {
	cols := data.GetDimensionLen("cols")
	rows := data.GetDimensionLen("rows")
	if data.Error() != nil {
		return
	}

	lats := make([]float32, rows)

	for i := uint64(0); i < rows; i++ {
		lats[i] = dims.Lat[i*cols]
	}

	dims.Lat = lats
	dims.Lon = dims.Lon[:rows]
}

// Dimensions ...
type Dimensions struct {
	Lat                    []float32
	Lon                    []float32
	Instants               []time.Time
	Cappi2, Cappi3, Cappi5 []float32
}

func filenameForVar(dirname, varname, dt string) string {
	return fmt.Sprintf("%s/%s-%s.nc", dirname, dt, varname)
}

func writeRadarData(f io.Writer, val float32, height float64) {
	if val < 0 {
		fmt.Fprintf(f, "       %8.1f -888888.000 -88 -888888.000   -888888.000 -88 -888888.000\n", height)
		return
	}

	fmt.Fprintf(
		f,
		"       %8.1f -888888.000 -88 -888888.000   %11.3f   0       5.000\n",
		height,
		val,
	)

}

func writeConvertedDataTo(result io.WriteCloser, dims *Dimensions) {
	maxLon := dims.Lat[len(dims.Lat)-1]
	maxLat := dims.Lon[len(dims.Lon)-1]

	instant := dims.Instants[0].Format("2006-01-02_15:04")

	fmt.Fprintf(result, "TOTAL NUMBER =  1\n")
	fmt.Fprintf(result, "#-----------------#\n")
	fmt.Fprintf(result, "\n")
	fmt.Fprintf(result, "RADAR             %8.3f  %7.3f    100.0  %s:00 %9d    3\n",
		maxLat,
		maxLon,
		instant,
		len(dims.Lat)*len(dims.Lon),
	)
	fmt.Fprintf(result, "#-------------------------------------------------------------------------------#\n")
	fmt.Fprintf(result, "\n")

	for x, lon := range dims.Lon {
		for y := len(dims.Lat) - 1; y >= 0; y-- {
			lat := dims.Lat[y]
			//for y, lat := range dims.Lat {
			f2 := dims.Cappi2[x+y*len(dims.Lon)]
			f3 := dims.Cappi3[x+y*len(dims.Lon)]
			f5 := dims.Cappi5[x+y*len(dims.Lon)]

			fmt.Fprintf(
				result,
				"FM-128 RADAR   %s:00       %7.3f      %8.3f     100.0       3\n",
				instant,
				lat,
				lon)

			writeRadarData(result, f2, 2000.0)
			writeRadarData(result, f3, 3000.0)
			writeRadarData(result, f5, 5000.0)
		}
	}

	result.Close()
}

// Convert ...
func Convert(dirname, dt string) (io.Reader, error) {
	ds := &CappiDataset{}
	ds.Open(filenameForVar(dirname, "CAPPI2", dt))

	dims := Dimensions{}

	dims.Lat = ds.ReadFloatVar("latitude")
	dims.Lon = ds.ReadFloatVar("longitude")
	dims.Instants = ds.ReadTimeVar("time")
	dims.FlattenLanLot(ds)

	dims.Cappi2 = ds.ReadFloatVar("CAPPI2")
	ds.Close()

	ds.Open(filenameForVar(dirname, "CAPPI3", dt))
	dims.Cappi3 = ds.ReadFloatVar("CAPPI3")
	ds.Close()

	ds.Open(filenameForVar(dirname, "CAPPI5", dt))
	dims.Cappi5 = ds.ReadFloatVar("CAPPI5")
	ds.Close()

	if ds.Error() != nil {
		return nil, ds.Error()
	}

	reader, result := io.Pipe()

	go writeConvertedDataTo(result, &dims)

	return reader, nil
}
