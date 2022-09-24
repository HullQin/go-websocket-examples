// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Identity of room.
	roomId string

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub(roomId string) *Hub {
	return &Hub{
		roomId:     roomId,
		broadcast:  make(chan []byte),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	defer func() {
		close(h.unregister)
		close(h.broadcast)
	}()
	for {
		select {
		case client := <-h.unregister:
			roomMutex := roomMutexes[h.roomId]
			roomMutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if len(h.clients) == 0 {
					house.Delete(h.roomId)
					roomMutex.Unlock()
					mutexForRoomMutexes.Lock()
					if roomMutex.TryLock() {
						if len(h.clients) == 0 {
							delete(roomMutexes, h.roomId)
						}
						roomMutex.Unlock()
					}
					mutexForRoomMutexes.Unlock()
					return
				}
			}
			roomMutex.Unlock()
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
