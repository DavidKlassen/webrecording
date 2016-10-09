package main

import (
	"bytes"
	"encoding/json"
	"flag"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	port       = flag.String("port", "443", "The application port")
	storageDir = flag.String("storage", "/recording-storage/", "Path to the storage directory")
)

var (
	restURL    = "http://rest/"
	restClient = http.Client{}
)

type recording struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
}

func serve(ws *websocket.Conn) {
	// generate new recording id and send to the client
	id := uuid.NewV4()
	err := websocket.Message.Send(ws, id.String())
	check(err)
	log.Printf("Starting recording: %s\n", id.String())

	// open new webm file
	fileName := path.Join(*storageDir, id.String()+".webm")
	f, err := os.Create(fileName)
	check(err)
	defer f.Close()

	body, err := json.Marshal(recording{
		ID:       id.String(),
		FileName: id.String() + ".webm",
	})
	check(err)
	req, err := http.NewRequest(http.MethodPut, restURL, bytes.NewReader(body))
	check(err)
	_, err = restClient.Do(req)
	check(err)

	// receive chunks until the connection is closed
	for {
		var buffer []byte
		if err := websocket.Message.Receive(ws, &buffer); err != nil {
			// io.EOF in ws stream means the connection was closed, we can stop the reading loop
			if err == io.EOF {
				log.Printf("Done recording: %s\n", id.String())
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
	flag.Parse()

	http.Handle("/", websocket.Handler(serve))
	log.Printf("Starting recording service on port :%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
