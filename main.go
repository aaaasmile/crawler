package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		alt := e.Attr("alt")
		// Print link
		if strings.HasPrefix(link, "getChart") {
			fmt.Printf("Image found: %q -> %s - alt: %s\n", e.Text, link, alt)
		}

		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		//c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	//c.Visit("https://invido.it/")
	c.Visit("https://www.easycharts.at/index.asp?action=securities_chart&actionTypeID=0_2&typeID=99&id=tts-11057070&menuId=1&pathName=XTR%2EDBLCI+CO%2EO%2EY%2ESW%2E1CEOH")
}
