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
	var simulate = flag.Bool("simulate", false, "Simulate email send")
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	var resendmail = flag.Bool("resendmail", false, "Resend email with the last downloaded data")
	var usedbtoken = flag.Bool("dbtoken", false, "Use the refresh and auth token stored into the db (gmail)")
	var useserviceaccount = flag.Bool("useserviceaccount", false, "Use service account credential (gsuite)")

	flag.Parse()

	if *ver {
		fmt.Printf("%s  version %s", idl.Appname, idl.Buildnr)
		os.Exit(0)
	}

	crw := crawler.CrawlerOfChart{
		Simulate:          *simulate,
		ResendEmail:       *resendmail,
		UseDBToken:        *usedbtoken,
		UseServiceAccount: *useserviceaccount,
	}

	if err := crw.Start(*configfile); err != nil {
		panic(err)
	}
	log.Println("That's all folks.")
	os.Exit(0)
}
