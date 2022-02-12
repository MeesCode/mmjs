package plugins

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"strings"

	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/globals"
	"github.com/gorilla/websocket"
)

type Stats struct {
	Queue    []globals.Track
	Playing  bool
	Index    int
	Length   time.Duration
	Progress time.Duration
}

var clients = make(map[*websocket.Conn]bool)
var previousQueue []globals.Track

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// check if two playlists are the same
func identicalPlaylists(i1 []globals.Track, i2 []globals.Track) bool {
	if len(i1) != len(i2) { return false }
	for i, _ := range i1 {
		log.Println("compare " + strconv.Itoa(i1[i].ID) + " and " + strconv.Itoa(i2[i].ID))
		if i1[i].ID != i2[i].ID { return false }
	}
	return true
}

// serve the client webapp
func page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "plugins/webinterface.html")
}

// register client 
func stats(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// register client
	clients[ws] = true

	// send initial stats
	var statobject Stats
	statobject.Queue = audioplayer.Playlist
	statobject.Index = audioplayer.Songindex
	statobject.Playing = audioplayer.IsPlaying()
	statobject.Progress, statobject.Length = audioplayer.GetPlaytime()

	queue, _ := json.Marshal(statobject)

	err = ws.WriteMessage(websocket.TextMessage, []byte(queue))
	if err != nil {
		log.Printf("Websocket error: %s", err)
		ws.Close()
		delete(clients, ws)
	}

	defer ws.Close()
	receiver(ws)
}

// receive incoming command from a specific client
func receiver(ws *websocket.Conn) {
	for {
		// Read message from browser
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("No message recieved:", msg, err)
			return
		}

		c := strings.Split(string(msg), ":")
		command := c[0]
		var args []string
		if len(c) > 1 {
			args = strings.Split(c[1], ",")
		}

		switch command {
		case "play":
			if len(audioplayer.Playlist) == 0 { break }
			if !audioplayer.WillPlay() {
				audioplayer.PlaySong(audioplayer.Songindex)
			} else {
				audioplayer.TogglePause()
			}
			break
		case "pause":
			audioplayer.SetPause(true)
			break
		case "next":
			audioplayer.Nextsong()
			break
		case "previous":
			audioplayer.Previoussong()
			break
		case "shuffle":
			audioplayer.Shuffle()
			break
		case "clear":
			audioplayer.Clear()
			break
		case "playtrack":
			index, err := strconv.Atoi(args[0])
			if err == nil { audioplayer.PlaySong(index) }
		}
	}
}

// send periodic stats to all clients
func broadcaster() {
	for {
		// wait for periodic timer
		time.Sleep(1 * time.Second)

		// create stat object
		var statobject Stats
		statobject.Index = audioplayer.Songindex
		statobject.Playing = audioplayer.IsPlaying()
		statobject.Progress, statobject.Length = audioplayer.GetPlaytime()

		// update queue only if necessary
		if identicalPlaylists(previousQueue, audioplayer.Playlist) && previousQueue != nil {
			statobject.Queue = nil
		} else {
			statobject.Queue = audioplayer.Playlist
			previousQueue = append([]globals.Track(nil), audioplayer.Playlist...)
		}

		// get current stats in json format
		queue, _ := json.Marshal(statobject)

		// send to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(queue))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func Webinterface() {
	http.HandleFunc("/", page)
	http.HandleFunc("/socket", stats)

	// start broadcaster routine
	go broadcaster()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(globals.Config.Webinterface.Port), nil))
}
