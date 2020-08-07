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
	reader, err := Convert("/home/parroit/cimarepos/radar2wrf/data")
	assert.NoError(t, err)
	buf := bufio.NewReader(reader)

	assertReadLine(t, "TOTAL NUMBER =  1\n", buf)
	assertReadLine(t, "#-----------------#\n", buf)
	assertReadLine(t, "\n", buf)
	assertReadLine(t, "RADAR               17.920   47.570    100.0  2020-07-20_00:00:00   1520289    3\n", buf)

}
