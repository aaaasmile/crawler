package scrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/scraper/util"
	"github.com/chromedp/chromedp"
)

const (
	// following selectors are all inside the chart. They can change, so you have to inspect it inside the browser and copy the selector
	sel_6month = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(3)`
	sel_1year  = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(4)`
	//sel_svgnode should be exact the <svg> node, or you need a post process
	sel_svgnode     = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div > div.chart-container`
	sel_spinner     = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div > div.chart-container > div.loading-overlay.center-spinner`
	cookie_selector = `#onetrust-accept-btn-handler`
)

const (
	service_svgtopng = "http://localhost:5903/svg/"
)

type ScrapItem struct {
	_id       int
	_png_path string
	_png_name string
	_err      error
}

type Scrap struct {
	liteDB          *db.LiteDB
	_svgs           []*ScrapItem
	_takeScreenshot bool
	_cookies        bool
}

func NewScrap(takescreen, cookies bool) *Scrap {
	return &Scrap{
		_takeScreenshot: takescreen,
		_cookies:        cookies,
	}
}

func (sc *Scrap) Scrap(dbPath string, limit int) error {
	if err := util.CleanPNGData(); err != nil {
		return err
	}

	sc.liteDB = &db.LiteDB{
		SqliteDBPath: dbPath,
	}
	if err := sc.liteDB.OpenSqliteDatabase(); err != nil {
		return err
	}
	var err error
	upperlimit := 100
	if limit != -1 {
		fmt.Println("Limit scrap to file num: ", limit)
		upperlimit = limit
	}
	sc._svgs = []*ScrapItem{}
	stockList, err := sc.liteDB.SelectEnabledStockInfos(upperlimit)
	if err != nil {
		return err
	}
	for _, stockItem := range stockList {
		charturl := stockItem.ChartURL // `https://www.easybank.at/markets/etf/tts-23270949/XTR-FTSE-DEV-EUR-R-EST-1C`
		err = sc.scrapItem(charturl, int(stockItem.ID))
		if err != nil {
			log.Println("error on scraping ", charturl, err) // continue scraping ignoring wrong items
		}
	}
	return nil
}

