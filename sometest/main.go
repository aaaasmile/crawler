package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

type PageCtx struct {
	Buildnr string
}

func handleGet(w http.ResponseWriter, req *http.Request) error {
	var err error
	u, _ := url.Parse(req.RequestURI)
	log.Println("GET requested ", u)
	pagectx := PageCtx{}

	templName := "templates/index.html"

	tmplIndex := template.Must(template.New("AppIndex").ParseFiles(templName))

	err = tmplIndex.ExecuteTemplate(w, "base", pagectx)
	if err != nil {
		return err
	}

	return nil
}

func APiHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		if err := handleGet(w, req); err != nil {
			log.Println("Error on process request: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func main() {
	//scrap.Scrap()
	serverurl := "127.0.0.1:5903"
	rootURLPattern := "/svg/"
	finalServURL := fmt.Sprintf("http://%s%s", strings.Replace(serverurl, "0.0.0.0", "localhost", 1), rootURLPattern)

	finalServURL = strings.Replace(finalServURL, "127.0.0.1", "localhost", 1)
	log.Println("Server started with URL ", serverurl)
	log.Println("Try this url: ", finalServURL)

	http.Handle(rootURLPattern+"static/", http.StripPrefix(rootURLPattern+"static", http.FileServer(http.Dir("static"))))
	http.HandleFunc(rootURLPattern, APiHandler)

	srv := &http.Server{
		Addr:         serverurl,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      nil,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Println("Server is not listening anymore: ", err)
	}
}
