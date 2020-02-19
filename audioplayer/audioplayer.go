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

const (
	bufferSize time.Duration   = 200 * time.Millisecond
	gsr        beep.SampleRate = 48000 // the global sample rate
)

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

func Play(file globals.Track) (globals.Track, time.Duration) {
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

	return file, length
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

func GetPlaytime() (time.Duration, time.Duration, bool) {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl == nil {
		return time.Duration(0), time.Duration(0), false
	}

	speaker.Lock()
	var playtime = playingFile.Format.SampleRate.D(playingFile.Streamer.Position())
	var totaltime = playingFile.Length
	speaker.Unlock()

	return playtime, totaltime, true

}

// Initializes the speaker
func Init() {

	err := speaker.Init(gsr, gsr.N(bufferSize))
	if err != nil {
		log.Fatalf("failed to initialize audio device")
	}

}
