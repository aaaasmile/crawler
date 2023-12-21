package main

import (
	"flag"
	"log"
	"time"

	"github.com/aaaasmile/crawler/scraper/scrap"
	"github.com/aaaasmile/crawler/scraper/web"
)

func main() {
	start := time.Now()
	var skipscrap = flag.Bool("skipscrap", false, "skip scrap if defined")
	var noautofinish = flag.Bool("noautofinish", false, "avoid termination when all scraps anconversions are finished (to inspect the web server)")
	var skipsave = flag.Bool("skipsave", false, "skip save to png if defined")
	var dbpath = flag.String("dbpath", "../chart-info.db", "path to the db")
	var limit = flag.Int("limit", -1, "limit the scraping file, (-1 is all)")
	var screenshot = flag.Bool("screenshot", false, "take a screenshot of the chart page")
	flag.Parse()
	stopch := make(chan struct{})
	msgch := make(chan string)

	go func() {
		web.StartServer(stopch)
		log.Println("server exit")
		msgch <- "OK"
	}()

	log.Println("Svg scraping and Png conversion")
	sc := scrap.Scrap{
		TakeScreenshot: *screenshot,
	}
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
