package main

import (
	"flag"
	"log"

	"github.com/aaaasmile/crawler/sometest/scrap"
	"github.com/aaaasmile/crawler/sometest/web"
)

func main() {
	var skipscrap = flag.Bool("skipscrap", false, "skip scrap if defined")
	var autofinish = flag.Bool("autofinish", false, "terminate when all scrap anconversions are finished")
	flag.Parse()
	stopch := make(chan struct{})
	msgch := make(chan string)

	go func() {
		web.StartServer(stopch)
		log.Println("server exit")
		msgch <- "OK"
	}()

	log.Println("Testing svg scraping and conversion")
	if !*skipscrap {
		if err := scrap.Scrap(); err != nil {
			log.Fatal("Scraping error ", err)
		}
	} else {
		log.Println("[WARN] scrap skipped")
	}

	if err := scrap.SaveToPng(); err != nil {
		log.Println("[ERR] error on save png ", err)
	}
	if *autofinish {
		log.Println("all stuff done")
		stopch <- struct{}{}
	}

	msg := <-msgch
	log.Println("terminate with: ", msg)
}
