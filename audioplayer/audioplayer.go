package audioplayer

import (
	"log"
	"os"
	"sync"
	"time"

	"mp3bak2/globals"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const bufferSize = 100 * time.Millisecond

var (
	ctrl         *beep.Ctrl
	audioLock    = new(sync.Mutex)
	speakerevent = make(chan string)
	playingFile  audioFile
)

type audioFile struct {
	Path     string
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
}

func openFile(file string) audioFile {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening the file: %s", err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatalf("error decoding file: %s", err)
	}
	if ctrl != nil {
		playingFile.Streamer.Close()
	}

	if ctrl == nil || format.SampleRate != playingFile.Format.SampleRate {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(bufferSize))

		if err != nil {
			log.Fatalf("failed to initialize audio device: %s", err)
		}
	}

	// set the length of the track
	length := format.SampleRate.D(streamer.Len()).Round(time.Second)
	return audioFile{file, streamer, format, length}
}

func playFile(file audioFile) {

	audioLock.Lock()
	defer audioLock.Unlock()

	speaker.Lock()
	ctrl = &beep.Ctrl{Paused: false, Streamer: beep.Seq(file.Streamer, beep.Callback(func() {
		// when the track ends let the tui know so it can start a new one
		// keep the current one running so possible commands can still be entered
		globals.Audiostate <- globals.AudioStats{
			Path:     file.Path,
			Length:   file.Length,
			Playtime: file.Format.SampleRate.D(file.Streamer.Position()).Round(time.Second),
			Finished: true}
	}))}
	speaker.Unlock()

	speaker.Clear()
	speaker.Play(ctrl)

}

func pause() {
	if ctrl == nil {
		return
	}

	audioLock.Lock()
	defer audioLock.Unlock()

	speaker.Lock()
	ctrl.Paused = !ctrl.Paused
	speaker.Unlock()
}

// Controller : take control of the speaker
func Controller() {

	// event loop
	for {
		select {

		case file := <-globals.Playfile:
			playingFile = openFile(file.Path)
			playFile(playingFile)

		// when a command comes in, handlle it
		case command := <-globals.Speakercommand:
			switch command {
			case "pauze":
				pause()
			}

		// resend metadata every second (for the timer)
		case <-time.After(time.Second):
			if ctrl == nil {
				continue
			}
			speaker.Lock()
			globals.Audiostate <- globals.AudioStats{
				Path:     playingFile.Path,
				Length:   playingFile.Length,
				Playtime: playingFile.Format.SampleRate.D(playingFile.Streamer.Position()).Round(time.Second),
				Finished: false}
			speaker.Unlock()

		}
	}
}
