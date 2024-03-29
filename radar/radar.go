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

// ReadDoubleVar ...
func (data *CappiDataset) ReadDoubleVar(name string) []float64 {
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

	varval := make([]float64, varlen)

	err = res.ReadFloat64s(varval)
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
	//fmt.Printf("OPENING FILE %s: %s\n", filename, err)
	data.ds, data.err = &ds, err
}

// Dimensions ...
type Dimensions struct {
	Lat                                                    []float32
	Lon                                                    []float32
	Width                                                  int64
	Height                                                 int64
	Instants                                               []time.Time
	Cappi2, Cappi3, Cappi4, Cappi5, Cappi6, Cappi7, Cappi8 []float32
}

func filenameForVar(dirname, varname, dt string) string {
	pt := fmt.Sprintf("%s/%s-%s.nc", dirname, dt, varname)
	return pt
}

func writeRadarData(f io.Writer, val float32, height float64) {

	if val < 0 {
		// write(301,'(3x,f12.1,2(f12.3,i4,f12.3,2x))')
		// hgt(i,m), rv_data(i,m), rv_qc(i,m), rv_err(i,m), rf_data(i,m), rf_qc(i,m), rf_err(i,m)
		fmt.Fprintf(f, "   %12.1f -888888.000 -88 -888888.000   -888888.000 -88 -888888.000\n", height)
		return
	}

	fmt.Fprintf(
		f,
		// write(301,'(3x,f12.1,2(f12.3,i4,f12.3,2x))')
		// hgt(i,m), rv_data(i,m), rv_qc(i,m), rv_err(i,m), rf_data(i,m), rf_qc(i,m), rf_err(i,m)
		"   %12.1f -888888.000 -88 -888888.000  %12.3f   0       5.000\n",
		height,
		val,
	)

}

func writeConvertedDataTo(resultW io.WriteCloser, dims *Dimensions, dtRequested time.Time) {

	defer resultW.Close()
	result := bufio.NewWriterSize(resultW, 1000000)
	defer result.Flush()

	maxLon := float32(-1)
	maxLat := float32(-1)
	instant := dtRequested.Format("2006-01-02_15:04")
	totObs := 0

	if dims.Cappi2 != nil ||
		dims.Cappi3 != nil ||
		dims.Cappi4 != nil ||
		dims.Cappi5 != nil ||
		dims.Cappi6 != nil ||
		dims.Cappi7 != nil ||
		dims.Cappi8 != nil {
		for _, l := range dims.Lon {
			if l > maxLon {
				maxLon = l
			}
		}
		for _, l := range dims.Lat {
			if l > maxLat {
				maxLat = l
			}
		}
	} else {
		maxLon = float32(1)
		maxLat = float32(1)
	}

	for i := int64(0); i < dims.Width*dims.Height; i++ {
		f2 := float32(-1)
		f3 := float32(-1)
		f4 := float32(-1)
		f5 := float32(-1)
		f6 := float32(-1)
		f7 := float32(-1)
		f8 := float32(-1)

		if dims.Cappi2 != nil {
			f2 = dims.Cappi2[i]
		}

		if dims.Cappi3 != nil {
			f3 = dims.Cappi3[i]
		}

		if dims.Cappi4 != nil {
			f4 = dims.Cappi4[i]
		}

		if dims.Cappi5 != nil {
			f5 = dims.Cappi5[i]
		}

		if dims.Cappi6 != nil {
			f6 = dims.Cappi6[i]
		}

		if dims.Cappi7 != nil {
			f7 = dims.Cappi7[i]
		}

		if dims.Cappi8 != nil {
			f8 = dims.Cappi8[i]
		}

		if f2 >= 0 || f3 >= 0 || f4 >= 0 || f5 >= 0 || f6 >= 0 || f7 >= 0 || f8 >= 0 {
			totObs++
		}
	}

	fmt.Fprintf(result, "TOTAL NUMBER =  1\n")
	fmt.Fprintf(result, "#-----------------#\n")
	fmt.Fprintf(result, "\n")
	//  write(301,'(a5,2x,a12,2(f8.3,2x),f8.1,2x,a19,2i6)') 'RADAR', &
	//  radar_name, rlonr(irad), rlatr(irad), raltr(irad)*1000., &
	//  trim(radar_date), np, imdv_nz(irad)
	fmt.Fprintf(result, "RADAR              %8.3f  %8.3f     100.0  %s:00%6d     4\n",
		maxLon,
		maxLat,
		instant,
		totObs,
	)

	fmt.Fprintf(result, "#-------------------------------------------------------------------------------#\n")
	fmt.Fprintf(result, "\n")

	if dims.Cappi2 == nil &&
		dims.Cappi3 == nil &&
		dims.Cappi4 == nil &&
		dims.Cappi5 == nil &&
		dims.Cappi6 == nil &&
		dims.Cappi7 == nil &&
		dims.Cappi8 == nil {
		return
	}

	instant = dims.Instants[0].Format("2006-01-02_15:04")

	for x := int64(0); x < dims.Width; x++ {
		for y := int64(dims.Height) - 1; y >= int64(0); y-- {
			i := x + y*dims.Width

			lat := dims.Lat[i]
			lon := dims.Lon[i]

			f2 := float32(-1)
			f3 := float32(-1)
			f4 := float32(-1)
			f5 := float32(-1)
			f6 := float32(-1)
			f7 := float32(-1)
			f8 := float32(-1)

			if dims.Cappi2 != nil {
				f2 = dims.Cappi2[i]
			}
			if dims.Cappi3 != nil {
				f3 = dims.Cappi3[i]
			}
			if dims.Cappi4 != nil {
				f4 = dims.Cappi4[i]
			}
			if dims.Cappi5 != nil {
				f5 = dims.Cappi5[i]
			}
			if dims.Cappi6 != nil {
				f6 = dims.Cappi6[i]
			}
			if dims.Cappi7 != nil {
				f7 = dims.Cappi7[i]
			}
			if dims.Cappi8 != nil {
				f8 = dims.Cappi8[i]
			}

			if f2 >= 0 || f3 >= 0 || f4 >= 0 || f5 >= 0 || f6 >= 0 || f7 >= 0 || f8 >= 0 {
				fmt.Fprintf(
					result,
					//!----Write data
					//do i = 1,np ! np: # of total horizontal data points
					//write(301,'(a12,3x,a19,2x,2(f12.3,2x),f8.1,2x,i6)') 'FM-128 RADAR', &
					// trim(radar_date), plat(i), plon(i), raltr(irad)*1000, count_nz(i)

					//"FM-128 RADAR   %s:00       %7.3f      %8.3f     100.0       3\n",
					"FM-128 RADAR   %s:00  %12.3f  %12.3f     100.0       4\n",
					instant,
					lat,
					lon)

				if dims.Cappi2 != nil {
					writeRadarData(result, f2, 2000.0)
				}
				if dims.Cappi3 != nil {
					writeRadarData(result, f3, 3000.0)
				}
				if dims.Cappi4 != nil {
					writeRadarData(result, f4, 4000.0)
				}
				if dims.Cappi5 != nil {
					writeRadarData(result, f5, 5000.0)
				}
				if dims.Cappi6 != nil {
					writeRadarData(result, f6, 6000.0)
				}
				if dims.Cappi7 != nil {
					writeRadarData(result, f7, 7000.0)
				}
				if dims.Cappi8 != nil {
					writeRadarData(result, f8, 8000.0)
				}
			}
		}
	}
}

