package main

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func serve(ws *websocket.Conn) {
	var fileName string
	websocket.Message.Receive(ws, &fileName)
	log.Printf("Starting recording: %s\n", fileName)

	// open new webm file
	f, _ := os.Create(path.Join("/recording-storage/", fileName))
	defer f.Close()

	// stream to file
	n, _ := io.Copy(f, ws)
	log.Printf("Done recording: %s; bytes written: %d\n", fileName, n)
}

func main() {
	http.Handle("/", websocket.Handler(serve))
	log.Println("Starting recording service on port 443")
	log.Fatal(http.ListenAndServe(":443", nil))
}
