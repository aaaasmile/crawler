package main

import (
	"log"

	"github.com/aaaasmile/crawler/sometest/scrap"
	"github.com/aaaasmile/crawler/sometest/web"
)

func main() {
	log.Println("Testing svg scraping and conversion")
	if err := scrap.Scrap(); err != nil {
		log.Fatal("Scraping error ", err)
	}
	//scrap.Scrap2()
	web.StartServer()
}
