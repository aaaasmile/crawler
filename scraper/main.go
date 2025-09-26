package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aaaasmile/crawler/scraper/scrap"
)

const (
	build_nr = "01.20250926.01.00"
)

// Used to download the svg and convert it to png
// Try it with: go run .\main.go -skipsave  -limit 1
// -screenshot if you want to capture the browser
func main() {
	start := time.Now()
	var ver = flag.Bool("ver", false, "Prints the current version")
	var dbpath = flag.String("dbpath", "../chart-info.db", "path to the db")
	var limit = flag.Int("limit", -1, "limit the scraping file, (-1 is all)")
	var screenshot = flag.Bool("screenshot", false, "take a screenshot of the chart page")
	var cookie = flag.Bool("cookie", true, "expect cookie to click away")
	flag.Parse()
	if *ver {
		fmt.Printf("Scraper version %s", build_nr)
		os.Exit(0)
	}

	log.Println("Svg scraping and Png conversion")
	sc := scrap.NewScrap(*screenshot, *cookie)
	if err := sc.Scrap(*dbpath, *limit); err != nil {
		log.Fatal("Scraping error ", err)
	}

	log.Println("processed files: ", sc.ReportProcessed())
	t := time.Now()
	elapsed := t.Sub(start)
	log.Printf("That's all folks. (elapsed time %v)\n", elapsed)
}
