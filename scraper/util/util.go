package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	datapngdir = "../data/"
)

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

func CleanPNGData() error {
	log.Println("Clean up data dir (remove png files)")
	{
		png_filter := fmt.Sprintf("%s*.png", datapngdir)
		if err := removeGlob(png_filter); err != nil {
			return err
		}
	}
	return nil
}
