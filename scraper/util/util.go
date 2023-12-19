package util

import (
	"fmt"
	"path/filepath"
)

const (
	datadir = "static/data/"
)

func GetChartSVGFileName(id int) string {
	svg_filename := fmt.Sprintf("chart%02d.svg", id)
	svg_fullfilename := filepath.Join(datadir, svg_filename)
	return svg_fullfilename
}
