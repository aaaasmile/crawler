package scrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func Scrap() {
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
	sel6month := `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.btn-group.btn-group-toggle.btn-group-left.chart-level-buttons > label.btn.btn-link.active`
	sel_svgnode := `body > div.page-content > main > article > div:nth-child(3) > section:nth-child(1) > div.card-body > div.chart-container > div > div`
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
		// click on chart 6 Monate Use Browser Copy Selector for this link
		chromedp.Click(sel6month, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** Click done")
			//log.Println("sleep some seconds...")
			//time.Sleep(3 * time.Second)
			return nil
		}),
		chromedp.WaitReady(sel_svgnode, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("*** svg container is ready")
			log.Println("sleep some seconds...")
			time.Sleep(10 * time.Second)
			return nil
		}),
		//chromedp.Value(`#highcharts-ymhu649-482 > svg`, &example),
		chromedp.InnerHTML(sel_svgnode,
			&example,
			chromedp.NodeVisible),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("run scraping terminated ok")
	//log.Printf("SVG after get:\n%s", example)
	outfname := "static/data/chart02.svg"
	os.WriteFile(outfname, []byte(example), 0)
	log.Println("svg file written ", outfname)
}
