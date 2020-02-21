// Package audioplayer controls the audio.
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

// these values can be ajusted to improve playback.
// For a more detailed explanation check the documentation of
// the beep package.
const (
	bufferSize time.Duration   = 200 * time.Millisecond
	gsr        beep.SampleRate = 48000 // the global sample rate
	quality    int             = 4
)

var (
	ctrl        *beep.Ctrl
	audioLock   = new(sync.Mutex)
	playingFile audioFile
)

// a struct that holds information about the currently playing track.
type audioFile struct {
	Track    globals.Track
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
	finished bool
}

// Play stops playback of the currently playing song (if any) and start the playback
// of the provided song. It will open, decode resample and play the file in that order.
func Play(file globals.Track) (globals.Track, time.Duration) {
	audioLock.Lock()
	defer audioLock.Unlock()

	f, err := os.Open(path.Join(globals.Root, file.Path))
	if err != nil {
		log.Println("Error opening the file", err)
		speaker.Clear()
		return file, time.Duration(0)
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
		log.Println("error decoding file", err)
		speaker.Clear()
		return file, time.Duration(0)
	}

	st := beep.Seq(beep.Resample(quality, format.SampleRate, gsr, streamer))

	speaker.Lock()
	length := format.SampleRate.D(streamer.Len())
	ctrl = &beep.Ctrl{Paused: false, Streamer: st}
	playingFile = audioFile{file, streamer, format, length, false}
	speaker.Unlock()

	speaker.Clear()
	speaker.Play(ctrl)

	return file, length
}

// Stop stops the playback of the currently playing track (if any).
func Stop() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl != nil {
		playingFile.Streamer.Close()
	}
	speaker.Clear()
}

// Pause the currently playing track (if any).
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

// GetPlaytime returns the play time, and the total time of the track.
// If no track is playing the returned boolean will be false and the
// timings will be zero.
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

// Initialize the speaker with the specification defined at the top.
func Initialize() {
	err := speaker.Init(gsr, gsr.N(bufferSize))
	if err != nil {
		log.Fatalln("failed to initialize audio device", err)
	}
}
