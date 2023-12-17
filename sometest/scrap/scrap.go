package scrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	sel_1month  = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(2)`
	sel_6month  = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label:nth-child(3)`
	sel_svgnode = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.chart-container > div > div`
	sel_spinner = `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.chart-container > div.loading-overlay.center-spinner`
)

func Scrap() error {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var example string
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Navigate to chart")
			return nil
		}),
		chromedp.Navigate(`https://www.easybank.at/markets/etf/tts-23270949/XTR-FTSE-DEV-EUR-R-EST-1C`),
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
		return err
	}
	log.Println("run scraping terminated ok")
	//log.Printf("SVG after get:\n%s", example)
	outfname := "static/data/chart02.svg"
	if err = os.WriteFile(outfname, []byte(example), 0644); err != nil {
		return err
	}

	log.Println("svg file written ", outfname)
	return nil
}
