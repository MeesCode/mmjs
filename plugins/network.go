package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"
)

var files = make([]globals.Track, 0)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func searchhandler(w http.ResponseWriter, r *http.Request) {
	query, ok := r.URL.Query()["query"]
	if !ok || len(query[0]) < 1 {
		return
	}
	key := query[0]
	files = database.GetSearchResults(key)
	files = files[:min(len(files), 10)]

	res, _ := json.Marshal(files)
	fmt.Fprintf(w, string(res))
}

func addhandler(w http.ResponseWriter, r *http.Request) {
	query, ok := r.URL.Query()["query"]
	if !ok || len(query[0]) < 1 {
		return
	}
	key := query[0]
	i, err := strconv.Atoi(key)

	if err != nil {
		fmt.Fprintf(w, "query could not be converted to integer")
		return
	}

	if len(files) > i {
		track := files[i]
		audioplayer.Addsong(track)

		res, _ := json.Marshal(track)
		fmt.Fprintf(w, string(res))
		return
	}

	fmt.Fprintf(w, "track not found in current search query")
}

func playpausehandler(w http.ResponseWriter, r *http.Request) {
	audioplayer.Pause()
	res, _ := json.Marshal(audioplayer.Playlist[audioplayer.Songindex])
	fmt.Fprintf(w, string(res))
}

func queuehandler(w http.ResponseWriter, r *http.Request) {
	res, _ := json.Marshal(audioplayer.Playlist[audioplayer.Songindex:])
	fmt.Fprintf(w, string(res))
}

func skiphandler(w http.ResponseWriter, r *http.Request) {
	audioplayer.Nextsong()
	res, _ := json.Marshal(audioplayer.Playlist[audioplayer.Songindex])
	fmt.Fprintf(w, string(res))
}

// Webserver starts an entry port for https requests
func Webserver(port int) {
	http.HandleFunc("/search", searchhandler)
	http.HandleFunc("/add", addhandler)
	http.HandleFunc("/skip", skiphandler)
	http.HandleFunc("/queue", queuehandler)
	http.HandleFunc("/playpause", playpausehandler)

	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
