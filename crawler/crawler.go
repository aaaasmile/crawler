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
	Error      error
	FileDst    string
	Link       string
	Text       string
	Alt        string
	ID         int64
	PriceFinal string
	ClosedAt   string
}

func (cc *CrawlerOfChart) Start(configfile string) error {
	current, err := conf.ReadConfig(configfile)
	if err != nil {
		return err
	}
	log.Println("Configuration is read")

	cc.list = make([]*idl.ChartInfo, 0)
	cc.liteDB = &db.LiteDB{
		DebugSQL:     current.DebugSQL,
		SqliteDBPath: current.DBPath,
	}
	cc.serverURI = current.ServerURI

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
	// TEST the fetch first and then enable the e-mail
	// if err := cc.sendChartEmail(); err != nil {
	// 	return err
	// }

	return nil
}

func (cc *CrawlerOfChart) buildChartListFromLastDown() error {
	log.Println("Build list from last download")

	cc.list = make([]*idl.ChartInfo, 0)
	stockList, err := cc.liteDB.SelectEnabledStockInfos(100)
	if err != nil {
		return err
	}
	log.Println("Found stocks ", len(stockList))
	for _, v := range stockList {
		chartItem := idl.ChartInfo{}
		fileNameDst := fmt.Sprintf("data/chart_%d.png", v.ID)
		chartItem.Description = v.Description
		chartItem.MoreInfoURL = v.MoreInfoURL
		chartItem.SimpleDescr = v.SimpleDescr
		chartItem.ChartURL = v.ChartURL
		chartItem.DownloadFilename = fileNameDst

		cc.list = append(cc.list, &chartItem)
	}

	log.Println("Chart items are", len(cc.list))

	return nil
}

