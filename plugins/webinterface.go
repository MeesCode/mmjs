package plugins

import (
	"log"
	"time"
    "net/http"
	"encoding/json"
	"strconv"

    "github.com/gorilla/websocket"
	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/globals"
)

type Stats struct {
    Queue []globals.Track `json:"Queue"`
    Playing bool `json:"Playing"`
    Index int `json:"Index"`
}

var statobject Stats
var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true	
	},
}

func commands(w http.ResponseWriter, r *http.Request) {
	// upgrade the connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln("Could upgrade the connection to a websocket", err)
	}

	defer conn.Close()

	// for every mesage received execute command
	for {
		// Read message from browser
		_, msg, err := conn.ReadMessage()
		if err != nil {
            log.Println("No message recieved:", msg, err)
			return
        }

		switch string(msg) {
			case "play":
				if !audioplayer.IsLoaded(){
					audioplayer.PlaySong(audioplayer.Songindex)
					break;
				}
				audioplayer.Resume()
				break;
			case "pause":
				audioplayer.Pause()
				break;
			case "next":
				audioplayer.Nextsong()
				break;
			case "previous":
				audioplayer.Previoussong()
				break;
		}
	}
}

// register client for receiving periodic stats
func stats(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
			log.Fatal(err)
	}

	// register client
	clients[ws] = true	
}

// send periodic stats
func broadcaster() {
	for {
		// wait for periodic timer
		time.Sleep(1 * time.Second)

		// create stat object
		statobject.Queue = audioplayer.Playlist
		statobject.Index = audioplayer.Songindex
		statobject.Playing = audioplayer.IsPlaying()

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
    http.HandleFunc("/stats", stats)
	http.HandleFunc("/commands", commands)

	// start broadcaster routine
	go broadcaster()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(globals.Config.Webinterface.Port), nil))
}