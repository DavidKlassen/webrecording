package main

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"encoding/binary"
	"os"
	"path"
)

type chunk struct {
	index int32
	length int32
	data []byte
}

type endpoint struct {
	rc chan chunk
	done chan int
	name string
	chunks []chunk
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
			// do assembling here
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
	e, ok := endpoints[name];
	if !ok {
		e = endpoint{
			rc: make(chan chunk),
			done: make(chan int),
			name: name,
		}
		endpoints[name] = e
		go e.serve()
	}

	for {
		var chunk chunk
		err = binary.Read(ws, binary.LittleEndian, &chunk.index)
		check(err)
		if chunk.index == -1 {
			e.done <- 0
			return
		}
		err = binary.Read(ws, binary.LittleEndian, &chunk.length)
		check(err)
		log.Printf("len: %d\n", chunk.length)
		chunk.data = make([]byte, chunk.length)
		_, err = io.ReadFull(ws, chunk.data)
		check(err)
		e.rc <- chunk
	}
}

func main() {
	http.Handle("/", websocket.Handler(serve))
	log.Println("Starting recording service on port 443")
	log.Fatal(http.ListenAndServe(":443", nil))
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
