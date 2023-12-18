package web

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"text/template"
	"time"
)

type PageCtx struct {
	Buildnr     string
	SvgData     string
	SvgDataBa64 string
}

const (
	buildnr = "00.00.10.00"
)

func handleGet(w http.ResponseWriter, req *http.Request) error {
	var err error
	u, _ := url.Parse(req.RequestURI)
	log.Println("GET requested ", u)
	dat, err := os.ReadFile("static/data/chart02.svg")
	if err != nil {
		return err
	}

	svgstr := string(dat)
	svgstr = strings.ReplaceAll(svgstr, "'2", "2")
	svgstr = strings.ReplaceAll(svgstr, "'3", "3")
	svgstr = strings.ReplaceAll(svgstr, "\n", "")
	svgstr = strings.ReplaceAll(svgstr, "\r", "")
	svgstrtoba := base64.StdEncoding.EncodeToString([]byte(svgstr))
	svgstrtoba = fmt.Sprintf("data:image/svg+xml;base64, %s", svgstrtoba)
	pagectx := PageCtx{
		Buildnr:     buildnr,
		SvgData:     svgstr,
		SvgDataBa64: svgstrtoba,
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
		if err := handleGet(w, req); err != nil {
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
