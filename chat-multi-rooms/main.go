// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
)

var addr = flag.String("addr", ":8080", "http service address")
var house sync.Map
var roomMutexes = make(map[string]*sync.Mutex)
var mutexForRoomMutexes = new(sync.Mutex)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/{room}", serveHome)
	r.HandleFunc("/ws/{room}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomId := vars["room"]
		mutexForRoomMutexes.Lock()
		roomMutex, ok := roomMutexes[roomId]
		if ok {
			roomMutex.Lock()
		} else {
			roomMutexes[roomId] = new(sync.Mutex)
			roomMutexes[roomId].Lock()
		}
		mutexForRoomMutexes.Unlock()
		room, ok := house.Load(roomId)
		var hub *Hub
		if ok {
			hub = room.(*Hub)
		} else {
			hub = newHub(roomId)
			house.Store(roomId, hub)
			go hub.run()
		}
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
