// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tmngo/crossword-server/util"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		log.Print(origin)
		return origin == "http://localhost:3000"
	},
}

type Message struct {
	Tag  MessageTag      `json:"tag"`
	Data json.RawMessage `json:"data"`
}

type Text string

type PuzzleRequest struct {
	Sources []string `json:"sources"`
	Day     int      `json:"day"`
	Month   int      `json:"month"`
	Year    int      `json:"year"`
}

type Register struct {
	Id string `json:"id"`
}

type PlayerUpdate struct {
	State   string             `json:"state"`
	Players map[string]*Player `json:"players"`
}

type PlayerAction struct {
	Position Position
	Value    string
}

type RoomMessage struct {
	room    string
	message []byte
}

type ClientMessage struct {
	client  *Client
	message []byte
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	id string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send       chan []byte
	sendBinary chan []byte
}

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type TaggedMessage struct {
	Tag  MessageTag  `json:"tag"`
	Data interface{} `json:"data"`
}

type MessageTag int

const (
	TagText MessageTag = iota
	TagPuzzle
	TagRegister
	TagState
	TagPlayerUpdate
	TagPlayerAction
	TagPlayerClick
	TagPuzzleLoad
	TagNewPuzzle
)

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (s Subscription) readPump() {
	c := s.client
	defer func() {
		log.Print("Deferred readPump.")
		c.hub.unregister <- s
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// _, message, err := c.conn.ReadMessage()
		log.Print("Entering readPump loop.")
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Print("Error reading JSON.")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				// break
			}
			log.Printf("expected error: %v", err)
			break
		}

		log.Print("---")
		log.Printf("Received type \"%v\" from room %s\n", msg.Tag, s.room)

		var err error
		switch msg.Tag {
		case TagText:
			err = s.handleText(msg.Data)
		case TagPuzzle:
			err = s.handlePuzzleRequest(msg.Data)
		case TagPlayerAction:
			err = s.handlePlayerAction(msg.Data)
		case TagPlayerClick:
			err = s.handlePlayerClick(msg.Data)
		case TagNewPuzzle:
			err = s.handleNewPuzzle(msg.Data)
		case TagPuzzleLoad:
			err = s.handlePuzzleLoad(msg.Data)
		}

		// log.Printf("Received type (%s): %s from room %s\n", msg.Type, message, s.room)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
}

func (s *Subscription) handleText(input json.RawMessage) error {
	var text string
	if err := json.Unmarshal([]byte(input), &text); err != nil {
		return err
	}
	data := bytes.TrimSpace(bytes.Replace([]byte(text), newline, space, -1))
	log.Print(string(data))
	err := s.broadcastToRoom(TagText, string(data))
	return err
}

