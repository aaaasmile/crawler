package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	datasvgdir = "static/data/"
	datapngdir = "../data/"
)

func GetChartSVGFullFileName(id int) string {
	svg_filename := GetChartSVGFileNameOnly(id)
	svg_fullfilename := filepath.Join(datasvgdir, svg_filename)
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

func removeGlob(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
}

func CleanSVGPNGData() error {
	log.Println("Clean up data dir (remove png and svg files)")
	{
		svg_filter := fmt.Sprintf("%s*.svg", datasvgdir)
		if err := removeGlob(svg_filter); err != nil {
			return err
		}
	}
	{
		png_filter := fmt.Sprintf("%s*.png", datapngdir)
		if err := removeGlob(png_filter); err != nil {
			return err
		}
	}
	return nil
}
