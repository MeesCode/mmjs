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

func randomhandler(w http.ResponseWriter, r *http.Request) {
	files = database.GetRandomTracks(10)
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

func TogglePausehandler(w http.ResponseWriter, r *http.Request) {
	if !audioplayer.WillPlay() {
		audioplayer.PlaySong(audioplayer.Songindex)
	} else {
		audioplayer.SetPause(true)
	}
	res, _ := json.Marshal(audioplayer.GetPlaying())
	fmt.Fprintf(w, string(res))
}

func queuehandler(w http.ResponseWriter, r *http.Request) {
	res, _ := json.Marshal(audioplayer.Playlist[audioplayer.Songindex:])
	fmt.Fprintf(w, string(res))
}

func skiphandler(w http.ResponseWriter, r *http.Request) {
	audioplayer.Nextsong()
	res, _ := json.Marshal(audioplayer.GetPlaying())
	fmt.Fprintf(w, string(res))
}

func incplaycounterhandler(w http.ResponseWriter, r *http.Request) {
	query, ok := r.URL.Query()["query"]
	if !ok || len(query[0]) < 1 {
		fmt.Fprintf(w, "failed")
		return
	}

	key := query[0]

	i, err := strconv.Atoi(key)

	if err != nil {
		fmt.Fprintf(w, "failed")
		return
	}

	database.IncrementPlayCounter(i)
	fmt.Fprintf(w, "success")
}

func popularhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handler")
	files = database.GetPopularTracks(10)
	res, _ := json.Marshal(files)
	fmt.Fprintf(w, string(res))
}

// Webserver starts an entry port for https requests
func Webserver() {
	http.HandleFunc("/search", searchhandler)
	http.HandleFunc("/add", addhandler)
	http.HandleFunc("/skip", skiphandler)
	http.HandleFunc("/queue", queuehandler)
	http.HandleFunc("/TogglePause", TogglePausehandler)
	http.HandleFunc("/random", randomhandler)
	http.HandleFunc("/incplaycounter", incplaycounterhandler)
	http.HandleFunc("/popular", popularhandler)

	http.ListenAndServe(":"+strconv.Itoa(globals.Config.Webserver.Port), nil)
}
