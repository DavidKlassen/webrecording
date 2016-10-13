package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"github.com/gorilla/mux"
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

func deleteVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	videoName := vars["name"]

	err := os.Remove("/storage/" + videoName)
	check(err)
	log.Println("File deleted " + videoName)

}

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", index)
	router.HandleFunc("/deleteVideo/{name}", deleteVideo)
	log.Println("Starting rest service on port 80")
	log.Fatal(http.ListenAndServe(":80", router))

}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
