package util

import (
	"fmt"
	"path/filepath"
)

const (
	datadir    = "static/data/"
	datapngdir = "../data/"
)

func GetChartSVGFullFileName(id int) string {
	svg_filename := GetChartSVGFileNameOnly(id)
	svg_fullfilename := filepath.Join(datadir, svg_filename)
	return svg_fullfilename
}

func GetChartSVGFileNameOnly(id int) string {
	svg_filename := fmt.Sprintf("chart%02d.svg", id)
	return svg_filename
}

func GetChartPNGFullFileName(id int) string {
	png_filename := GetChartPNGFileNameOnly(id)
	png_fullfilename := filepath.Join(datapngdir, png_filename)
	return png_fullfilename
}

func GetChartPNGFileNameOnly(id int) string {
	png_filename := fmt.Sprintf("chart_%d.png", id)
	return png_filename
}
