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
}

func (e endpoint) serve() {
	log.Printf("Starting recording: %s\n", e.name)
	f, _ := os.Create(path.Join("/recording-storage/", e.name))
	defer f.Close()

	for {
		select {
		case chunk := <-e.rc:
			log.Printf("chunk: %+v\n", chunk.index)
		case <-e.done:
			log.Printf("Done recording: %s\n", e.name)
			return
		}
	}
}

func serve(ws *websocket.Conn) {
	e := endpoint{
		rc: make(chan chunk),
		done: make(chan int),
	}
	websocket.Message.Receive(ws, &e.name)
	go e.serve()
	for {
		var chunk chunk
		binary.Read(ws, binary.LittleEndian, &chunk.index)
		if chunk.index == -1 {
			e.done <- 0
			return
		}
		binary.Read(ws, binary.LittleEndian, &chunk.length)
		log.Printf("len: %d\n", chunk.length)
		chunk.data = make([]byte, chunk.length)
		io.ReadFull(ws, chunk.data)
		e.rc <- chunk
	}
}

func main() {
	http.Handle("/", websocket.Handler(serve))
	log.Println("Starting recording service on port 443")
	log.Fatal(http.ListenAndServe(":443", nil))
}
