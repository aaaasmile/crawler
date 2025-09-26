package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aaaasmile/crawler/scraper/scrap"
	"github.com/aaaasmile/crawler/scraper/web"
)

// Used to download the svg and convert it to png
// Try it with: go run .\main.go -skipsave  -limit 1
// -screenshot if you want to capture the browser
func main() {
	start := time.Now()
	var ver = flag.Bool("ver", false, "Prints the current version")
	var skipscrap = flag.Bool("skipscrap", false, "skip scrap if defined")
	var noautofinish = flag.Bool("noautofinish", false, "avoid termination when all scraps anconversions are finished (to inspect the web server)")
	var skipsave = flag.Bool("skipsave", false, "skip save to png if defined")
	var dbpath = flag.String("dbpath", "../chart-info.db", "path to the db")
	var limit = flag.Int("limit", -1, "limit the scraping file, (-1 is all)")
	var screenshot = flag.Bool("screenshot", false, "take a screenshot of the chart page")
	var cookie = flag.Bool("cookie", true, "expect cookie to click away")
	flag.Parse()
	if *ver {
		fmt.Printf("Scraper version %s", web.Buildnr)
		os.Exit(0)
	}

	stopch := make(chan struct{})
	msgch := make(chan string)

	go func() {
		web.StartServer(stopch)
		log.Println("server exit")
		msgch <- "OK"
	}()

	log.Println("Svg scraping and Png conversion")
	sc := scrap.NewScrap(*screenshot, *cookie)
	if !*skipscrap {
		if err := sc.Scrap(*dbpath, *limit); err != nil {
			log.Fatal("Scraping error ", err)
		}
	} else {
		log.Println("[WARN] scrap skipped")
		sc.PrepareTestSVG()
	}
	if !*skipsave {
		if err := sc.SaveToPng(); err != nil {
			log.Println("[ERR] error on save png ", err)
		}
	}
	blocked := *noautofinish
	if !blocked {
		log.Println("all stuff done")
		stopch <- struct{}{}
	}

	msg := <-msgch
	log.Println("processed files: ", sc.ReportProcessed())
	log.Println("terminate with: ", msg)
	t := time.Now()
	elapsed := t.Sub(start)
	log.Printf("That's all folks. (elapsed time %v)\n", elapsed)
}