func (s *Subscription) handlePuzzleRequest(input json.RawMessage) error {
	message := []byte{}
	var puzzleRequest PuzzleRequest
	if err := json.Unmarshal([]byte(input), &puzzleRequest); err != nil {
		return err
	}

	url := fmt.Sprintf("https://herbach.dnsalias.com/wsj/wsj%02d%02d%02d.puz",
		puzzleRequest.Year%100, puzzleRequest.Month, puzzleRequest.Day)

	log.Print(url)

	id := fmt.Sprintf("%v-%v-%v-%v",
		puzzleRequest.Sources[0],
		puzzleRequest.Year,
		puzzleRequest.Month,
		puzzleRequest.Day)

	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("Error %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}

	resp.Body.Close()
	message = body

	// // Offline
	// err = os.WriteFile("ws/wsj.puz", body, 0644)
	// if err != nil {
	// 	log.Printf("Error %v", err)
	// }
	// offlineMessage, err := os.ReadFile("ws/wsj.puz")
	// if err != nil {
	// 	return err
	// }
	// log.Print(len(message))

	// log.Print(string(message) == string(offlineMessage))
	// message = offlineMessage

	puzzle, err := parsePuz(message, id)
	if err != nil {
		return err
	}

	log.Printf("%#v", puzzle)
	room := GlobalHub.rooms[s.room]
	room.puzzle = puzzle.Grid
	room.state = make([]byte, len(puzzle.Grid))
	room.height = puzzle.Height
	room.width = puzzle.Width

	err = s.broadcastToRoom(TagPuzzle, puzzle)
	return err
}

func parsePuz(data []byte, id string) (Puzzle, error) {
	width := int(data[44])
	height := int(data[45])
	n := width * height
	log.Printf("%v x %v", width, height)

	gextIndex := bytes.Index(data[52+2*n:], []byte("GEXT"))

	var stringsSection []byte
	var gextSection []byte
	if gextIndex != -1 {
		stringsSection = util.Utf8(data[52+2*n : 52+2*n+gextIndex])
		gextSection = data[52+2*n+gextIndex:]
	} else {
		stringsSection = util.Utf8(data[52+2*n:])
		gextSection = []byte{}
	}

	clueString := string(stringsSection)
	lines := strings.Split(clueString, "\u0000")

	clueLines := lines[3 : len(lines)-1]
	for i, clue := range clueLines {
		log.Printf("[%v] %#v (%v)", i, clue, []byte(clue))
	}

	var acrossClues []Clue
	var downClues []Clue
	num := 1
	clueIndex := 0

	grid := string(data[52 : 52+n])
	for row := 0; row < int(height); row++ {
		for col := 0; col < int(width); col++ {
			index := row*int(width) + col
			if grid[index] == '.' {
				continue
			}
			clue := Clue{
				Number: num,
				Row:    row,
				Column: col,
			}
			hasClue := false
			// Across clue.
			if col == 0 || grid[index-1] == '.' {
				clue.Text = clueLines[clueIndex]
				clue.Direction = Across
				for k := col; k < width && grid[row*width+k] != '.'; k++ {
					clue.Length++
				}
				acrossClues = append(acrossClues, clue)
				clueIndex++
				hasClue = true
			}
			// Down clue.
			if row == 0 || grid[index-int(width)] == '.' {
				clue.Length = 0
				clue.Text = clueLines[clueIndex]
				clue.Direction = Down
				for k := row; k < height && grid[k*width+col] != '.'; k++ {
					clue.Length++
				}
				downClues = append(downClues, clue)
				clueIndex++
				hasClue = true
			}
			if hasClue {
				num++
			}
		}
	}

	for i, clue := range acrossClues {
		log.Printf("[%v] %#v", i, clue)
	}

	puzzle := Puzzle{
		ID:          id,
		Width:       width,
		Height:      height,
		NumClues:    int(binary.LittleEndian.Uint16(data[46:48])),
		Grid:        grid,
		AcrossClues: acrossClues,
		DownClues:   downClues,
		Title:       lines[0],
		Creators:    lines[1],
		Attribution: lines[2],
		Gext:        string(gextSection),
	}
	return puzzle, nil
}

func (s *Subscription) handlePlayerAction(input json.RawMessage) error {
	var key Text
	if err := json.Unmarshal([]byte(input), &key); err != nil {
		return err
	}
	log.Printf("handlePlayerAction: %v", key)
	room := GlobalHub.rooms[s.room]
	if room == nil {
		return errors.New("Room is nil.")
	}
	player := room.players[s.client.id]
	if player == nil {
		return errors.New("Player is nil.")
	}
	row := player.Position.Row
	col := player.Position.Col
	dir := player.Position.Dir

	switch string(key) {
	case KeySpace:
		s.setPlayerPosition(row, col, dir.flip())
	case KeyBackspace, KeyDelete:
		// s.setCellValue(row, col, ' ')
		room.state[row*room.width+col] = 0
		if dir == Across {
			s.setPlayerPosition(row, col-1, dir)
		} else {
			s.setPlayerPosition(row-1, col, dir)
		}
	case KeyArrowDown:
		s.setPlayerPosition(row+1, col, dir)
	case KeyArrowLeft:
		s.setPlayerPosition(row, col-1, dir)
	case KeyArrowRight:
		s.setPlayerPosition(row, col+1, dir)
	case KeyArrowUp:
		s.setPlayerPosition(row-1, col, dir)
	default:
		if len(key) == 1 {
			code := key[0]
			if code < 97 || code > 122 {
				return errors.New("Key code is not a lowercase letter.")
			}
			room.state[row*room.width+col] = code - 32
			log.Printf("code: %v %v", string(code), string(code-32))
			log.Printf(string(room.state))
			if dir == Across {
				s.setPlayerPosition(row, col+1, dir)
			} else {
				s.setPlayerPosition(row+1, col, dir)
			}
		}
	}

	return nil
}

func (s *Subscription) handlePlayerClick(input json.RawMessage) error {
	var position Position
	if err := json.Unmarshal([]byte(input), &position); err != nil {
		return err
	}
	log.Printf("handlePlayerClick: %#v", position)
	room := GlobalHub.rooms[s.room]
	if room == nil {
		return errors.New("Room is nil.")
	}
	player := room.players[s.client.id]
	if player == nil {
		return errors.New("Player is nil.")
	}
	row := position.Row
	col := position.Col
	if row == player.Position.Row && col == player.Position.Col {
		s.setPlayerPosition(row, col, player.Position.Dir.flip())
	} else {
		s.setPlayerPosition(row, col, player.Position.Dir)
	}
	return nil
}

func (s *Subscription) handleNewPuzzle(input json.RawMessage) error {
	var puzzleRequest PuzzleRequest
	log.Print(string(input))
	if err := json.Unmarshal([]byte(input), &puzzleRequest); err != nil {
		return err
	}
	var puzzleData []PuzzleData

	for _, source := range puzzleRequest.Sources {
		url := fmt.Sprintf("https://herbach.dnsalias.com/wsj/wsj%02d%02d%02d.puz",
			puzzleRequest.Year%100, puzzleRequest.Month, puzzleRequest.Day)

		log.Print(url)

		id := fmt.Sprintf("%v-%v-%v-%v",
			source,
			puzzleRequest.Year,
			puzzleRequest.Month,
			puzzleRequest.Day)

		var puzzle Puzzle
		puzzle, ok := GlobalPuzzleCache[id]
		if !ok {
			resp, err := httpClient.Get(url)
			if err != nil {
				log.Printf("Error %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("Puzzle not found")
				continue
			}
			message, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error %v", err)
			}
			resp.Body.Close()
			puzzle, err = parsePuz(message, id)
			if err != nil {
				return err
			}
			GlobalPuzzleCache[id] = puzzle
		}

		puzzleData = append(puzzleData, PuzzleData{
			ID:     id,
			Source: source,
			Year:   puzzleRequest.Year,
			Month:  puzzleRequest.Month,
			Day:    puzzleRequest.Day,
			Title:  puzzle.Title,
		})
	}
	err := s.broadcastToRoom(TagNewPuzzle, puzzleData)
	return err
}

func (s *Subscription) handlePuzzleLoad(input json.RawMessage) error {
	return nil
}

func (s *Subscription) setCellValue(row, col int, value byte) {
	room := GlobalHub.rooms[s.room]
	if room == nil {
		return
	}
	index := row*room.width + col
	room.state[index] = value
}

func (s *Subscription) broadcastToRoom(tag MessageTag, data interface{}) error {
	message := TaggedMessage{tag, data}
	encoded, err := json.Marshal(message)
	if err != nil {
		return err
	}
	s.client.hub.broadcast <- RoomMessage{s.room, encoded}
	return nil
}

func (s *Subscription) setPlayerPosition(row, col int, dir Direction) {
	room := GlobalHub.rooms[s.room]
	if room == nil {
		log.Print("room is nil")
		return
	}
	w := room.width
	h := room.height
	if row < 0 || col < 0 || row >= h || col >= w {
		log.Print("position out of bounds")
		return
	}

	player := room.players[s.client.id]
	if player == nil {
		log.Print("player is nil")
		return
	}
	currentRow := player.Position.Row
	currentCol := player.Position.Col

	for room.puzzle[row*w+col] == '.' {
		if col == currentCol {
			if row > currentRow && row < h-1 {
				row += 1
			} else if row < currentRow && row > 0 {
				row -= 1
			} else {
				return
			}
		} else if row == currentRow {
			if col > currentCol && col < w-1 {
				col += 1
			} else if col < currentCol && col > 0 {
				col -= 1
			} else {
				return
			}
		}
	}

	player.Position = Position{row, col, dir}
	taggedMessage := TaggedMessage{
		Tag: TagPlayerUpdate,
		Data: PlayerUpdate{
			string(room.state),
			room.players,
		},
	}

	log.Print(string(room.state))

	output, err := json.Marshal(taggedMessage)
	if err != nil {
		return
	}

	s.client.hub.broadcastMessage(room, "", output)

	for id, p := range room.players {
		log.Printf("%v: %v", id, p.Position)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (s Subscription) writePump() {
	c := s.client
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Print("Deferred writePump.")
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			log.Print("Sending message using writePump.")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Print("Send not ok.")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}

			// The message must be valid UTF-8 encoded text.
			// w, err := c.conn.NextWriter(websocket.TextMessage)
			// if err != nil {
			// 	return
			// }
			// w.Write(message)

			// // Add queued chat messages to the current websocket message.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write(newline)
			// 	w.Write(<-c.send)
			// }

			// if err := w.Close(); err != nil {
			// 	return
			// }
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Print("[serveWs]: ")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:  hub,
		id:   util.NewId(8),
		conn: conn,
		send: make(chan []byte, 256),
	}
	log.Println(r.URL)
	log.Printf("%#v", client)
	subscription := Subscription{client, r.URL.Path}
	client.hub.register <- subscription

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go subscription.writePump()
	go subscription.readPump()
}
