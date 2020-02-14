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
	ctrl        *beep.Ctrl
	audioLock   = new(sync.Mutex)
	playingFile audioFile
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
	speaker.Lock()
	length := format.SampleRate.D(streamer.Len())
	speaker.Unlock()
	return audioFile{file, streamer, format, length, false}
}

func playFile(file audioFile) {
	speaker.Lock()
	ctrl = &beep.Ctrl{Paused: false, Streamer: beep.Seq(file.Streamer)}
	speaker.Unlock()

	globals.Audiostate <- globals.AudioStats{
		Track:  file.Track,
		Length: file.Length,
	}

	speaker.Clear()

	var speakerDone = make(chan bool)
	speaker.Play(ctrl, beep.Callback(func() {
		speakerDone <- true
	}))
	<-speakerDone

}

func Stop() {
	audioLock.Lock()
	defer audioLock.Unlock()

	speaker.Clear()
}

func Pause() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl == nil {
		return
	}

	speaker.Lock()
	ctrl.Paused = !ctrl.Paused
	speaker.Unlock()
}

func Play(file globals.Track) {
	audioLock.Lock()
	defer audioLock.Unlock()

	playingFile = openFile(file)
	playFile(playingFile)
}

func sendDuration() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl == nil {
		return
	}

	speaker.Lock()
	globals.DurationState <- globals.DurationStats{
		Playtime: playingFile.Format.SampleRate.D(playingFile.Streamer.Position()),
		Length:   playingFile.Length,
	}
	speaker.Unlock()
}

// Initializes the duration timer
func Init() {

	// event loop
	for {
		<-time.After(time.Second / 2)
		sendDuration()
	}
}
