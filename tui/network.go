package tui

import (
	"fmt"
	"strconv"
	"net/http"
	"github.com/MeesCode/mmjs/globals"
	"encoding/json"
	"math"
	"github.com/MeesCode/mmjs/audioplayer"
)

func searchhandler(w http.ResponseWriter, r *http.Request) {
	query, ok := r.URL.Query()["query"]
	if !ok || len(query[0]) < 1 {
		return
	}
	key := query[0]
	searchQuery(key)

	res, _ := json.Marshal(filelistFiles[:int(math.Min(float64(len(filelistFiles)), float64(10)))])
	fmt.Fprintf(w, string(res))
}

func addhandler(w http.ResponseWriter, r *http.Request){
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

	if len(filelistFiles) > i {
		myTui.filelist.SetCurrentItem(i)
		track := filelistFiles[i]
		addsong()

		res, _ := json.Marshal(track)
		fmt.Fprintf(w, string(res))
		return
	}

	fmt.Fprintf(w, "track not found in current search query")
}

func playpauzehandler(w http.ResponseWriter, r *http.Request) {
	_, _, playing := audioplayer.GetPlaytime()
	if !playing {
		playsong()
		fmt.Fprintf(w, "play")
		return
	} else {
		audioplayer.Pause()
		fmt.Fprintf(w, "pauze")
	}
}

func queuehandler(w http.ResponseWriter, r *http.Request){
	res, _ := json.Marshal(playlistFiles[:int(math.Min(float64(len(playlistFiles)), float64(10)))])
	fmt.Fprintf(w, string(res))
}

func skiphandler(w http.ResponseWriter, r *http.Request) {
	nextsong()
	res, _ := json.Marshal(playlistFiles[songindex])
	fmt.Fprintf(w, string(res))
}

func Webserver() {
	http.HandleFunc("/search", searchhandler)
	http.HandleFunc("/add", addhandler)
	http.HandleFunc("/skip", skiphandler)
	http.HandleFunc("/queue", queuehandler)
	http.HandleFunc("/playpauze", playpauzehandler)

	http.ListenAndServe(":" + globals.Port, nil)
}
