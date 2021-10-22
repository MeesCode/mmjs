// Package audioplayer the audio.
package audioplayer

import (
	"math/rand"
	"time"

	"github.com/MeesCode/mmjs/globals"
)

// global variables
var (
	Playlist    = make([]globals.Track, 0)
	Songindex   = 0 // the index of the currently playing track
	TogglePause func() error
	Stop        func() error
	SetPause    func(bool) error
)

// PlaySong plays the song at the index of the playlist
func PlaySong(index int) {
	if len(Playlist) == 0 || Songindex > len(Playlist) {
		return
	}
	Songindex = index
	go updateTrack()
}

// Nextsong plays the next song (if available)
func Nextsong() {
	if len(Playlist) > Songindex+1 {
		Songindex++
		go updateTrack()
	}
}

// Previoussong plays the previous song (if available)
func Previoussong() {
	if Songindex > 0 {
		Songindex--
		go updateTrack()
	}
}

// Addsong adds a song to the playlist
func Addsong(track globals.Track) {
	Playlist = append(Playlist, track)
}

// Deletesong removes the currently selected song from the playlist.
func Deletesong(index int) {

	// if list is empty do nothing
	if len(Playlist) == 0 {
		return
	}

	// remove selected song from the list
	Playlist = append(Playlist[:index], Playlist[index+1:]...)

	// if after deleting an item the list is empty stop playback
	if len(Playlist) == 0 {
		Stop()
		return
	}

	// stop the music when last song is deleted
	if len(Playlist) == Songindex && index == Songindex {
		Stop()
		Songindex--
		return
	}

	// play the next song when the current song is deleted
	// but there is a next song on the list
	if index == Songindex {
		go updateTrack()
	}

	// if we delete a song that is before the current one
	// match the Songindex to the new list
	if index < Songindex {
		Songindex--
	}

}

// Insertsong inserts a song into the playlist directly after the song that
// is currently playing.
func Insertsong(track globals.Track) {
	Playlist = append(Playlist[:Songindex+1], append([]globals.Track{track}, Playlist[Songindex+1:]...)...)
}

// Shuffle shuffles the playlist and places the currently playing track as the first
// track in the playlist. It will not halt playback.
func Shuffle() {
	if len(Playlist) == 0 || Songindex > len(Playlist) {
		return
	}

	// remove current song from list
	var cursong = Playlist[Songindex]
	Playlist = append(Playlist[:Songindex], Playlist[Songindex+1:]...)

	// shuffle the list
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(Playlist), func(i, j int) {
		Playlist[i], Playlist[j] = Playlist[j], Playlist[i]
	})

	// prepend current song to the list
	Playlist = append([]globals.Track{cursong}, Playlist...)
	Songindex = 0
}

// Clear removes all entries from the playlist and stops playback.
func Clear() {
	Songindex = 0
	Playlist = make([]globals.Track, 0)
	Stop()
}

// MoveUp swaps the currently selected track in the playlist with the one above it.
func MoveUp(index int) {
	if index == 0 {
		return
	}

	if index == Songindex {
		Songindex--
	} else if index == Songindex+1 {
		Songindex++
	}

	Playlist[index], Playlist[index-1] = Playlist[index-1], Playlist[index]
}

// MoveDown swaps the currently selected track in the playlist with the one below it.
func MoveDown(index int) {
	if index+1 == len(Playlist) {
		return
	}

	if index == Songindex {
		Songindex++
	} else if index == Songindex-1 {
		Songindex--
	}

	Playlist[index], Playlist[index+1] = Playlist[index+1], Playlist[index]
}
