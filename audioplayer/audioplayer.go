package audioplayer

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"mp3bak2/globals"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

const bufferSize = 100 * time.Millisecond

var (
	speakerDone  = make(chan bool)
	ctrl         *beep.Ctrl
	audioLock    = new(sync.Mutex)
	speakerevent = make(chan string)
	playingFile  audioFile
)

type audioFile struct {
	Track    globals.Track
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
	finished bool
}

func openFile(file globals.Track) audioFile {
	f, err := os.Open(file.Path)
	if err != nil {
		log.Fatalf("Error opening the file: %s", err)
	}

	var (
		streamer beep.StreamSeekCloser
		format   beep.Format
	)

	switch strings.ToLower(path.Ext(file.Path)) {
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	default:
		err = fmt.Errorf("unsupported file format")
	}

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
	return audioFile{file, streamer, format, length, false}
}

func playFile(file audioFile) {

	audioLock.Lock()
	defer audioLock.Unlock()

	speaker.Lock()
	ctrl = &beep.Ctrl{Paused: false, Streamer: beep.Seq(file.Streamer)}
	speaker.Unlock()

	globals.Audiostate <- globals.AudioStats{
		Track:  file.Track,
		Length: file.Length,
	}

	speaker.Clear()
	speaker.Play(ctrl, beep.Callback(func() {
		speakerDone <- true
	}))
	<-speakerDone

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
			playingFile = openFile(file)
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
			globals.DurationState <- globals.DurationStats{
				Playtime: playingFile.Format.SampleRate.D(playingFile.Streamer.Position()).Round(time.Second),
				Length:   playingFile.Length,
			}
			speaker.Unlock()

		}
	}
}