func (cc *CrawlerOfChart) buildTheChartList() error {
	log.Println("Build the chart list")
	start := time.Now()
	cc.list = make([]*idl.ChartInfo, 0)
	stockList, err := cc.liteDB.SelectEnabledStockInfos(100)
	if err != nil {
		return err
	}
	log.Println("Found stocks in DB ", len(stockList))

	chRes := make(chan *InfoChart)
	mapStock := make(map[int64]*db.StockInfo)
	for _, v := range stockList {
		if mapStock[v.ID] != nil {
			return fmt.Errorf("duplicate key %d", v.ID)
		}
		mapStock[v.ID] = v
	}
	for _, v := range stockList {
		go pickChartDetail(v.ChartURL, v.ID, cc.serverURI, chRes)
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
			chartItem := idl.ChartInfo{ColorWL: "green"}
			chartItem.HasError = res.Error != nil
			if res.Error != nil {
				log.Println("chartItem has an error:", res.Error)
				chartItem.HasError = true
				chartItem.ErrorText = res.Error.Error()
			} else {
				chartItem.CurrentPrice = res.PriceFinal
				chartItem.PriceInfo, err = parseForPriceInfo(res.PriceFinal, res.ClosedAt)
				if err != nil {
					log.Println("Parse price info error", err)
					chartItem.HasError = true
					chartItem.ErrorText = err.Error()
				}
			}

			if v, ok := mapStock[res.ID]; ok {
				chartItem.Description = v.Description
				chartItem.MoreInfoURL = v.MoreInfoURL
				chartItem.SimpleDescr = v.SimpleDescr
				chartItem.ChartURL = v.ChartURL
				chartItem.ID = res.ID
				if chartItem.PriceInfo != nil {
					priceCurr := chartItem.PriceInfo.Price
					totval := priceCurr * v.Quantity
					winorloss := totval - v.Cost
					chartItem.WinOrLoss = fmt.Sprintf("%.2f", winorloss)
					if winorloss < 0 {
						chartItem.ColorWL = "red"
					}
					if v.Cost != 0 {
						wlper := winorloss / v.Cost * 100.0
						chartItem.WinOrLossPerc = fmt.Sprintf("%.2f", wlper)
					}
					chartItem.TotCurrValue = fmt.Sprintf("%.2f", totval)
					chartItem.TotCost = fmt.Sprintf("%.2f", v.Cost)
					chartItem.Quantity = fmt.Sprintf("%.2f", v.Quantity)

				}
			} else {
				log.Println("WARN: ID not recognized ", res.ID, res)
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
	log.Printf("buildTheChartList: items %d total call duration: %v\n", len(cc.list), elapsed)

	return nil
}

func (cc *CrawlerOfChart) insertPriceList() error {
	log.Println("Insert price list")
	var id int64
	var err error
	var tx *sql.Tx
	var pps []*db.Price
	tx, err = cc.liteDB.GetNewTransaction()
	if err != nil {
		return err
	}
	count := 0
	for _, v := range cc.list {
		if v.PriceInfo == nil {
			log.Println("WARN: no price info avalible for ", v)
			continue
		}
		pps, err = cc.liteDB.SelectPrice(v.ID, v.PriceInfo.Price, v.PriceInfo.TimestampInt)
		if err != nil {
			return err
		}
		if len(pps) == 0 {
			id, err = cc.liteDB.InsertPrice(tx, v.ID, v.PriceInfo.Price, v.PriceInfo.TimestampInt)
			if err != nil {
				return err
			}
			log.Printf("Inserted price id %d for stock id %d", id, v.ID)
			count++
		} else {
			log.Println("Price already inserted", v.ID, v.PriceInfo.Price)
		}
		pps, err = cc.liteDB.SelectPreviousPriceInStock(v.ID, v.PriceInfo.TimestampInt)
		if err != nil {
			return err
		}
		if len(pps) == 1 {
			prev := pps[0]
			log.Println("Found previous price ", prev.Price)
			v.PreviousPrice = prev.Price
			if prev.Price != 0 {
				v.DiffPreviousPrice = (v.PriceInfo.Price - prev.Price) / prev.Price * 100.0
			}
		} else if len(pps) > 1 {
			return fmt.Errorf("some strange previous %d %v %d", len(pps), pps, v.ID)
		}
	}
	if count > 0 {
		log.Println("Commit insert transactions ", count)
		cc.liteDB.CommitTransaction(tx)
	}
	return nil
}

func (cc *CrawlerOfChart) fillWithSomeTesdata() {
	// example without the crawler
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
	cc.list = append(cc.list, &idl.ChartInfo{Description: "chart 1", DownloadFilename: "data/chart_01.png"})
}

func (cc *CrawlerOfChart) sendChartEmail() error {

	log.Println("Send email with num of items", len(cc.list))

	mm := mail.NewMailSender(cc.liteDB, cc.Simulate)

	if err := mm.FetchSecretFromDb(); err != nil {
		return err
	}
	return cc.sendMailViaRelay(mm)

}
func (cc *CrawlerOfChart) sendMailViaRelay(mm *mail.MailSender) error {
	log.Println("Using relay to send the mail")

	templFileName := "templates/chart-mail.html"
	if err := mm.SendEmailViaRelay(templFileName, cc.list); err != nil {
		return err
	}

	return nil
}

func pickChartDetail(URL string, id int64, serverURI string, chItem chan *InfoChart) {
	log.Println("Fetching chart for ", id, URL)
	c := colly.NewCollector()
	sent := false
	item := InfoChart{
		ID: id,
	}
	// https://github.com/PuerkitoBio/goquery
	// https://github.com/gocolly/colly/blob/master/_examples
	c.OnHTML("section.card", func(e *colly.HTMLElement) {
		// section card has an header and a table as children
		// identofy both and address the text directly using ChildText selector
		hh := e.ChildText("header > h2")
		if strings.HasPrefix(hh, "Basisinformationen") {
			//fmt.Println("*** H ", hh)
			psfinlbl := e.ChildText("table > tbody > tr:nth-child(2) > td:nth-child(1)")
			psfinval := e.ChildText("table > tbody > tr:nth-child(2) > td:nth-child(2)")
			//fmt.Println("***  ", psfinlbl, psfinval)
			item.PriceFinal = psfinval
			item.ClosedAt = psfinlbl
			//sent = true
			//chItem <- &item
		} else if strings.HasPrefix(hh, "Aktuelle Entwicklung") {
			fmt.Println("*** H ", hh)
			//svg := e.DOM.ChildrenFiltered("svg")
			// svghtml, err := e.DOM.ChildrenMatcher("div > div").Html()
			// if err != nil {
			// 	log.Println("SVG html error", err)
			// 	item.Error = err
			// 	sent = true
			// 	chItem <- &item
			// }
			e.ForEach("div.card-body > div", func(_ int, el *colly.HTMLElement) {
				//svghtml, _ := el.DOM.ChildrenFiltered(".chart-container").Html()
				//svghtml := el.DOM.ChildrenFiltered(".chart-container")
				svghtml, _ := el.DOM.Html()
				fmt.Println("*** SVG ", svghtml)
			})
			// svg := e.DOM.ChildrenFiltered("div.card-body ")
			// svghtml, _ := svg.Html()
			// fmt.Println("*** SVG ", svghtml)

			//fileNameDst := fmt.Sprintf("data/chart_%d.svg", id)
		}
	})
	// On every a element which has href attribute call callback
	// c.OnHTML("img[src]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("src")
	// 	alt := e.Attr("alt")
	// 	if strings.HasPrefix(link, "getChart") {
	// 		fileNameDst := fmt.Sprintf("data/chart_%d.png", id)
	// 		log.Printf("Image found: %q -> %s - alt: %s\n", e.Text, link, alt)
	// 		item := InfoChart{
	// 			Alt:     alt, //IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)
	// 			Link:    link,
	// 			Text:    e.Text,
	// 			FileDst: fileNameDst,
	// 			ID:      id,
	// 		}
	// 		err := downloadFile(serverURI+item.Link, item.FileDst)
	// 		if err != nil {
	// 			log.Println("Downloading image error", err)
	// 			item.Error = err
	// 		}
	// 		found = true
	// 		chItem <- &item
	// 	}
	// 	//fmt.Println("*** link image", link, alt)
	// 	//something like: *** link image getChart.asp?action=getChart&chartID=71C233968F97F40CD296DA8A36E792DF6A50394A IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)
	// })

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})
	c.OnError(func(e *colly.Response, err error) {
		log.Println("Error on scrap", err)
		if !sent {
			log.Println("Chart image error")
			item.Error = err
			chItem <- &item
			sent = true
		}
	})
	c.Visit(URL)

	log.Println("Terminate request")
	if !sent {
		log.Println("Chart not found")
		item.Error = fmt.Errorf("chart not recognized (service html layout changed?) on %s", URL)
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
		return errors.New("received non 200 response code")
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

func parseForPriceInfo(pricestr string, closed string) (*db.Price, error) {
	// price is something like: 16,34
	// closed is like: Schluss 15.01.23
	arr := strings.Split(closed, " ")
	if len(arr) < 2 {
		return nil, fmt.Errorf("expect at least one space")
	}

	pricestr = strings.Replace(pricestr, ",", ".", 1)
	price, err := strconv.ParseFloat(pricestr, 64)
	if err != nil {
		return nil, err
	}

	datestr := arr[1]
	pparr := strings.Split(datestr, ".")
	if len(pparr) != 3 {
		return nil, fmt.Errorf("expected 3 date field separated with dot")
	}
	dd := strings.Trim(pparr[0], " ")
	mm := strings.Trim(pparr[1], " ")
	yy := strings.Trim(pparr[2], " ")
	if yy == "" {
		yy = fmt.Sprintf("%d", time.Now().Year())
	} else if len(yy) == 2 {
		yy = fmt.Sprintf("20%s", yy)
	}

	hh := "17" // use a fixed closed time because it is not provided anymore
	min := "36"
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
