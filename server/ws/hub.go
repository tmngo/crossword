// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
)

type Subscription struct {
	client *Client
	room   string
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Set of registered clients.
	clients map[*Client]bool

	send chan ClientMessage

	// Outbound messages to the clients.
	broadcast chan RoomMessage

	// Register requests from the clients.
	register chan Subscription

	// Unregister requests from clients.
	unregister chan Subscription

	rooms map[string]*Room
}

type Room struct {
	// Registered clients in the room
	clients map[*Client]bool
	puzzle  string
	state   []byte
	height  int
	width   int
	players map[string]*Player
}

var GlobalHub *Hub
var GlobalPuzzleCache map[string]Puzzle

func NewHub() *Hub {
	return &Hub{
		send:       make(chan ClientMessage),
		broadcast:  make(chan RoomMessage),
		register:   make(chan Subscription),
		unregister: make(chan Subscription),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// Register new clients.
		case subscription := <-h.register:
			log.Println("Registering client.")
			client := subscription.client
			roomName := subscription.room
			h.clients[client] = true
			player := Player{
				Name:  "player01",
				ID:    client.id,
				Color: randomColor(0.6),
				// Color:    Color{0.1, 0.9, 0.4, 1.0},
				Position: Position{0, 0, Across},
			}
			if _, ok := h.rooms[roomName]; ok {
				// Room exists.
				log.Printf("Room %v exists.", subscription.room)
				player.Name = fmt.Sprintf("player%02d", len(h.rooms[roomName].players))
				h.rooms[roomName].clients[client] = true
				h.rooms[roomName].players[client.id] = &player
			} else {
				// Create new room.
				log.Printf("Creating room %v.", subscription.room)
				room := Room{
					map[*Client]bool{client: true},
					"",
					make([]byte, 0),
					0,
					0,
					map[string]*Player{client.id: &player},
				}
				h.rooms[subscription.room] = &room
			}
			log.Print(client.id)
			message, _ := json.Marshal(TaggedMessage{
				Tag:  TagRegister,
				Data: Register{client.id},
			})
			client.send <- message
			room := h.rooms[subscription.room]
			message, _ = json.Marshal(TaggedMessage{
				Tag: TagPlayerUpdate,
				Data: PlayerUpdate{
					"",
					h.rooms[subscription.room].players,
				},
			})
			h.broadcastMessage(room, "", message)
		case subscription := <-h.unregister:
			log.Println("Unregistering client.")
			client := subscription.client
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if _, ok := h.rooms[subscription.room].clients[client]; ok {
					delete(h.rooms[subscription.room].clients, client)
					delete(h.rooms[subscription.room].players, client.id)
				}
				close(client.send)
				room := h.rooms[subscription.room]
				taggedMessage := TaggedMessage{
					Tag: TagPlayerUpdate,
					Data: PlayerUpdate{
						string(room.state),
						room.players,
					},
				}
				output, err := json.Marshal(taggedMessage)
				if err != nil {
					return
				}
				client.hub.broadcastMessage(room, "", output)
			}
		case clientMessage := <-h.send:
			client := clientMessage.client
			select {
			case client.send <- clientMessage.message:
			default:
				log.Print("Default send.")
				close(client.send)
				delete(h.clients, client)
			}
		case data := <-h.broadcast:
			room := h.rooms[data.room]
			h.broadcastMessage(room, "", data.message)
		}
	}
}

func randomColor(lightness float64) Color {
	r := 0.05 + 0.9*rand.Float64()
	g := 0.05 + 0.9*rand.Float64()
	b := 0.05 + 0.9*rand.Float64()
	rlin := math.Pow((r+0.055)/1.055, 2.4)
	glin := math.Pow((g+0.055)/1.055, 2.4)
	blin := math.Pow((b+0.055)/1.055, 2.4)
	// rlin := r
	// glin := g
	// blin := g
	luminance := 0.2126*rlin + 0.7152*glin + 0.0722*blin
	perceivedLightness := luminance * 903.3
	if luminance > 0.008856 {
		perceivedLightness = 1.16*math.Pow(luminance, 0.333) - 0.16
	}
	scale := lightness / perceivedLightness
	color := Color{
		r * scale,
		g * scale,
		b * scale,
		1.0,
	}
	return color
	// color = hsl(rand.Float64(), 0.9, 0.5)
	// log.Printf("Luminance: %v", luminance)
	// log.Printf("Perceived: %v", perceivedLightness)
	// log.Print(color)
}

func hsl(h, s, l float64) Color {
	if s == 0.0 {
		return Color{l, l, l, 1.0}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hue2rgb(p, q, h+0.3333)
	g := hue2rgb(p, q, h)
	b := hue2rgb(p, q, h+0.3333)
	return Color{r, g, b, 1.0}
}

func hue2rgb(p, q, h float64) float64 {
	if h < 0 {
		h += 1
	} else if h > 1 {
		h -= 1
	}
	if h < 1.0/6.0 {
		return p + (q-p)*6*h
	} else if h < 0.5 {
		return q
	} else if h < 2.0/3.0 {
		return p + (q-p)*(4.0-6.0*h)
	}
	return p
}

func (h *Hub) broadcastMessage(room *Room, excludedClient string, message []byte) {
	log.Printf("Broadcasting to all %v clients.\n", len(room.clients))
	for client := range room.clients {
		if client.id == excludedClient {
			continue
		}
		select {
		case client.send <- message:
		default:
			// The default case is run if no other case is ready.
			log.Print("Default broadcast.")
			close(client.send)
			delete(h.clients, client)
		}
	}
}
