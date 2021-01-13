package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaaasmile/crawler/crawler"
)

func main() {
	var ver = flag.Bool("ver", false, "Prints the current version")
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	flag.Parse()

	if *ver {
		fmt.Println("Crawler version 0.1.0")
		os.Exit(0)
	}

	if err := crawler.Start(*configfile); err != nil {
		panic(err)
	}
	log.Println("That's it folks.")
	os.Exit(0)
}
