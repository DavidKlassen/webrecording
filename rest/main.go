package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
)

type ByModTime []os.FileInfo

func (mt ByModTime) Len() int           { return len(mt) }
func (mt ByModTime) Swap(i, j int)      { mt[i], mt[j] = mt[j], mt[i] }
func (mt ByModTime) Less(i, j int) bool { return mt[i].ModTime().After(mt[j].ModTime()) }

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	files, _ := ioutil.ReadDir("/storage")
	sort.Sort(ByModTime(files))
	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func main() {
	http.HandleFunc("/", index)
	log.Println("Starting rest service on port 80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
