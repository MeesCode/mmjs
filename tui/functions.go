// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"database/sql"
	"fmt"
	"math/rand"
	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"
	"os"
	"path"
	"strconv"
	"time"
	"unicode"

	"github.com/dhowden/tag"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// trackToDisplayText takes a Track object and returns a string in thee
// Artist - Title format if available. If no artists is found it will
// omit this and if no track is found the filename will be used as title.
func trackToDisplayText(track globals.Track) string {
	var display = ""

	if track.Artist.Valid {
		display += track.Artist.String + " - "
	}

	if track.Title.Valid {
		display += track.Title.String
	} else {
		display += path.Base(track.Path)
	}

	return display
}

// stringOrUnknown takes a sql.NullString and return the sring if it is
// valid. Otherwise it will return "unknown"
func stringOrUnknown(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "unknown"
}

// audioStateUpdater is a function that should be ran as a goroutine.
// It will ask the audioplayer for the playing time of the current track.
// It will also start the next sond if the current song is finished.
func audioStateUpdater() {
	for {

		// update the play timer every half second
		<-time.After(time.Second / 2)
		// QueueUpdateDraw since this is performed outside the main thread
		myTui.app.QueueUpdateDraw(func() {
			playtime, totaltime, playing := audioplayer.GetPlaytime()
			if playing {
				drawprogressbar(playtime, totaltime)
				if playtime == totaltime {
					nextsong()
				}
			}
		})

	}
}

// updateInfoBox updates one of the two information boxes with track information
func updateInfoBox(track globals.Track, box *tview.Table) {
	dir, name := path.Split(track.Path)
	box.SetCell(0, 1, tview.NewTableCell(stringOrUnknown(track.Title)))
	box.SetCell(1, 1, tview.NewTableCell(stringOrUnknown(track.Artist)))
	box.SetCell(2, 1, tview.NewTableCell(stringOrUnknown(track.Album)))
	box.SetCell(3, 1, tview.NewTableCell(stringOrUnknown(track.Genre)))
	if track.Year.Valid {
		box.SetCell(4, 1, tview.NewTableCell(strconv.FormatInt(track.Year.Int64, 10)))
	} else {
		box.SetCell(4, 1, tview.NewTableCell("unknown"))
	}
	box.SetCell(5, 1, tview.NewTableCell(name))
	box.SetCell(6, 1, tview.NewTableCell(dir))
}

// drawplaylist draws the playlist. This function should be called after every
// function that alters this list.
func drawplaylist() {
	myTui.playlist.Clear()
	for index, track := range playlistFiles {
		if songindex == index {
			myTui.playlist.AddItem(trackToDisplayText(track), "", '>', playsong)
		} else {
			myTui.playlist.AddItem(trackToDisplayText(track), "", 0, playsong)
		}
	}
	myTui.playlist.SetCurrentItem(songindex)
}

// drawfilelist draws the file list. This function should be called after every
// function that alters this list.
func drawfilelist() {
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(trackToDisplayText(track), "", 0, addsong)
	}
}

// drawdirectorylist draws the directory list. This function should be called after every
// function that alters this list.
func drawdirectorylist(parentFunc func(), isRoot bool) {
	myTui.directorylist.Clear()
	for index, folder := range directorylistFolders {

		// first folder shows .. instead of the folder name
		// (unless we are in the root directory)
		if index == 0 && !isRoot {
			myTui.directorylist.AddItem("..", "", 0, parentFunc)
		} else {
			myTui.directorylist.AddItem(path.Base(folder.Path), "", 0, parentFunc)
		}

	}
}

// startTrack starts a given track and updates the ui accordingly
func startTrack(t globals.Track) {
	_, length := audioplayer.Play(t)
	drawplaylist()
	updateInfoBox(t, myTui.infobox)
	drawprogressbar(time.Duration(0), length)
}

// playsong plays the song currently selected track on the playlist
func playsong() {
	if len(playlistFiles) == 0 || songindex > len(playlistFiles) {
		return
	}
	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	startTrack(playlistFiles[myTui.playlist.GetCurrentItem()])
}

