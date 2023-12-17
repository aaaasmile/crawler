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

// Scrap2 è un test che non funziona, nel tentativo di scaricare tutti i dati del grafico
// func Scrap2() error {
// 	log.Println("Scrap 2")
// 	// create context
// 	ctx, cancel := chromedp.NewContext(
// 		context.Background(),
// 		chromedp.WithLogf(log.Printf),
// 	)
// 	defer cancel()

// 	// create a timeout as a safety net to prevent any infinite wait loops
// 	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
// 	defer cancel()

// 	// set up a channel, so we can block later while we monitor the download
// 	// progress
// 	done := make(chan bool)

// 	// set the download url as the chromedp GitHub user avatar
// 	urlstr := `https://www.easybank.at/markets/etf/tts-23270949/XTR-FTSE-DEV-EUR-R-EST-1C`

// 	// this will be used to capture the request id for matching network events
// 	var requestID network.RequestID

// 	// set up a listener to watch the network events and close the channel when
// 	// complete the request id matching is important both to filter out
// 	// unwanted network events and to reference the downloaded file later
// 	chromedp.ListenTarget(ctx, func(v interface{}) {
// 		switch ev := v.(type) {
// 		case *network.EventRequestWillBeSent:
// 			log.Printf("EventRequestWillBeSent: %v: %v", ev.RequestID, ev.Request.URL)
// 			if ev.Request.URL == urlstr {
// 				requestID = ev.RequestID
// 			}
// 		case *network.EventLoadingFinished:
// 			log.Printf("EventLoadingFinished: %v", ev.RequestID)
// 			if ev.RequestID == requestID {
// 				close(done)
// 			}
// 		}
// 	})

// 	// all we need to do here is navigate to the download url
// 	if err := chromedp.Run(ctx,
// 		chromedp.Navigate(urlstr),
// 		chromedp.ActionFunc(func(ctx context.Context) error {
// 			fmt.Println("*** navigated: ", urlstr)
// 			return nil
// 		}),
// 		chromedp.ActionFunc(func(ctx context.Context) error {
// 			fmt.Println("*** Wait visible")
// 			return nil
// 		}),
// 		// wait for footer element is visible (ie, page is loaded)
// 		chromedp.WaitVisible(`body > footer`),
// 		// click on chart 6 Monate Use Browser Copy Selector for this link
// 		chromedp.Click(sel_6month, chromedp.NodeVisible),
// 		chromedp.ActionFunc(func(ctx context.Context) error {
// 			fmt.Println("*** Click done")
// 			//log.Println("sleep some seconds...")
// 			//time.Sleep(3 * time.Second)
// 			return nil
// 		}),
// 		chromedp.WaitReady(sel_svgnode, chromedp.NodeVisible),
// 	); err != nil {
// 		return err
// 	}

// 	// This will block until the chromedp listener closes the channel
// 	<-done
// 	// get the downloaded bytes for the request id
// 	var buf []byte
// 	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
// 		var err error
// 		buf, err = network.GetResponseBody(requestID).Do(ctx)
// 		return err
// 	})); err != nil {
// 		return err
// 	}

// 	// write the file to disk - since we hold the bytes we dictate the name and
// 	// location
// 	outfile := "download"
// 	if err := os.WriteFile(outfile, buf, 0644); err != nil {
// 		return err
// 	}
// 	log.Print(outfile)
// 	return nil
// }
