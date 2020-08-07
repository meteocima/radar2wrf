package radar

import (
	"fmt"
	"io"
	"time"

	"github.com/fhs/go-netcdf/netcdf"
)

func readDims(dirname string) ([]float32, []float32, []time.Time, error) {
	filename := fmt.Sprintf("%s/CAPPI2.nc", dirname)

	data, err := netcdf.OpenFile(filename, netcdf.NOWRITE)
	if err != nil {
		return nil, nil, nil, err
	}
	defer data.Close()

	colsDim, err := data.Dim("cols")
	if err != nil {
		return nil, nil, nil, err
	}

	cols, err := colsDim.Len()
	if err != nil {
		return nil, nil, nil, err
	}

	rowsDim, err := data.Dim("rows")
	if err != nil {
		return nil, nil, nil, err
	}

	rows, err := rowsDim.Len()
	if err != nil {
		return nil, nil, nil, err
	}

	latvar, err := data.Var("latitude")
	if err != nil {
		return nil, nil, nil, err
	}

	latval := make([]float32, rows*cols)
	err = latvar.ReadFloat32s(latval)
	if err != nil {
		return nil, nil, nil, err
	}

	lonvar, err := data.Var("longitude")
	if err != nil {
		return nil, nil, nil, err
	}

	lonval := make([]float32, rows*cols)
	err = lonvar.ReadFloat32s(lonval)
	if err != nil {
		return nil, nil, nil, err
	}

	lats := make([]float32, rows)

	for i := uint64(0); i < rows; i++ {
		lats[i] = latval[i*cols]
	}

	lons := lonval[:rows]
	return lats, lons, []time.Time{
		time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
	}, nil
}

func readVarFile(dirname, varname string) ([]float32, error) {
	filename := fmt.Sprintf("%s/%s.nc", dirname, varname)

	data, err := netcdf.OpenFile(filename, netcdf.NOWRITE)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	cappivar, err := data.Var(varname)
	if err != nil {
		return nil, err
	}

	rowsDim, err := data.Dim("rows")
	if err != nil {
		return nil, err
	}

	colsDim, err := data.Dim("cols")
	if err != nil {
		return nil, err
	}

	rows, err := rowsDim.Len()
	if err != nil {
		return nil, err
	}

	cols, err := colsDim.Len()
	if err != nil {
		return nil, err
	}

	values := make([]float32, rows*cols)
	err = cappivar.ReadFloat32s(values)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// Convert ...
func Convert(dirname string) (io.Reader, error) {
	cappi2, err := readVarFile(dirname, "CAPPI2")
	if err != nil {
		return nil, err
	}

	cappi3, err := readVarFile(dirname, "CAPPI3")
	if err != nil {
		return nil, err
	}

	cappi5, err := readVarFile(dirname, "CAPPI5")
	if err != nil {
		return nil, err
	}

	lat, lon, instants, err := readDims(dirname)
	if err != nil {
		return nil, err
	}

	_, _, _ = cappi2, cappi3, cappi5

	reader, result := io.Pipe()
	maxLon, maxLat := lat[len(lat)-1], lon[len(lon)-1]
	go func() {
		fmt.Fprintf(result, "TOTAL NUMBER =  1\n")
		fmt.Fprintf(result, "#-----------------#\n")
		fmt.Fprintf(result, "\n")
		//fmt.Fprintf(result, "lat (%v-%v) lon(%v-%v)\n", lat[0], lat[len(lat)-1], lon[0], lon[len(lon)-1])
		fmt.Fprintf(result, "RADAR             %8.3f  %7.3f    100.0  %s:00 %9d    3\n",
			maxLat,
			maxLon,
			instants[0].Format("2006-01-02_15:04"),
			len(lat)*len(lon),
		)
		result.Close()
	}()

	return reader, nil
}
