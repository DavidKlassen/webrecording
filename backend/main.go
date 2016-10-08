package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/bwmarrin/snowflake"
	_ "github.com/lib/pq"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	port       = flag.String("port", "443", "The application port")
	nodeId     = flag.Int64("node", 0, "Snowflake node")
	storageDir = flag.String("storage", "/recording-storage/", "Path to the storage directory")
)

var (
	restURL = "http://rest/"
	restClient = http.Client{}
)

var (
	node *snowflake.Node
)

type recording struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

func serve(ws *websocket.Conn) {
	// generate new recording id and send to the client
	id := node.Generate()
	err := websocket.Message.Send(ws, id.String())
	check(err)
	log.Printf("Starting recording: %d\n", id)

	// open new webm file
	fileName := path.Join(*storageDir, id.String()+".webm")
	f, err := os.Create(fileName)
	check(err)
	defer f.Close()

	rec := recording{
		id.Int64(),
		id.String() + ".webm",
	}

	body, err := json.Marshal(rec)
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
	flag.Parse()

	var err error
	node, err = snowflake.NewNode(*nodeId)
	check(err)

	http.Handle("/ws", websocket.Handler(serve))
	log.Printf("Starting recording service on port :%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
