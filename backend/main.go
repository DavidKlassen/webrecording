package main

import (
	"encoding/binary"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
)

type chunk struct {
	index  int32
	length int32
	data   []byte
}

type endpoint struct {
	rc     chan chunk
	done   chan int
	name   string
	chunks []chunk
}

type byIndex []chunk

func (i byIndex) Len() int {
	return len(i)
}
func (chunk byIndex) Swap(i, j int) {
	chunk[i], chunk[j] = chunk[j], chunk[i]
}
func (chunk byIndex) Less(i, j int) bool {
	return chunk[i].index < chunk[j].index
}

func (e endpoint) serve() {
	log.Printf("Starting recording: %s\n", e.name)
	f, _ := os.Create(path.Join("/recording-storage/", e.name))
	defer f.Close()

	for {
		select {
		case chunk := <-e.rc:
			log.Printf("chunk: %+v\n", chunk.index)
			e.chunks = append(e.chunks, chunk)
		case <-e.done:
			log.Printf("Done recording: %s\n; got %d chunks", e.name, len(e.chunks))
			sort.Sort(byIndex(e.chunks))
			for _, chunk := range e.chunks {
				_, err := f.Write(chunk.data)
				check(err)
			}
			delete(endpoints, e.name)
			return
		}
	}
}

var endpoints = make(map[string]endpoint)

func serve(ws *websocket.Conn) {
	var name string
	err := websocket.Message.Receive(ws, &name)
	check(err)
	e := fetchEndpoint(name)
	for {
		if !readData(ws, e) {
			return
		}
	}
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Name, Access-Control-Allow-Headers, Authorization, X-Requested-With")

	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	name := r.Header.Get("X-Name")
	readData(r.Body, fetchEndpoint(name))
}

func fetchEndpoint(name string) endpoint {
	e, ok := endpoints[name]
	if !ok {
		e = endpoint{
			rc:   make(chan chunk),
			done: make(chan int),
			name: name,
		}
		endpoints[name] = e
		go e.serve()
	}
	return e
}

func readData(r io.Reader, e endpoint) bool {
	var chunk chunk
	err := binary.Read(r, binary.LittleEndian, &chunk.index)
	check(err)
	if chunk.index == -1 {
		e.done <- 0
		return false
	}
	err = binary.Read(r, binary.LittleEndian, &chunk.length)
	check(err)
	log.Printf("len: %d\n", chunk.length)

	chunk.data = make([]byte, chunk.length)
	_, err = io.ReadFull(r, chunk.data)
	check(err)
	e.rc <- chunk

	return true
}

func main() {
	http.Handle("/", websocket.Handler(serve))
	http.HandleFunc("/chunk", serveHttp)
	log.Println("Starting recording service on port 443")
	log.Fatal(http.ListenAndServe(":443", nil))
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
