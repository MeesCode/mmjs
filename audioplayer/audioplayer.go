// Package audioplayer controls the audio.
package audioplayer

import (
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"github.com/h2non/filetype"
)

// these values can be ajusted to improve playback.
// For a more detailed explanation check the documentation of
// the beep package.
const (
	bufferSize time.Duration   = 100 * time.Millisecond
	gsr        beep.SampleRate = 48000 // the global sample rate
	quality    int             = 3
)

var (
	ctrl        *beep.Ctrl
	audioLock   = new(sync.Mutex)
	playingFile audioFile
	done        chan bool
	mode        string
)

// a struct that holds information about the currently playing track.
type audioFile struct {
	Track    globals.Track
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
}

// Play stops playback of the currently playing song (if any) and start the playback
// of the provided song. It will open, decode resample and play the file in that order.
func Play(file globals.Track) {
	audioLock.Lock()
	defer audioLock.Unlock()

	filePath := path.Join(globals.Root, file.Path)

	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening the file", err)
		speaker.Clear()
		return
	}

	// read the max file header and reset to begin
	head := make([]byte, 261)
	f.Read(head)
	f.Seek(0, io.SeekStart)

	// determine filetype by header
	kind, _ := filetype.Match(head)

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch kind.MIME.Value {
	case "audio/x-wav":
		streamer, format, err = wav.Decode(f)
	case "audio/mpeg":
		streamer, format, err = mp3.Decode(f)
	case "audio/ogg":
		streamer, format, err = vorbis.Decode(f)
	case "audio/x-flac":
		streamer, format, err = flac.Decode(f)
	default:
		log.Println("filetype not supported")
		speaker.Clear()
		return
	}

	if err != nil {
		log.Println("error decoding file", err)
		speaker.Clear()
		return
	}

	st := beep.Seq(beep.Resample(quality, format.SampleRate, gsr, streamer))

	speaker.Lock()
	length := format.SampleRate.D(streamer.Len())
	ctrl = &beep.Ctrl{Paused: false, Streamer: st}
	playingFile = audioFile{file, streamer, format, length}
	speaker.Unlock()

	speaker.Clear()
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		done <- true
	})))
}

// Stop stops the playback of the currently playing track (if any).
func Stop() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if ctrl != nil {
		playingFile.Streamer.Close()
	}

	speaker.Lock()
	ctrl = nil
	speaker.Unlock()
	speaker.Clear()
}

// Pause the currently playing track (if any).
func Pause() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if !IsPlaying() {
		return
	}

	speaker.Lock()
	ctrl.Paused = !ctrl.Paused
	speaker.Unlock()
}

// GetPlaytime returns the play time, and the total time of the track.
// If no track is playing the returned boolean will be false and the
// timings will be zero.
func GetPlaytime() (time.Duration, time.Duration) {
	// audioLock.Lock()
	// defer audioLock.Unlock()

	if !IsPlaying() {
		return time.Duration(0), time.Duration(0)
	}

	speaker.Lock()
	var playtime = playingFile.Format.SampleRate.D(playingFile.Streamer.Position())
	var totaltime = playingFile.Length
	speaker.Unlock()

	return playtime, totaltime

}

// GetPlaying returns track that is either loaded or being loaded.
// in transition, new file will be returned
func GetPlaying() globals.Track {
	if len(Playlist) == 0 {
		return globals.Track{}
	}
	return Playlist[Songindex]
}

// IsPlaying returns true when a file is loaded, either playing or paused
func IsPlaying() bool {
	if ctrl == nil {
		return false
	}
	return true
}

// wait for a signal that the track has finished playing.
// automatically play the next song
func waitForNext() {
	for {
		<-done

		// if in database mode, add one to the play counter
		if mode == "database" {
			database.IncrementPlayCounter(GetPlaying())
		}

		Nextsong()
	}
}

// Initialize the speaker with the specification defined at the top.
func Initialize(m string) {
	err := speaker.Init(gsr, gsr.N(bufferSize))
	if err != nil {
		log.Fatalln("failed to initialize audio device", err)
	}
	mode = m
	done = make(chan bool)
	go waitForNext()
}