func (sc *Scrap) scrapItem(charturl string, id int) error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(func(s string, i ...interface{}) {
		// 	fmt.Printf(s, i...)
		// }),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	pageCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	/// ackCtx is created from pageCtx.
	// when ackCtx exceeds the deadline, pageCtx is not affected.
	// This is needed beacuse the cookie popup is not always here. Without cookie the ctx is fault.
	ackCtx, cancel := context.WithTimeout(pageCtx, 10*time.Second)
	defer cancel()
	var screenbuf []byte

	// navigate to a page, wait for an element
	err := chromedp.Run(pageCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Navigate to chart - url: ", charturl)
			if err := chromedp.Navigate(charturl).Do(ctx); err != nil {
				log.Println("[scrapItem] error on navigation")
				return err
			}
			fmt.Println("*** navigate ok")
			if err := chromedp.WaitVisible(`body > footer`).Do(ctx); err != nil {
				log.Println("[scrapItem] error on visible footer")
				return err
			}
			fmt.Println("*** footer ok")
			if err := chromedp.WaitReady(sel_spinner, chromedp.NodeNotVisible).Do(ctx); err != nil {
				log.Println("[scrapItem] error on sel_spinner")
				return err
			}
			fmt.Println("*** initial spinner invisible")
			return nil
		}),
	)
	if err != nil {
		log.Println("[scrapItem] error on chromedp.Run Navigate", err)
		sc._svgs = append(sc._svgs, &ScrapItem{_err: err})
		return err
	}

	if sc._cookies {
		// note: here the ackCtx is used
		log.Println("[scrapItem] expect cookie, try to click away")
		err = chromedp.Run(ackCtx,
			chromedp.EmulateViewport(1920, 1080), // for cookies visibility
			chromedp.ActionFunc(func(ctx context.Context) error {
				// why chromedp.Click( cookie_selector) is not working?
				if err := chromedp.Query(cookie_selector, chromedp.NodeReady).Do(ctx); err == nil {
					log.Println("[scrapItem] cookie recognized")
					// since Click is not working, use MouseClickXY
					// to get the correct coordinate, please check the screenshot an try better coordinates
					// use the flag -screenshot and inspect pagechart001.png
					if err := chromedp.MouseClickXY(960, 650).Do(ctx); err != nil {
						log.Println("[scrapItem] cookie click error")
						return err
					}
					log.Println("[scrapItem] cookie accepted")
				}
				return nil
			}),
		)
		if err != nil {
			log.Println("[scrapItem] some error on cookie", err)
		}
	}
	log.Println("[scrapItem] continue")
	err = chromedp.Run(pageCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** select 1 year")
			if err := chromedp.WaitVisible(sel_1year, chromedp.NodeVisible).Do(ctx); err != nil {
				log.Println("[scrapItem] error sel_1year")
				return err
			}
			if err := chromedp.Click(sel_1year, chromedp.NodeVisible).Do(ctx); err != nil {
				log.Println("[scrapItem] error click sel_1year")
				return err
			}
			fmt.Println("*** Click sel_1year done")
			if err := chromedp.WaitReady(sel_svgnode, chromedp.NodeVisible).Do(ctx); err != nil {
				log.Println("[scrapItem] error on svg visible")
				return err
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** svg container is ready")
			//log.Println("sleep after svg container...")
			//time.Sleep(2 * time.Second) // this is important because data are loaded in background and is not clear wich selector is active after that
			return nil
		}),
		chromedp.WaitReady(sel_spinner, chromedp.NodeNotVisible), // this is also important to get all data
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** spinner invisible")
			time.Sleep(1 * time.Second) // wait some progress after spinner disappear beacuse chart is fading
			scitem := &ScrapItem{
				_id: id,
			}
			sc._svgs = append(sc._svgs, scitem)

			if err := sc.saveSVGtoPng(&ctx, scitem); err != nil {
				return err
			}

			if sc._takeScreenshot {
				log.Println("Take a small screenshot")
				act := chromedp.CaptureScreenshot(&screenbuf)
				act.Do(ctx)
				if err := os.WriteFile("static/data/fullpage001.png", screenbuf, 0644); err != nil {
					return err
				}
				log.Println("Screenshot saved ok")
			}
			return nil
		}),
	) // SVG is done

	if err != nil {
		log.Println("[scrapItem] error on chromedp.Run", err)
		sc._svgs = append(sc._svgs, &ScrapItem{_err: err})
		return err
	}
	log.Println("run scraping terminated ok")
	return nil
}

func (sc *Scrap) saveSVGtoPng(ctx *context.Context, scrapItem *ScrapItem) error {
	// probably this is better then saveToPngItem and simpler
	log.Println("[saveSVGtoPng] start")
	var buf []byte
	if err := chromedp.Screenshot(sel_svgnode, &buf).Do(*ctx); err != nil {
		log.Println("[saveSVGtoPng] screenshot error: ", err)
		scrapItem._err = err
		return err
	}
	destPath := util.GetChartPNGFullFileName(scrapItem._id)
	scrapItem._png_name = util.GetChartPNGFileNameOnly(scrapItem._id)
	scrapItem._png_path = destPath

	if err := os.WriteFile(destPath, buf, 0o644); err != nil {
		log.Println("[saveSVGtoPng] save error: ", err)
		return err
	}
	log.Println("[saveSVGtoPng] saved to ", destPath)
	return nil
}

func (sc *Scrap) ReportProcessed() string {
	png_count := 0
	err_on_scrap := 0
	for _, item := range sc._svgs {
		if item._err != nil {
			err_on_scrap += 1
			continue
		}
		if item._png_name != "" {
			png_count += 1
		}
		//fmt.Println("*** item ", *item)
	}
	return fmt.Sprintf("processed %d, png %d, png errors %d",
		len(sc._svgs), png_count, err_on_scrap)
}
