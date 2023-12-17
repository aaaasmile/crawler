package main

import (
	"log"

	"github.com/aaaasmile/crawler/sometest/scrap"
	"github.com/aaaasmile/crawler/sometest/web"
)

func main() {
	log.Println("Testin svg scraping and conversion")
	scrap.Scrap()
	web.StartServer()
}
