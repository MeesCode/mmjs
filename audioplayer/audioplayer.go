package audioplayer

import (
	"log"
	"os"
	"time"

	"mp3bak2/globals"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var speakerevent = make(chan string) //
const bufferSize = 100 * time.Millisecond

// stopSpeaker : send a stop command to a running adio player and
// block until it has fully stopped
func stopPlayer() {
	globals.Speakercommand <- "stop"
	event := <-speakerevent
	for {
		if event != "stopped" {
			event = <-speakerevent
		}
		return
	}
}

// dummySpeaker : an empty husk of what could be an audio player
func dummySpeaker() {
	command := <-globals.Speakercommand
	for {
		if command == "stop" {
			speakerevent <- "stopped"
			return
		}
		command = <-globals.Speakercommand
	}
}

// Init : initialize the dummy speaker so stopSpeaker() doesn't break
// when no audioplayer has started yet
func Init() {
	go dummySpeaker()
}

// Play : stop the previous audio player and replace by a new one
func Play(file string) {

	// stops the previous player (or dummy)
	stopPlayer()

	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening the file: %s", err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatalf("error decoding file: %s", err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(bufferSize))
	if err != nil {
		log.Fatalf("failed to initialize audio device: %s", err)
	}

	// set the length of the track
	speaker.Lock()
	length := format.SampleRate.D(streamer.Len()).Round(time.Second)
	speaker.Unlock()

	ctrl := &beep.Ctrl{Paused: false, Streamer: beep.Seq(streamer, beep.Callback(func() {
		// when the track ends let the tui know so it can start a new one
		// keep the current one running so possible commands can still be entered
		globals.Audiostate <- globals.Metadata{
			Path:     file,
			Length:   length,
			Playtime: format.SampleRate.D(streamer.Position()).Round(time.Second),
			Finished: true}
	}))}

	speaker.Play(ctrl)

	// send initial metadata
	speaker.Lock()
	globals.Audiostate <- globals.Metadata{
		Path:     file,
		Length:   length,
		Playtime: time.Duration(0),
		Finished: false}
	speaker.Unlock()

	// event loop
	for {
		select {

		// when a command comes in, handlle it
		case command := <-globals.Speakercommand:
			switch command {
			case "pauze":
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
				break
			case "stop":
				speaker.Close()
				speakerevent <- "stopped"
				return
			}

		// resend metadata every second (for the timer)
		case <-time.After(time.Second):
			speaker.Lock()
			globals.Audiostate <- globals.Metadata{
				Path:     file,
				Length:   length,
				Playtime: format.SampleRate.D(streamer.Position()).Round(time.Second),
				Finished: false}
			speaker.Unlock()

		}
	}
}
