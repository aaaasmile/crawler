package main

import (
	"log"
	"net/http"
)

func main() {
	//scrap.Scrap()
	port := ":5903"
	handler := http.FileServer(http.Dir("./static"))
	log.Printf("check: http://localhost%s/index.html", port)
	http.ListenAndServe(port, handler)
}
