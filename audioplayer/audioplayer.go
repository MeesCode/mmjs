package audioplayer

import (
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

const bufferSize = 200 * time.Millisecond

var (
	ctrl        *beep.Ctrl
	audioLock   = new(sync.Mutex)
	playingFile audioFile
	gsr         beep.SampleRate = 48000 // the global sample rate
)

type audioFile struct {
	Track    globals.Track
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
	finished bool
}

func Play(file globals.Track) {
	audioLock.Lock()
	defer audioLock.Unlock()

	f, err := os.Open(file.Path)
	if err != nil {
		log.Fatalf("Error opening the file: %s", err)
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format

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
		log.Fatalf("error decoding file")
	}

	st := beep.Seq(beep.Resample(4, format.SampleRate, gsr, streamer))

	speaker.Lock()
	length := format.SampleRate.D(streamer.Len())
	ctrl = &beep.Ctrl{Paused: false, Streamer: st}
	playingFile = audioFile{file, streamer, format, length, false}
	speaker.Unlock()

	speaker.Clear()
	speaker.Play(ctrl)

	go sendState()
}

func Stop() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl != nil {
		playingFile.Streamer.Close()
	}
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

func sendState() {
	globals.Audiostate <- globals.AudioStats{
		Track:  playingFile.Track,
		Length: playingFile.Length,
	}
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

// Initializes the speaker
func Init() {

	err := speaker.Init(gsr, gsr.N(bufferSize))
	if err != nil {
		log.Fatalf("failed to initialize audio device")
	}

	// event loop
	for {
		<-time.After(time.Second / 2)
		sendDuration()
	}
}
