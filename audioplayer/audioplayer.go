// Package audioplayer controls the audio.
package audioplayer

import (
	"log"
	"time"
	"path"

	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"

	vlc "github.com/adrg/libvlc-go/v3"
)

var (
	player  *vlc.Player
	manager *vlc.EventManager
	quit    chan struct{}
	eventID vlc.EventID
)

// player is always playing, this function never returns
// and only runs the first time it is called because only 
// then it has something to play
func play() {

    err := player.Play()
    if err != nil {
        Playlist[Songindex].Error = true
        log.Println(err)
		Nextsong()
    }

    <-quit
}

// SetTrack changes the track that is playing
func SetTrack(){
	media, err := player.LoadMediaFromPath(path.Clean(path.Join(globals.Root, Playlist[Songindex].Path)))
	if err != nil {
		Playlist[Songindex].Error = true
		log.Println(err)
		defer Nextsong()
		return
	}
	defer media.Release()

	if !player.IsPlaying(){
		play()
	}
}

func IsLoaded() bool {
	return player.WillPlay()
}

// Stop stops the playback of the currently playing track (if any).
func Stop() {
	player.Stop()
}

// Close closes the audio engine
func Close() {
	close(quit)
	manager.Detach(eventID)
	player.Stop()
    player.Release()
}


// Pause the currently playing track (if any).
func Pause() {
	player.SetPause(true)
}

// Resume the currently playing track (if any).
func Resume() {
	player.SetPause(false)
}

// GetPlaytime returns the play time, and the total time of the track.
// If no track is playing the returned timings will be zero.
func GetPlaytime() (time.Duration, time.Duration) {
	t, _ := player.MediaLength()
	total := time.Duration(t) * time.Millisecond
	c, _ := player.MediaTime()
	current := time.Duration(c) * time.Millisecond
	return current, total
}

// GetPlaying returns track that is either loaded or being loaded.
// in transition, new file will be returned
func GetPlaying() globals.Track {
	if len(Playlist) == 0 {
		return globals.Track{}
	}
	return Playlist[Songindex]
}

// IsPlaying returns true when a file is loaded and playing
func IsPlaying() bool {
	return player.IsPlaying()
}

// IsPaused returns true when a file is loaded and paused
func IsPaused() bool {
	return !player.IsPlaying()
}

// PlayPause is a shorthand, it will play if paused (and a song is loaded)
// it will pause if playing (and a song is loaded)
func PlayPause(){
	player.TogglePause()
}

// wait for a signal that the track has finished playing.
// automatically play the next song
func finishTrack() {
	// if in database mode, add one to the play counter
	if globals.Config.Mode == "database" {
		database.IncrementPlayCounter(GetPlaying().ID)
	}

	Nextsong()
}

// Initialize the speaker with the specification defined at the top.
func Initialize() {
	var err error = nil

	if err = vlc.Init("--no-video", "--quiet"); err != nil {
        log.Fatal(err)
    }

	player, err = vlc.NewPlayer()
    if err != nil {
        log.Fatal(err)
    }

	manager, err = player.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	quit = make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		finishTrack()
	}

	eventID, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal(err)
	}
}
