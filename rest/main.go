package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"fmt"
)

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	files, _ := ioutil.ReadDir("/storage")
	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func main() {
	http.HandleFunc("/", index)
	log.Println("Starting rest service on port 80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
