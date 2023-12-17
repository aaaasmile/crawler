package web

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	svgstr := `<svg xmlns="http://www.w3.org/2000/svg">
	<ellipse cx="200" cy="80" rx="100" ry="50" style="fill:yellow;stroke:purple;stroke-width:2" />
  </svg>`
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

func StartServer() {
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
	if err := srv.ListenAndServe(); err != nil {
		log.Println("Server is not listening anymore: ", err)
	}
}