// Convert ...
func Convert(dirname, radarOutFileName, dt string) (io.Reader, error) {
	dims := Dimensions{}
	ds := &CappiDataset{}

	setDims := func() {
		if dims.Width > 0 {
			return
		}
		if ds.ReadFloatVar("latitude"); ds.err != nil {
			ds.err = nil
			return
		}
		dims.Width = int64(ds.GetDimensionLen("cols"))
		dims.Height = int64(ds.GetDimensionLen("rows"))
		dims.Lat = ds.ReadFloatVar("latitude")
		dims.Lon = ds.ReadFloatVar("longitude")
		dims.Instants = ds.ReadTimeVar("time")
	}

	ds.Open(filenameForVar(dirname, "CAPPI2", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi2 = ds.ReadFloatVar("CAPPI2")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI3", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi3 = ds.ReadFloatVar("CAPPI3")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI4", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi4 = ds.ReadFloatVar("CAPPI4")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI5", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi5 = ds.ReadFloatVar("CAPPI5")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI6", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi6 = ds.ReadFloatVar("CAPPI6")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI7", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi7 = ds.ReadFloatVar("CAPPI7")
		setDims()
		ds.Close()
	}

	ds.Open(filenameForVar(dirname, "CAPPI8", dt))
	if ds.Error() == netcdf.Error(2) {
		ds.err = nil
		ds.Close()
		ds.err = nil
	} else {
		dims.Cappi8 = ds.ReadFloatVar("CAPPI8")
		setDims()
		ds.Close()
	}

	if ds.err != nil {
		return nil, ds.Error()
	}

	reader, result := io.Pipe()

	reqDt, err := time.Parse("200601021500", dt)
	if err == nil {
		go writeConvertedDataTo(result, &dims, reqDt)
	}
	return reader, err
}
