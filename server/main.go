// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/tmngo/crossword-server/ws"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL, r.URL)
	log.Printf("room: %s\n", r.URL.Path)
	// if r.URL.Path == "/favicon.ico" {
	// 	http.ServeFile(w, r, "favicon.ico")
	// 	return
	// }
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	log.Println("FAVICON")
	http.ServeFile(w, r, "favicon.png")
}

func main() {
	flag.Parse()
	ws.GlobalHub = ws.NewHub()
	go ws.GlobalHub.Run()
	ws.GlobalPuzzleCache = make(map[string]ws.Puzzle)

	// Matches all paths not matched by other patterns.
	http.HandleFunc("/favicon.png", serveFavicon)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ws.GlobalHub, w, r)
	})

	log.Printf("Listening on %s.", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
