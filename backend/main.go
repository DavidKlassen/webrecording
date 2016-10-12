package main

import (
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func serve(ws *websocket.Conn) {
	// generate new recording id and send to the client
	id := uuid.NewV4()
	websocket.Message.Send(ws, id.String())
	log.Printf("Starting recording: %s\n", id.String())

	// open new webm file
	fileName := path.Join("/recording-storage/", id.String()+".webm")
	f, _ := os.Create(fileName)
	defer f.Close()

	// stream to file
	n, _ := io.Copy(f, ws)
	log.Printf("Done recording: %s; bytes writtes: %d\n", id.String(), n)
}

func main() {
	http.Handle("/", websocket.Handler(serve))
	log.Println("Starting recording service on port 443")
	log.Fatal(http.ListenAndServe(":443", nil))
}
