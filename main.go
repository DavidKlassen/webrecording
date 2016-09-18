package main

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var id = 0

func serve(ws *websocket.Conn) {
	// generate new recording id and send to the client
	id += 1
	err := websocket.Message.Send(ws, strconv.Itoa(id))
	check(err)
	log.Printf("Starting recording: %d\n", id)

	// open new webm file
	f, err := os.Create(strconv.Itoa(id) + ".webm")
	check(err)
	defer f.Close()

	// receive chunks until the connection is closed
	for {
		var buffer []byte
		if err := websocket.Message.Receive(ws, &buffer); err != nil {
			// io.EOF in ws stream means the connection was closed, we can stop the reading loop
			if err == io.EOF {
				log.Printf("Done recording: %d\n", id)
				return
			}

			panic(err)
		}
		// append buffer to webm file
		_, err := f.Write(buffer)
		check(err)
	}
}

func main() {
	http.Handle("/ws", websocket.Handler(serve))
	log.Println("Starting recording service on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
