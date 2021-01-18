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
	Link    string
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
	log.Println("Build the chart list")
	start := time.Now()
	cc.list = make([]*idl.ChartInfo, 0)
	stockList, err := cc.liteDB.FetchStockInfo(100)
	if err != nil {
		return err
	}
	log.Println("Found stocks ", len(stockList))

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

loop:
	for {
		select {
		case res = <-chRes:
			chartItem := idl.ChartInfo{}
			err := downloadFile(conf.Current.ChatServerURI+res.Link, res.FileDst)
			if err != nil {
				log.Println("Downloading image error")
				chartItem.HasError = true
				chartItem.ErrorText = err.Error()
			} else {
				chartItem.HasError = res.Error != nil
				if res.Error != nil {
					chartItem.ErrorText = res.Error.Error()
				} else {
					chartItem.DownloadFilename = res.FileDst
					chartItem.CurrentPrice = res.Alt
				}
			}

			if v, ok := mapStock[res.Ix]; ok {
				chartItem.Description = v.Description
				chartItem.MoreInfoURL = v.MoreInfoURL
				chartItem.ChartURL = v.ChartURL
			}

			cc.list = append(cc.list, &chartItem)
			log.Println("Append a new chart with ", res.FileDst, res.Ix, counter)
			counter--
			if counter <= 0 {
				log.Println("All images are donwloaded")
				break loop
			}
		case <-chTimeout:
			log.Println("Timeout on shutdown, something was blocked")
			cc.list = append(cc.list, &idl.ChartInfo{HasError: true, ErrorText: "Timeout on fetching chart"})
			break loop
		}
	}
	t := time.Now()
	elapsed := t.Sub(start)
	log.Printf("Fetchart items %d total call duration: %v\n", len(cc.list), elapsed)

	return nil
}

func (cc *CrawlerOfChart) fillWithSomeDummy() {
	// example without the crawler
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
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

func pickPicture(URL string, ix int, chItem chan *InfoChart) {
	log.Println("Fetching chart for ", ix, URL)
	c := colly.NewCollector()
	found := false
	// On every a element which has href attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		alt := e.Attr("alt")
		if strings.HasPrefix(link, "getChart") {
			fileNameDst := fmt.Sprintf("data/chart_%d.png", ix)
			log.Printf("Image found: %q -> %s - alt: %s\n", e.Text, link, alt)
			item := InfoChart{
				Alt:     alt, //IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)
				Link:    link,
				Text:    e.Text,
				FileDst: fileNameDst,
				Ix:      ix,
			}
			found = true
			chItem <- &item
		}
		//fmt.Println("*** link image", link, alt)
		//something like: *** link image getChart.asp?action=getChart&chartID=71C233968F97F40CD296DA8A36E792DF6A50394A IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
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
	c.Visit(URL)

	log.Println("Terminate request")
	if !found {
		log.Println("Chart image not recognized")
		item := InfoChart{
			Error: fmt.Errorf("Chart not recognized on %s", URL),
			Ix:    ix,
		}
		chItem <- &item
	}

}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	log.Println("Downloading the URL to the filename: ", fileName)
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
	time.Sleep(200)
	return nil
}