// nextsong plays the next song (if available)
func nextsong() {
	if len(playlistFiles) > songindex+1 {
		songindex++
		drawplaylist()
		startTrack(playlistFiles[songindex])
	}
}

// previoussong plays the previous song (if available)
func previoussong() {
	if songindex > 0 {
		songindex--
		drawplaylist()
		startTrack(playlistFiles[songindex])
	}
}

// addsong adds a song to the playlist
func addsong() {
	track := filelistFiles[myTui.filelist.GetCurrentItem()]
	playlistFiles = append(playlistFiles, track)
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// insertsong inserts a song into the playlist directly after the song that
// is currently playing.
func insertsong() {
	track := filelistFiles[myTui.filelist.GetCurrentItem()]
	playlistFiles = append(playlistFiles[:songindex+1], append([]globals.Track{track}, playlistFiles[songindex+1:]...)...)
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// deletesong removes the currently selected song from the playlist.
func deletesong() {

	// if list is empty do nothing
	if len(playlistFiles) == 0 {
		return
	}

	// remove selected song from the list
	var i = myTui.playlist.GetCurrentItem()
	playlistFiles = append(playlistFiles[:i], playlistFiles[i+1:]...)

	// if after deleting an item the list is empty stop playback
	if len(playlistFiles) == 0 {
		audioplayer.Stop()
		drawplaylist()
		return
	}

	// stop the music when last song is deleted
	if len(playlistFiles) == songindex && i == songindex {
		audioplayer.Stop()
		songindex--
		drawplaylist()
		return
	}

	// play the next song when the current song is deleted
	// but there is a next song on the list
	if i == songindex {
		startTrack(playlistFiles[songindex])
	}

	// if we delete a song that is before the current one
	// match the songindex to the new list
	if i < songindex {
		songindex--
	}

	drawplaylist()
	myTui.playlist.SetCurrentItem(i)
}

// drawprogressbar draws the progressbar and timestamps.
// It will simply return whitout drawing if the total time of the track
// has a length of 0
func drawprogressbar(playtime time.Duration, length time.Duration) {
	if length == 0 {
		return
	}

	myTui.progressbar.Clear()
	myTui.playtime.Clear()
	myTui.totaltime.Clear()

	// update the timestamps
	_, _, width, _ := myTui.progressbar.GetInnerRect()
	fill := width * int(playtime) / int(length)
	for i := 0; i < fill-1; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneCkBoard)
	}
	fmt.Fprintf(myTui.progressbar, "%s%c%s", "[crimson]", tcell.RuneBlock, "[white]")
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

	// update the progress bar
	ph, pm, ps := int64(playtime.Hours()), int64(playtime.Minutes()), int64(playtime.Seconds())
	fmt.Fprintf(myTui.playtime, "%02d:%02d:%02d", ph, pm-ph*60, ps-pm*60)

	th, tm, ts := int64(length.Hours()), int64(length.Minutes()), int64(length.Seconds())
	fmt.Fprintf(myTui.totaltime, "%02d:%02d:%02d", th, tm-th*60, ts-tm*60)

}

// shuffle shuffles the playlist and places the currently playing track as the first
// track in the playlist. It will not halt playback.
func shuffle() {
	if len(playlistFiles) == 0 {
		return
	}

	// remove current song from list
	var cursong = playlistFiles[songindex]
	playlistFiles = append(playlistFiles[:songindex], playlistFiles[songindex+1:]...)

	// shuffle the list
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(playlistFiles), func(i, j int) {
		playlistFiles[i], playlistFiles[j] = playlistFiles[j], playlistFiles[i]
	})

	// prepend current song to the list
	playlistFiles = append([]globals.Track{cursong}, playlistFiles...)
	songindex = 0

	drawplaylist()
}

// openSearch removes the keybinds box and replaces it with the search box.
func openSearch() {
	if myTui.searchinput.HasFocus() || myTui.playlistinput.HasFocus() {
		return
	}
	myTui.mainFlex.RemoveItem(myTui.keybinds)
	myTui.mainFlex.AddItem(myTui.searchinput, 3, 0, false)
	myTui.searchinput.SetText("")
	focusWithColor(myTui.searchinput)
}

