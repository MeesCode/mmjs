// Package audioplayer controls the audio.
package audioplayer

import (
	"log"
	"path"
	"sync"
	"time"

	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"

	vlc "github.com/adrg/libvlc-go/v3"
)

var (
	player    *vlc.Player
	manager   *vlc.EventManager
	eventID   vlc.EventID
	audioLock = new(sync.Mutex)
	WillPlay  func() bool
	IsPlaying func() bool
)

// updateTrack changes the track that is playing
func updateTrack() {

	audioLock.Lock()
	defer audioLock.Unlock()

	// race condition after deleting songs to quickly
	if Songindex < len(Playlist){
		return
	}

	media, err := player.LoadMediaFromPath(path.Join(globals.Root, Playlist[Songindex].Path))
	if err != nil {
		Playlist[Songindex].Error = true
		log.Println("Error media file could not be opened", err)
		defer Nextsong()
		return
	}

	defer media.Release()

	err = player.Play()
	if err != nil {
		Playlist[Songindex].Error = true
		log.Println("Error media file could not be played", err)
		defer Nextsong()
		return
	}

	log.Println(Songindex)

	Playlist[Songindex].Error = false
}

// Close closes the audio engine
func Close() {
	manager.Detach(eventID)
	player.Stop()
	player.Release()
	vlc.Release()
}

// GetPlaytime returns the play time, and the total time of the track.
// If no track is playing the returned timings will be zero.
func GetPlaytime() (time.Duration, time.Duration) {
	t, _ := player.MediaLength()
	total := time.Duration(t) * time.Millisecond
	c, _ := player.MediaTime()
	current := time.Duration(c) * time.Millisecond

	if current < 0 { current = 0 }
	if total < 0   { total = 0   }

	return current, total
}

// GetPlaying returns that is currently selected in the playlist
// if there is no such track second returned value is false
func GetPlaying() globals.Track {
	if len(Playlist) == 0 {
		return globals.Track{}
	}
	return Playlist[Songindex]
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

	// expose player functions directly
	IsPlaying = player.IsPlaying
	WillPlay = player.WillPlay
	TogglePause = player.TogglePause
	Stop = player.Stop
	SetPause = player.SetPause

	manager, err = player.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	eventCallback := func(event vlc.Event, userData interface{}) {
		finishTrack()
	}

	eventID, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// SetMediaPosition sets media position as percentage between 0.0 and 1.0.
// Some formats and protocols do not support this.
func SetMediaPosition(percentage float32) {
	log.Println(player.IsSeekable());
	if(!player.IsSeekable()) {
		log.Println("Song is not seekable");
		return;
	}
	player.SetMediaPosition(percentage);
}

// GetMediaPosition returns media position as a
// float percentage between 0.0 and 1.0.
func GetMediaPosition() (float32, error) {
	return player.MediaPosition();
}
