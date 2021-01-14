package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaaasmile/crawler/crawler"
	"github.com/aaaasmile/live-omxctrl/web/idl"
)

func main() {
	var ver = flag.Bool("ver", false, "Prints the current version")
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	flag.Parse()

	if *ver {
		fmt.Printf("%s  version %s", idl.Appname, idl.Buildnr)
		os.Exit(0)
	}

	crw := crawler.CrawlerOfChart{}

	if err := crw.Start(*configfile); err != nil {
		panic(err)
	}
	log.Println("That's all folks.")
	os.Exit(0)
}