// closeSearch removes the search box and replaces it with the keybinds box.
func closeSearch() {
	myTui.filelist.SetTitle(" Search results ")
	myTui.mainFlex.RemoveItem(myTui.searchinput)
	myTui.mainFlex.AddItem(myTui.keybinds, 3, 0, false)
	focusWithColor(myTui.filelist)
	drawfilelist()
}

// clear removes all entries from the playlist and stops playback.
func clear() {
	audioplayer.Stop()
	songindex = 0
	playlistFiles = nil
	drawplaylist()
}

// jump to a new element in the list depending on the key pressed.
func jump(r rune) {
	for index, folder := range directorylistFolders {
		if unicode.ToLower(rune(path.Base(folder.Path)[0])) == unicode.ToLower(r) {
			myTui.directorylist.SetCurrentItem(index)
			return
		}
	}
}

// goback selects the top item in the directory list and enters it.
func goback() {
	// BUG(mees): when in the root folder we enter the top folder instead of going to
	// the parent folder.
	myTui.directorylist.SetCurrentItem(0)
	changedir()
}

// parseTrack takes a path to a playable file, extracts the metadata and returns a file
// object containing this metadata. The metadata might not be found and defaulted to nil.
func parseTrack(file string) globals.Track {
	f, _ := os.Open(path.Join(globals.Root, file))
	m, err := tag.ReadFrom(f)

	var track globals.Track

	// relative path
	rpath := path.Clean(file[len(globals.Root):])

	_, filename := path.Split(file)

	// if no tags were found default to nil
	if err != nil {
		track = globals.Track{
			ID:       -1,
			Path:     rpath,
			FolderID: -1,
			Title:    sql.NullString{String: filename, Valid: true},
			Album:    sql.NullString{String: "", Valid: false},
			Artist:   sql.NullString{String: "", Valid: false},
			Genre:    sql.NullString{String: "", Valid: false},
			Year:     sql.NullInt64{Int64: -1, Valid: false}}
	} else {
		track = globals.Track{
			ID:       -1,
			Path:     rpath,
			FolderID: -1,
			Title:    database.StringToSQLNullableString(m.Title()),
			Artist:   database.StringToSQLNullableString(m.Artist()),
			Album:    database.StringToSQLNullableString(m.Album()),
			Genre:    database.StringToSQLNullableString(m.Genre()),
			Year:     database.IntToSQLNullableInt(m.Year())}
	}

	return track
}

// moveUp swaps the currently selected track in the playlist with the one above it.
func moveUp() {
	selected := myTui.playlist.GetCurrentItem()

	if selected == 0 {
		return
	}

	if selected == songindex {
		songindex--
	} else if selected == songindex+1 {
		songindex++
	}

	playlistFiles[selected], playlistFiles[selected-1] = playlistFiles[selected-1], playlistFiles[selected]
	drawplaylist()
	myTui.playlist.SetCurrentItem(selected - 1)
}

// moveDown swaps the currently selected track in the playlist with the one below it.
func moveDown() {
	selected := myTui.playlist.GetCurrentItem()

	if selected+1 == len(playlistFiles) {
		return
	}

	if selected == songindex {
		songindex++
	} else if selected == songindex-1 {
		songindex--
	}

	playlistFiles[selected], playlistFiles[selected+1] = playlistFiles[selected+1], playlistFiles[selected]
	drawplaylist()
	myTui.playlist.SetCurrentItem(selected + 1)
}

func focusWithColor(primitive tview.Primitive) {
	myTui.directorylist.SetBorderColor(colorUnfocus)
	myTui.filelist.SetBorderColor(colorUnfocus)
	myTui.playlist.SetBorderColor(colorUnfocus)
	myTui.searchinput.SetBorderColor(colorUnfocus)
	myTui.playlistinput.SetBorderColor(colorUnfocus)

	list, ok := primitive.(*tview.List)
	if ok {
		list.SetBorderColor(colorFocus)
	}

	inputfield, ok := primitive.(*tview.InputField)
	if ok {
		inputfield.SetBorderColor(colorFocus)
	}

	myTui.app.SetFocus(primitive)
}
