package main

import (
	"log"
	"os"

	"github.com/meteocima/radar2wrf/radar"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: r2w <inputdir> YYYYMMDDHH")
	}
	radar.Convert(os.Args[1], os.Args[2])
}
