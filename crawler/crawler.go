package crawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aaaasmile/crawler/conf"
	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/idl"
	"github.com/aaaasmile/crawler/mail"

	"github.com/gocolly/colly/v2"
)

type CrawlerOfChart struct {
	liteDB *db.LiteDB
	list   []*idl.ChartInfo
}

func (cc *CrawlerOfChart) Start(configfile string) error {
	conf.ReadConfig(configfile)
	log.Println("Configuration is read")

	cc.list = make([]*idl.ChartInfo, 0)
	cc.liteDB = &db.LiteDB{
		DebugSQL:     conf.Current.DebugSQL,
		SqliteDBPath: conf.Current.DBPath,
	}

	if err := cc.liteDB.OpenSqliteDatabase(); err != nil {
		return err
	}

	if err := cc.buildTheChartList(); err != nil {
		return err
	}
	if err := cc.sendChartEmail(); err != nil {
		return err
	}

	return nil
}

func (cc *CrawlerOfChart) buildTheChartList() error {
	cc.list = make([]*idl.ChartInfo, 0)
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", ImgURI: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 2", ImgURI: "data/chart_01.png"})

	return nil
}

func (cc *CrawlerOfChart) sendChartEmail() error {

	log.Println("Send email with num of items", len(cc.list))

	mm, err := mail.NewMailSender(cc.liteDB)
	if err != nil {
		return err
	}

	templFileName := "templates/chart-mail.html"
	if err := mm.SendEmailOAUTH2(templFileName, cc.list); err != nil {
		return err
	}

	return nil
}

func pickPicture(URL string) error {
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
	// if err := sendChartMail(); err != nil {
	// 	log.Println("Error on sending mail: ", err)
	// }

	return nil
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
