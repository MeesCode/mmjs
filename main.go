package main

import (
	"mp3bak2/audioplayer"
)

// "local" global variables
var root = "/home/mees"
var tracklist = make([]string, 0)
var songindex = 0
var formats = [6]string{".wav", ".mp3", ".ogg", ".weba", ".webm", ".flac"}

func main() {

	go audioplayer.Init()
	startInterface()

}

func contains(arr [6]string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
