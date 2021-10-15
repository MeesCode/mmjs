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
	ctrl          *beep.Ctrl
	audioLock     = new(sync.Mutex)
	playingFile   audioFile
	done          chan bool
	prevSongindex = -1
)

// a struct that holds information about the currently playing track.
type audioFile struct {
	Track    globals.Track
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Length   time.Duration
}

func forcePlayOnNextPlay() {
	prevSongindex = -1
}

// Play stops playback of the currently playing song (if any) and start the playback
// of the current songindex. It will open, decode resample and play the file in that order.
// when calling play without changing the songindex nothing will happen
func play() {
	audioLock.Lock()
	defer audioLock.Unlock()

	// when scrolling with f12 (which is something people do) ignore
	// when the songindex didn't change. This has to do with multiple threads
	// calling this function and having to wait for every file to finish decoding
	// this makes scrolling very slow. This might not be the best way to do this, but i
	// can't think of another.
	if prevSongindex == Songindex {
		return
	}

	prevSongindex = Songindex

	file := GetPlaying()

	if file.Error {
		log.Println("Previously detected error, skipping file", file.Path)
		Nextsong()
		return
	}

	filePath := path.Join(globals.Root, file.Path)

	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening the file", err)
		Playlist[Songindex].Error = true
		Nextsong()
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
		Playlist[Songindex].Error = true
		Nextsong()
		return
	}

	if err != nil {
		log.Println("Error decoding file", err, file.Path)

		if !globals.Config.AudioConvert {
			Playlist[Songindex].Error = true
			Nextsong()
			return
		}

		// conversion failed
		log.Println("Try to convert file")
		if !convertFile(Playlist[Songindex]) {
			log.Println("File conversion failed")
			Playlist[Songindex].Error = true
			Nextsong()
			return
		}

		// conversion succeeded
		f, err := os.Open(filePath + ".flac")
		if err != nil {
			log.Println("Error opening the file, even after conversion", err)
			Playlist[Songindex].Error = true
			Nextsong()
			return
		}

		streamer, format, err = flac.Decode(f)

		// decoding failed still, just give up at this point
		if err != nil {
			log.Println("Error decoding file, cannot convert", err, file.Path)
			Playlist[Songindex].Error = true
			Nextsong()
			return
		}

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

	if !IsLoaded() {
		return
	}

	speaker.Lock()
	ctrl.Paused = true
	speaker.Unlock()
}

// Resume the currently playing track (if any).
func Resume() {
	audioLock.Lock()
	defer audioLock.Unlock()

	if !IsLoaded() {
		return
	}

	speaker.Lock()
	ctrl.Paused = false
	speaker.Unlock()
}

// GetPlaytime returns the play time, and the total time of the track.
// If no track is playing the returned timings will be zero.
func GetPlaytime() (time.Duration, time.Duration) {
	// audioLock.Lock()
	// defer audioLock.Unlock()

	if !IsLoaded() {
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

// IsLoaded returns true when a file is loaded, either playing or paused
func IsLoaded() bool {
	if ctrl == nil {
		return false
	}
	return true
}

// IsPlaying returns true when a file is loaded and playing
func IsPlaying() bool {
	if ctrl == nil {
		return false
	}
	return !ctrl.Paused
}

// IsPaused returns true when a file is loaded and paused
func IsPaused() bool {
	if ctrl == nil {
		return false
	}
	return ctrl.Paused
}

// PlayPause is a shorthand, it will play if paused (and a song is loaded)
// it will pause if playing (and a song is loaded)
func PlayPause(){
	if IsPaused(){
		Resume()
	} else {
		Pause()
	}
}

// wait for a signal that the track has finished playing.
// automatically play the next song
func waitForNext() {
	for {
		<-done

		// if in database mode, add one to the play counter
		if globals.Config.Mode == "database" {
			database.IncrementPlayCounter(GetPlaying().ID)
		}

		// always start a new song when the previous is finished
		forcePlayOnNextPlay()
		Nextsong()
	}
}

// Initialize the speaker with the specification defined at the top.
func Initialize() {
	err := speaker.Init(gsr, gsr.N(bufferSize))
	if err != nil {
		log.Fatalln("failed to initialize audio device", err)
	}
	done = make(chan bool)
	go waitForNext()
}
