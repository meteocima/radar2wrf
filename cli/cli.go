package main

import (
	"io"
	"log"
	"os"

	"github.com/meteocima/radar2wrf/radar"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: r2w <inputdir> YYYYMMDDHH")
	}
	reader, err := radar.Convert(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	outfile, err := os.OpenFile("ob.radar."+os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	_, err = io.Copy(outfile, reader)
	if err != nil {
		log.Fatal(err)
	}
}
