package main

import (
	"flag"
	"log"

	"github.com/aaaasmile/crawler/sometest/scrap"
	"github.com/aaaasmile/crawler/sometest/web"
)

func main() {
	var skipscrap = flag.Bool("skipscrap", false, "skip scrap if defined")

	flag.Parse()
	log.Println("Testing svg scraping and conversion")
	if !*skipscrap {
		if err := scrap.Scrap(); err != nil {
			log.Fatal("Scraping error ", err)
		}
	} else {
		log.Println("[WARN] scrap skipped")
	}

	web.StartServer()
}
