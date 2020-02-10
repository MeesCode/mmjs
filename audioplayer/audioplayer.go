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

const bufferSize = 100 * time.Millisecond

func stopPlayer() {
	globals.Speakercommand <- "stop"
	event := <-globals.Speakerevent
	for {
		if event != "stopped" {
			event = <-globals.Speakerevent
		}
		return
	}
}

func dummySpeaker() {
	command := <-globals.Speakercommand
	for {
		if command == "stop" {
			globals.Speakerevent <- "stopped"
			return
		}
		command = <-globals.Speakercommand
	}
}

func Init() {
	go dummySpeaker()
}

func Player(file string) {

	// stops the previous player
	stopPlayer()

	f, err := os.Open(file)

	if err != nil {
		println("hier gaat het mis")
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		println("nee hier")
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(bufferSize))
	if err != nil {
		log.Fatalf("failed to initialize audio device: %s", err)
	}

	speaker.Lock()
	length := format.SampleRate.D(streamer.Len()).Round(time.Second)
	speaker.Unlock()

	ctrl := &beep.Ctrl{Paused: false, Streamer: beep.Seq(streamer, beep.Callback(func() {
		globals.Audiostate <- globals.Metadata{
			Path:     file,
			Length:   length,
			Playtime: format.SampleRate.D(streamer.Position()).Round(time.Second),
			Finished: true}
	}))}

	speaker.Play(ctrl)

	speaker.Lock()
	globals.Audiostate <- globals.Metadata{
		Path:     file,
		Length:   length,
		Playtime: time.Duration(0),
		Finished: false}
	speaker.Unlock()

	for {
		select {
		case command := <-globals.Speakercommand:
			switch command {
			case "pauze":
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
				break
			case "stop":
				speaker.Close()
				globals.Speakerevent <- "stopped"
				return
			}
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
