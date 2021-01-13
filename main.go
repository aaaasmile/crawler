package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	//c.Visit("https://www.easycharts.at/index.asp?action=securities_chart&actionTypeID=0_2&typeID=99&id=tts-11057070&menuId=1&pathName=XTR%2EDBLCI+CO%2EO%2EY%2ESW%2E1CEOH")
	// if err := downloadFile("https://www.easycharts.at/getChart.asp?action=getChart&chartID=2BD8F3179A7535A822B51F6C72B0CAD80F40442A", "data/chart_01.png"); err != nil {
	// 	log.Println("Error on download file: ", err)
	// }
	if err := sendChartMail(); err != nil {
		log.Println("Error on sending mail: ", err)
	}
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
