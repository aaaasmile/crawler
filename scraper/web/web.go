package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aaaasmile/crawler/scraper/util"
)

type PageCtx struct {
	Buildnr string
	SvgData string
}

const (
	buildnr = "00.20231229.11.00"
)

func handleGetID(w http.ResponseWriter, req *http.Request) error {
	var err error
	u, _ := url.Parse(req.RequestURI)
	log.Println("GET requested ", u)
	match, err := regexp.MatchString(".*svg/([0-9]+)$", u.String())
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("svg id not recognized")
	}
	aa := strings.Split(u.String(), "/")
	num_str := aa[len(aa)-1]
	num_id, err := strconv.Atoi(num_str)
	if err != nil {
		return err
	}
	//fmt.Println("**>", match)
	svg_fullfilename := util.GetChartSVGFileName(num_id)
	dat, err := os.ReadFile(svg_fullfilename)
	if err != nil {
		return err
	}

	svgstr := string(dat)
	svgstr = strings.ReplaceAll(svgstr, "'2", "2")
	svgstr = strings.ReplaceAll(svgstr, "'3", "3")
	svgstr = strings.ReplaceAll(svgstr, "\n", "")
	svgstr = strings.ReplaceAll(svgstr, "\r", "")
	pagectx := PageCtx{
		Buildnr: buildnr,
		SvgData: svgstr,
	}

	templName := "templates/index.html"

	tmplIndex := template.Must(template.New("AppIndex").ParseFiles(templName))

	err = tmplIndex.ExecuteTemplate(w, "base", pagectx)
	if err != nil {
		return err
	}

	return nil
}

func apiHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		if err := handleGetID(w, req); err != nil {
			log.Println("Error on process request: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func StartServer(cr <-chan struct{}) {
	//scrap.Scrap()
	serverurl := "127.0.0.1:5903"
	rootURLPattern := "/svg/"
	finalServURL := fmt.Sprintf("http://%s%s", strings.Replace(serverurl, "0.0.0.0", "localhost", 1), rootURLPattern)

	finalServURL = strings.Replace(finalServURL, "127.0.0.1", "localhost", 1)
	log.Println("Server started with URL ", serverurl)
	log.Println("Try this url: ", finalServURL)

	http.Handle(rootURLPattern+"static/", http.StripPrefix(rootURLPattern+"static", http.FileServer(http.Dir("static"))))
	http.HandleFunc(rootURLPattern, apiHandler)

	srv := &http.Server{
		Addr:         serverurl,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      nil,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println("Server is not listening anymore: ", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt) //We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	log.Println("Enter in server loop")
loop:
	for {
		select {
		case <-sig:
			log.Println("stop because interrupt")
			break loop
		case <-cr:
			log.Println("stop because service shutdown")
			break loop
		}
	}
	var wait time.Duration
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Bye, service")
}
