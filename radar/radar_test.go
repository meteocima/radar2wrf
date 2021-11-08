package radar

import (
	"bufio"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertReadLine(t *testing.T, expected string, buf *bufio.Reader) {
	res, err := buf.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, expected, string(res))
}

func Test(t *testing.T) {
	//err := oldradar.Cappi2ascii()
	reader, err := Convert("/home/parroit/Desktop/cimarepos/radar2wrf/data", "", "2020072000")
	assert.NoError(t, err)
	buf := bufio.NewReader(reader)

	assertReadLine(t, "TOTAL NUMBER =  1\n", buf)
	assertReadLine(t, "#-----------------#\n", buf)
	assertReadLine(t, "\n", buf)
	assertReadLine(t, "RADAR               17.920   47.570    100.0  2020-07-20_00:00:00   1520289    3\n", buf)
	assertReadLine(t, "#-------------------------------------------------------------------------------#\n", buf)
	assertReadLine(t, "\n", buf)
	assertReadLine(t, "FM-128 RADAR   2020-07-20_00:00:00        47.570         5.600     100.0       3\n", buf)
	assertReadLine(t, "         2000.0 -888888.000 -88 -888888.000   -888888.000 -88 -888888.000\n", buf)
}
