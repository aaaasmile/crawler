package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaaasmile/crawler/crawler"
	"github.com/aaaasmile/crawler/idl"
)

func main() {
	var ver = flag.Bool("ver", false, "Prints the current version")
	var simulate = flag.Bool("simulate", false, "Simulate email send (build the message without sending it)")
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	var resendmail = flag.Bool("resendmail", false, "Resend email using the last data download")

	flag.Parse()

	if *ver {
		fmt.Printf("%s  version %s", idl.Appname, idl.Buildnr)
		os.Exit(0)
	}

	crw := crawler.CrawlerOfChart{
		Simulate:    *simulate,
		ResendEmail: *resendmail,
	}

	if err := crw.Start(*configfile); err != nil {
		panic(err)
	}
	log.Println("That's all folks.")
}
