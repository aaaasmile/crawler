package crawler

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aaaasmile/crawler/conf"
	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/idl"
	"github.com/aaaasmile/crawler/mail"

	"github.com/gocolly/colly/v2"
)

type CrawlerOfChart struct {
	liteDB      *db.LiteDB
	list        []*idl.ChartInfo
	serverURI   string
	Simulate    bool
	ResendEmail bool
}

type InfoChart struct {
	Error   error
	FileDst string
	Link    string
	Text    string
	Alt     string
	ID      int64
}

func (cc *CrawlerOfChart) Start(configfile string) error {
	conf.ReadConfig(configfile)
	log.Println("Configuration is read")

	cc.list = make([]*idl.ChartInfo, 0)
	cc.liteDB = &db.LiteDB{
		DebugSQL:     conf.Current.DebugSQL,
		SqliteDBPath: conf.Current.DBPath,
	}
	cc.serverURI = conf.Current.ServerURI

	if err := cc.liteDB.OpenSqliteDatabase(); err != nil {
		return err
	}

	if cc.ResendEmail {
		if err := cc.buildChartListFromLastDown(); err != nil {
			return err
		}
	} else if err := cc.buildTheChartList(); err != nil {
		return err
	}
	if err := cc.insertPriceList(); err != nil {
		return err
	}
	if err := cc.sendChartEmail(); err != nil {
		return err
	}

	return nil
}

func (cc *CrawlerOfChart) buildChartListFromLastDown() error {
	log.Println("Build list from last download")

	cc.list = make([]*idl.ChartInfo, 0)
	stockList, err := cc.liteDB.FetchStockInfo(100)
	if err != nil {
		return err
	}
	log.Println("Found stocks ", len(stockList))
	for _, v := range stockList {
		chartItem := idl.ChartInfo{}
		fileNameDst := fmt.Sprintf("data/chart_%d.png", v.ID)
		chartItem.Description = v.Description
		chartItem.MoreInfoURL = v.MoreInfoURL
		chartItem.ChartURL = v.ChartURL
		chartItem.DownloadFilename = fileNameDst

		cc.list = append(cc.list, &chartItem)
	}

	log.Println("Fetchart items", len(cc.list))

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
	mapStock := make(map[int64]*db.StockInfo)
	for _, v := range stockList {
		mapStock[v.ID] = v
		go pickPicture(v.ChartURL, v.ID, cc.serverURI, chRes)
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
			chartItem.HasError = res.Error != nil
			if res.Error != nil {
				chartItem.HasError = true
				chartItem.ErrorText = res.Error.Error()
			} else {
				chartItem.DownloadFilename = res.FileDst
				chartItem.CurrentPrice = res.Alt
				chartItem.PriceInfo, err = parseForPriceInfo(res.Alt)
				if err != nil {
					log.Println("Parse price info error", err)
					chartItem.HasError = true
					chartItem.ErrorText = err.Error()
				}
			}

			if v, ok := mapStock[res.ID]; ok {
				chartItem.Description = v.Description
				chartItem.MoreInfoURL = v.MoreInfoURL
				chartItem.ChartURL = v.ChartURL
				chartItem.ID = res.ID
			}

			cc.list = append(cc.list, &chartItem)
			log.Println("Append a new chart with ", res.FileDst, res.ID, counter)
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

func (cc *CrawlerOfChart) insertPriceList() error {
	log.Println("Insert price list")
	var id int64
	var err error
	var tx *sql.Tx
	tx, err = cc.liteDB.GetNewTransaction()
	if err != nil {
		return err
	}
	count := 0
	for _, v := range cc.list {
		if v.PriceInfo == nil {
			continue
		}
		id, err = cc.liteDB.InsertPrice(tx, v.ID, v.PriceInfo.Price, v.PriceInfo.TimestampInt)
		if err != nil {
			return err
		}
		log.Printf("Inserted price id %d for stock id %d", id, v.ID)
		count++
	}
	if count > 0 {
		log.Println("Commit insert transactions ", count)
		cc.liteDB.CommitTransaction(tx)
	}
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
	if err := mm.SendEmailViaOAUTH2(templFileName, cc.list); err != nil {
		return err
	}

	return nil
}

func pickPicture(URL string, ix int64, serverURI string, chItem chan *InfoChart) {
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
				ID:      ix,
			}
			err := downloadFile(serverURI+item.Link, item.FileDst)
			if err != nil {
				log.Println("Downloading image error", err)
				item.Error = err
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
				ID:    ix,
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
			ID:    ix,
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

func parseForPriceInfo(alt string) (*db.Price, error) {
	// alt is something like: IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)
	arr := strings.Split(alt, "-")
	if len(arr) < 2 {
		return nil, fmt.Errorf("Expect at least one dash")
	}
	item := arr[len(arr)-1]
	arr = strings.Split(item, ":")
	if len(arr) != 3 {
		return nil, fmt.Errorf("Expect 2 ':'")
	}
	item = strings.Join(arr[1:], ":")
	item = strings.Trim(item, " ") //16,34 (15.01. / 17:36)

	arr = strings.Split(item, " ")
	if len(arr) < 1 {
		return nil, fmt.Errorf("Expect date and time with space separation")
	}
	pricestr := arr[0] //16,34
	pricestr = strings.Replace(pricestr, ",", ".", 1)
	price, err := strconv.ParseFloat(pricestr, 64)
	if err != nil {
		return nil, err
	}

	datestr := strings.Join(arr[1:], " ")
	datestr = strings.Trim(datestr, "(")
	datestr = strings.Trim(datestr, ")") //15.01. / 17:36
	arr = strings.Split(datestr, "/")
	if len(arr) != 2 {
		return nil, fmt.Errorf("Expected one / separator")
	}
	datestr = arr[0] //15.01.
	pparr := strings.Split(datestr, ".")
	if len(pparr) != 3 {
		return nil, fmt.Errorf("Expected 3 date field separated with dot")
	}
	dd := strings.Trim(pparr[0], " ")
	mm := strings.Trim(pparr[1], " ")
	yy := strings.Trim(pparr[2], " ")
	if yy == "" {
		yy = fmt.Sprintf("%d", time.Now().Year())
	}

	timestr := arr[1] // 17:36
	timestr = strings.Trim(timestr, " ")
	pptimearr := strings.Split(timestr, ":")
	if len(pptimearr) != 2 {
		return nil, fmt.Errorf("Expected hour and minute separated with ':'")
	}
	hh := pptimearr[0]
	min := pptimearr[1]

	timeforparse := fmt.Sprintf("%s-%s-%sT%s:%s:00+00:00", yy, mm, dd, hh, min)
	//fmt.Println("** Time for parse is ", timeforparse)
	tt, err := time.Parse(time.RFC3339, timeforparse)
	if err != nil {
		return nil, err
	}
	priceItem := db.Price{
		Price:        price,
		TimestampInt: tt.Local().Unix(),
		Timestamp:    tt,
	}

	return &priceItem, nil
}
