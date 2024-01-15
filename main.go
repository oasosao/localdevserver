package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	Port    = "8080"
	RootDir = "./"
)

func main() {

	flag.StringVar(&Port, "port", "8080", "default port")
	flag.StringVar(&RootDir, "dir", "./", "default")
	flag.Parse()

	hub := newComm()

	go hub.run()
	go hub.watcherFile()

	fs := http.Dir(RootDir)
	http.Handle("/", fileServer(&fs, http.FileServer(&fs)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})

	server := &http.Server{
		Addr:              ":" + Port,
		ReadHeaderTimeout: 3 * time.Second,
	}

	outputStr := `
http listen: 
	http://localhost:%s
	http://%s:%s
websocket listen:
	ws://%s:%s/ws

`
	fmt.Printf(outputStr, Port, getLocalIP(), Port, getLocalIP(), Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
