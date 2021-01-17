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
	liteDB   *db.LiteDB
	list     []*idl.ChartInfo
	Simulate bool
}

type InfoChart struct {
	Error   error
	FileDst string
	Text    string
	Alt     string
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
	stockList, err := cc.liteDB.FetchStockInfo(2)
	if err != nil {
		return err
	}

	chRes := make(chan *InfoChart)
	for _, v := range stockList {
		go pickPicture(v.ChartURL, v.ID, chRes)
	}

	var res *InfoChart
	for range stockList {
		res = <-chRes
	}
	return nil
}

func (cc *CrawlerOfChart) fillWithSomeDummy() {
	// example without the crawler
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", Fullname: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", Fullname: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", Fullname: "data/chart_01.png"})
}

func (cc *CrawlerOfChart) sendChartEmail() error {

	log.Println("Send email with num of items", len(cc.list))

	mm, err := mail.NewMailSender(cc.liteDB, cc.Simulate)
	if err != nil {
		return err
	}

	templFileName := "templates/chart-mail.html"
	if err := mm.SendEmailOAUTH2(templFileName, cc.list); err != nil {
		return err
	}

	return nil
}

func pickPicture(URL string, ix int, chItem chan *InfoChart) error {
	c := colly.NewCollector()
	found := false
	// On every a element which has href attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		alt := e.Attr("alt")
		if strings.HasPrefix(link, "getChart") {
			fileNameDst := fmt.Sprintf("data/chart_%d.png", ix)
			log.Printf("Image found: %q -> %s - alt: %s\n", e.Text, link, alt)
			err := downloadFile(conf.Current.ChatServerURI+link, fileNameDst)
			item := InfoChart{
				Error:   err,
				Alt:     alt,
				Text:    e.Text,
				FileDst: fileNameDst,
			}
			found = true
			chItem <- &item
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnScraped(func(e *colly.Response) {
		log.Println("Terminate request scrap")
		if !found {
			log.Println("Chart image not recognized")
			item := InfoChart{
				Error: fmt.Errorf("Chart not recognized on %s", URL),
			}
			chItem <- &item
		}
	})
	c.OnError(func(e *colly.Response, err error) {
		log.Println("Error on scrap", err)
		if !found {
			log.Println("Chart image error")
			item := InfoChart{
				Error: err,
			}
			chItem <- &item
		}
	})

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
