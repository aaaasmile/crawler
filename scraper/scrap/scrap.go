package scrap

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/scraper/util"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"golang.design/x/clipboard"
)

const (
	sel_1month  = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(2)`
	sel_6month  = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(3)`
	sel_svgnode = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.chart-container > div > div`
	sel_spinner = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.chart-container > div.loading-overlay.center-spinner`
)

const (
	service_svgtopng = "http://localhost:5903/svg/"
)

type ScrapItem struct {
	_id       int
	_svg_path string
	_svg_name string
	_png_path string
	_png_name string
	_err      error
}

type Scrap struct {
	liteDB *db.LiteDB
	_svgs  []*ScrapItem
}

func (sc *Scrap) Scrap(dbPath string) error {
	sc.liteDB = &db.LiteDB{
		SqliteDBPath: dbPath,
	}
	if err := sc.liteDB.OpenSqliteDatabase(); err != nil {
		return err
	}
	var err error
	sc._svgs = []*ScrapItem{}
	stockList, err := sc.liteDB.SelectEnabledStockInfos(100)
	if err != nil {
		return err
	}
	for _, stockItem := range stockList {
		charturl := stockItem.ChartURL // `https://www.easybank.at/markets/etf/tts-23270949/XTR-FTSE-DEV-EUR-R-EST-1C`
		err = sc.scrapItem(charturl, int(stockItem.ID))
		if err != nil {
			log.Println("error on scraping ", charturl) // continue scraping ignoring wrong items
		}
	}
	return nil
}

func (sc *Scrap) SaveToPng() error {
	if len(sc._svgs) == 0 {
		return fmt.Errorf("no svg provided")
	}
	aa := []*ScrapItem{}
	for _, item := range sc._svgs {
		if item._err == nil {
			moditem, err := sc.saveToPngItem(item)
			if err != nil {
				log.Println("error on save to png ", err) // continue save ignoring wrong items
				aa = append(aa, item)
			} else {
				aa = append(aa, moditem)
			}
		} else {
			aa = append(aa, item)
		}
	}
	sc._svgs = aa
	return nil
}

func (sc *Scrap) PrepareTestSVG() {
	// some test files
	sc._svgs = []*ScrapItem{{_id: 1, _svg_name: "chart01.svg", _svg_path: "static/data/chart01.svg"}}
	fmt.Println("using some test data ", sc._svgs[0])
}

func (sc *Scrap) scrapItem(charturl string, id int) error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var example string
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Navigate to chart")
			return nil
		}),
		chromedp.Navigate(charturl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Wait visible")
			return nil
		}),
		// wait for footer element is visible (ie, page is loaded)
		chromedp.WaitVisible(`body > footer`),
		// dafault chart is intraday, not interesting for me
		chromedp.WaitReady(sel_spinner, chromedp.NodeNotVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** initial spinner invisible")
			return nil
		}),
		// click on chart  Monat,  use Browser Copy Selector for this link and make sure that the link is not active
		chromedp.Click(sel_6month, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Click month done")
			return nil
		}),
		chromedp.WaitReady(sel_svgnode, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** svg container is ready")
			log.Println("sleep some seconds...")
			time.Sleep(2 * time.Second) // this is important because data are loaded in background and is not clear wich selector is active after that
			return nil
		}),
		chromedp.WaitReady(sel_spinner, chromedp.NodeNotVisible), // this is also important to get all data
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** spinner invisible")
			return nil
		}),
		chromedp.InnerHTML(sel_svgnode, // finally get the chart
			&example,
			chromedp.NodeVisible),
	)
	if err != nil {
		sc._svgs = append(sc._svgs, &ScrapItem{_err: err})
		return err
	}
	log.Println("run scraping terminated ok")
	//log.Printf("SVG after get:\n%s", example)
	outfname := util.GetChartSVGFullFileName(id)
	if err = os.WriteFile(outfname, []byte(example), 0644); err != nil {
		sc._svgs = append(sc._svgs, &ScrapItem{_err: err})
		return err
	}

	log.Println("svg file written ", outfname)
	scitem := &ScrapItem{
		_id:       id,
		_svg_path: outfname,
		_svg_name: util.GetChartSVGFileNameOnly(id),
	}
	sc._svgs = append(sc._svgs, scitem)
	return nil
}

func EncodeFont() error {
	dat, err := os.ReadFile("static/css/fonts/DINPro-Regular.woff")
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(dat)
	clipboard.Write(clipboard.FmtText, []byte(encoded))
	fmt.Println("base64: ", encoded)
	return nil
}

func (sc *Scrap) saveToPngItem(scrapItem *ScrapItem) (*ScrapItem, error) {
	// reference: https://github.com/chromedp/examples/blob/master/download_file/main.go
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	done := make(chan string, 1)
	cr := make(chan struct{})

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			completed := "(unknown)"
			if ev.TotalBytes != 0 {
				completed = fmt.Sprintf("%0.2f%%", ev.ReceivedBytes/ev.TotalBytes*100.0)
			}
			log.Printf("state: %s, completed: %s\n", ev.State.String(), completed)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				close(done)
			} else if ev.State == browser.DownloadProgressStateCanceled {
				cr <- struct{}{}
				close(cr)
			}
		}
	})

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var err error
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	log.Println("using directory ", wd) // the best way
	// navigate to a page, wait for an element, click
	urlstr := fmt.Sprintf("%s%d", service_svgtopng, scrapItem._id)
	log.Println("using the service ", urlstr)
	if err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Navigate to svg converter")
			return nil
		}),
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Wait visible")
			return nil
		}),
		chromedp.WaitVisible(`#thesvg > svg`),
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.Click(`body > button`, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Click done")
			return nil
		}),
	); err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		return nil, err
	}
	log.Println("save to png started...", scrapItem._id)
	chTimeout := make(chan struct{})
	timeout := 30 * time.Second
	time.AfterFunc(timeout, func() {
		chTimeout <- struct{}{}
		close(chTimeout)
	})
	// This will block until the chromedp listener closes the channel
loop:
	for {
		select {
		case guid := <-done:
			srcpath := filepath.Join(wd, guid)
			destPath := util.GetChartPNGFullFileName(scrapItem._id)
			log.Printf("wrote %s", srcpath)
			if err := os.Rename(srcpath, destPath); err != nil {
				return nil, err
			}
			log.Println("source file moved to ", destPath)
			scrapItem._png_name = util.GetChartPNGFileNameOnly(scrapItem._id)
			scrapItem._png_path = destPath
			break loop
		case <-cr:
			log.Println("stop because service shutdown")
			break loop
		case <-chTimeout:
			log.Println("Timeout on save to png, something was blocked")
			break loop
		}
	}
	//fmt.Println("*** png scrapItem ", scrapItem)
	return scrapItem, nil
}

func (sc *Scrap) ReportProcessed() string {
	svg_count := 0
	png_count := 0
	err_on_svg := 0
	for _, item := range sc._svgs {
		if item._err != nil {
			err_on_svg += 1
			continue
		}
		if item._svg_name != "" {
			svg_count += 1
		}
		if item._png_name != "" {
			png_count += 1
		}
		//fmt.Println("*** item ", *item)
	}
	err_on_png := svg_count - png_count
	return fmt.Sprintf("processed %d, svg %d, png %d, png errors %d, svg errors %d",
		len(sc._svgs), svg_count, png_count, err_on_png, err_on_svg)
}
