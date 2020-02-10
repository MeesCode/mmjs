package main

import (
	"mp3bak2/audioplayer"
	"mp3bak2/tui"
)

func main() {

	// initialize audio player
	go audioplayer.Init()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start()

}
