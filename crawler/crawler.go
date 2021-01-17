package crawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	Ix      int
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
	start := time.Now()
	cc.list = make([]*idl.ChartInfo, 0)
	stockList, err := cc.liteDB.FetchStockInfo(2)
	if err != nil {
		return err
	}

	chRes := make(chan *InfoChart)
	mapStock := make(map[int]*db.StockInfo)
	for _, v := range stockList {
		mapStock[v.ID] = v
		go pickPicture(v.ChartURL, v.ID, chRes)
	}

	chTimeout := make(chan struct{})
	timeout := 120 * time.Second
	time.AfterFunc(timeout, func() {
		chTimeout <- struct{}{}
	})

	var res *InfoChart
	counter := len(stockList)
	select {
	case res = <-chRes:
		cc.list = append(cc.list, &idl.ChartInfo{
			HasError:    res.Error != nil,
			ErrorText:   res.Error.Error(),
			Alt:         res.Alt,
			Description: mapStock[res.Ix].Description,
			MoreInfoURL: mapStock[res.Ix].MoreInfoURL,
			Fname:       mapStock[res.Ix].Name,
		})
		counter--
		if counter <= 0 {
			break
		}
	case <-chTimeout:
		log.Println("Timeout on shutdown, something was blockd")
		cc.list = append(cc.list, &idl.ChartInfo{HasError: true, ErrorText: "Timeout on fetching chart"})
		break
	}
	t := time.Now()
	elapsed := t.Sub(start)
	log.Printf("Fetchart total call duration: %v\n", elapsed)

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
				Ix:      ix,
			}
			found = true
			chItem <- &item
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnError(func(e *colly.Response, err error) {
		log.Println("Error on scrap", err)
		if !found {
			log.Println("Chart image error")
			item := InfoChart{
				Error: err,
				Ix:    ix,
			}
			chItem <- &item
		}
	})

	log.Println("Terminate request scrap")
	if !found {
		log.Println("Chart image not recognized")
		item := InfoChart{
			Error: fmt.Errorf("Chart not recognized on %s", URL),
			Ix:    ix,
		}
		chItem <- &item
	}

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
