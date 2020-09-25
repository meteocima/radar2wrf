package radar

import (
	"bufio"
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

// Dimensions ...
type Dimensions struct {
	Lat                    []float32
	Lon                    []float32
	Width                  int64
	Height                 int64
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

func writeConvertedDataTo(resultW io.WriteCloser, dims *Dimensions, dtReq time.Time) {
	maxLon := float32(-1)

	for _, l := range dims.Lon {
		if l > maxLon {
			maxLon = l
		}
	}
	maxLat := float32(-1)
	for _, l := range dims.Lat {
		if l > maxLat {
			maxLat = l
		}
	}

	instant := dims.Instants[0].Format("2006-01-02_15:04")
	result := bufio.NewWriterSize(resultW, 1000000)

	totObs := 0
	for i := int64(0); i < dims.Width*dims.Height; i++ {
		f2 := dims.Cappi2[i]
		f3 := dims.Cappi3[i]
		f5 := dims.Cappi5[i]
		if f2 >= 0 || f3 >= 0 || f5 >= 0 {
			totObs++
		}
	}

	fmt.Fprintf(result, "TOTAL NUMBER =  1\n")
	fmt.Fprintf(result, "#-----------------#\n")
	fmt.Fprintf(result, "\n")
	fmt.Fprintf(result, "RADAR             %8.3f  %7.3f    100.0  %s:00 %9d    3\n",
		maxLon,
		maxLat,
		dtReq,
		totObs,
	)
	fmt.Fprintf(result, "#-------------------------------------------------------------------------------#\n")
	fmt.Fprintf(result, "\n")

	for x := int64(0); x < dims.Width; x++ {
		for y := int64(dims.Height) - 1; y >= int64(0); y-- {

			lat := dims.Lat[x+y*dims.Width]
			lon := dims.Lon[x+y*dims.Width]

			f2 := dims.Cappi2[x+y*dims.Width]
			f3 := dims.Cappi3[x+y*dims.Width]
			f5 := dims.Cappi5[x+y*dims.Width]
			if f2 >= 0 || f3 >= 0 || f5 >= 0 {
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
	}
	result.Flush()
	resultW.Close()
}

// Convert ...
func Convert(dirname, dt string) (io.Reader, error) {
	dtReq, err := time.Parse("2006010215", dt)
	if err != nil {
		return nil, err
	}
	ds := &CappiDataset{}
	ds.Open(filenameForVar(dirname, "CAPPI2", dt))

	dims := Dimensions{}

	dims.Width = int64(ds.GetDimensionLen("cols"))
	dims.Height = int64(ds.GetDimensionLen("rows"))
	dims.Lat = ds.ReadFloatVar("latitude")
	dims.Lon = ds.ReadFloatVar("longitude")
	dims.Instants = ds.ReadTimeVar("time")

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

	go writeConvertedDataTo(result, &dims, dtReq)

	return reader, nil
}
