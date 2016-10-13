package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"path"
)

type ByModTime []os.FileInfo

func (mt ByModTime) Len() int           { return len(mt) }
func (mt ByModTime) Swap(i, j int)      { mt[i], mt[j] = mt[j], mt[i] }
func (mt ByModTime) Less(i, j int) bool { return mt[i].ModTime().After(mt[j].ModTime()) }

func index(w http.ResponseWriter) {
	files, _ := ioutil.ReadDir("/storage")
	sort.Sort(ByModTime(files))
	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func del(w http.ResponseWriter, name string) {
	err := os.Remove(path.Join("/storage/", name))
	check(err)
	log.Println("File deleted " + name)
	w.WriteHeader(http.StatusNoContent)
}

func serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodGet {
		index(w)
	} else if r.Method == http.MethodDelete {
		del(w, r.RequestURI)
	} else if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, OPTIONS")
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", serve)
	log.Println("Starting rest service on port 80")
	log.Fatal(http.ListenAndServe(":80", nil))

}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
