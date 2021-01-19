// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"
	"unicode"

	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"

	"github.com/dhowden/tag"
	"github.com/gdamore/tcell/v2"
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
	trackID := -1
	for {

		// update the play timer every second
		<-time.After(time.Second)

		// update track info box when track id changed
		if audioplayer.IsPlaying() && audioplayer.GetPlaying().ID != trackID {
			myTui.app.QueueUpdateDraw(func() {
				updatePlayInfo()
				trackID = audioplayer.GetPlaying().ID
			})
		}

		// update the progress bar every second
		myTui.app.QueueUpdateDraw(func() {
			playtime, totaltime := audioplayer.GetPlaytime()
			drawprogressbar(playtime, totaltime)
		})

	}
}

// updatePlayInfo forces the interface to update. This is usefull when the application
// can be controlled from outside of the tui.
func updatePlayInfo() {
	updateInfoBox(audioplayer.GetPlaying(), myTui.infobox)
	drawplaylist()
}

// updateInfoBox updates one of the two information boxes with track information
func updateInfoBox(track globals.Track, box *tview.Table) {
	dir, name := path.Split(track.Path)
	box.SetCell(0, 1, tview.NewTableCell(tview.Escape(stringOrUnknown(track.Title))))
	box.SetCell(1, 1, tview.NewTableCell(tview.Escape(stringOrUnknown(track.Artist))))
	box.SetCell(2, 1, tview.NewTableCell(tview.Escape(stringOrUnknown(track.Album))))
	box.SetCell(3, 1, tview.NewTableCell(tview.Escape(stringOrUnknown(track.Genre))))
	if track.Year.Valid {
		box.SetCell(4, 1, tview.NewTableCell(strconv.FormatInt(track.Year.Int64, 10)))
	} else {
		box.SetCell(4, 1, tview.NewTableCell("unknown"))
	}
	box.SetCell(5, 1, tview.NewTableCell(tview.Escape(name)))
	box.SetCell(6, 1, tview.NewTableCell(tview.Escape(dir)))
}

// convert hex value encoded in an int to rgb notation
func hexToString(r int32) string {
	return "#" + fmt.Sprintf("%06s", (strconv.FormatInt(int64(r), 16)))
}

// drawplaylist draws the playlist. This function should be called after every
// function that alters this list.
func drawplaylist() {
	index := myTui.playlist.GetCurrentItem()
	myTui.playlist.Clear()
	for index, track := range audioplayer.Playlist {
		if audioplayer.Songindex == index {
			myTui.playlist.AddItem("["+hexToString(colorFocus.Hex())+"]▶[white] "+tview.Escape(trackToDisplayText(track)), "", 0, playsong)
		} else {
			myTui.playlist.AddItem("  "+tview.Escape(trackToDisplayText(track)), "", 0, playsong)
		}
	}
	itemCount := myTui.playlist.GetItemCount()
	if itemCount == 0 {
		return
	}
	if index >= itemCount {
		index = itemCount - 1
	}
	myTui.playlist.SetCurrentItem(index)
}

// drawfilelist draws the file list. This function should be called after every
// function that alters this list.
func drawfilelist() {
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(tview.Escape(trackToDisplayText(track)), "", 0, addsong)
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
			myTui.directorylist.AddItem(tview.Escape(path.Base(folder.Path)), "", 0, parentFunc)
		}

	}
}

// drawprogressbar draws the progressbar and timestamps.
// It will simply return whitout drawing if the total time of the track
// has a length of 0
func drawprogressbar(playtime time.Duration, length time.Duration) {
	myTui.progressbar.Clear()
	myTui.playtime.Clear()
	myTui.totaltime.Clear()

	// update the timestamps
	_, _, width, _ := myTui.progressbar.GetInnerRect()

	fill := 0
	if length > 0 {
		fill = width * int(playtime) / int(length)
	}

	for i := 0; i < fill-1; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneCkBoard)
	}
	fmt.Fprintf(myTui.progressbar, "%s%c%s", "["+hexToString(colorFocus.Hex())+"]", tcell.RuneBlock, "[white]")
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

	// update the progress bar
	ph, pm, ps := int64(playtime.Hours()), int64(playtime.Minutes()), int64(playtime.Seconds())
	fmt.Fprintf(myTui.playtime, "%02d:%02d:%02d", ph, pm-ph*60, ps-pm*60)

	th, tm, ts := int64(length.Hours()), int64(length.Minutes()), int64(length.Seconds())
	fmt.Fprintf(myTui.totaltime, "%02d:%02d:%02d", th, tm-th*60, ts-tm*60)

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

func playsong() {
	index := myTui.playlist.GetCurrentItem()
	audioplayer.PlaySong(myTui.playlist.GetCurrentItem())
	updatePlayInfo()
	myTui.playlist.SetCurrentItem(index)
}

func previoussong() {
	index := myTui.playlist.GetCurrentItem()
	audioplayer.Previoussong()
	updatePlayInfo()
	myTui.playlist.SetCurrentItem(index)
}

func nextsong() {
	index := myTui.playlist.GetCurrentItem()
	audioplayer.Nextsong()
	updatePlayInfo()
	myTui.playlist.SetCurrentItem(index)
}

func deletesong() {
	index := myTui.playlist.GetCurrentItem()
	audioplayer.Deletesong(index)
	updatePlayInfo()
	if index >= myTui.playlist.GetItemCount() {
		index = myTui.playlist.GetItemCount() - 1
	}
	myTui.playlist.SetCurrentItem(index)
}

func insertsong() {
	index := myTui.playlist.GetCurrentItem()
	filelistIndex := myTui.filelist.GetCurrentItem()
	if filelistIndex < len(filelistFiles)-1 {
		myTui.filelist.SetCurrentItem(filelistIndex + 1)
	}
	audioplayer.Insertsong(filelistFiles[filelistIndex])
	drawplaylist()
	if index >= myTui.playlist.GetItemCount() {
		index = myTui.playlist.GetItemCount() - 1
	}
	myTui.playlist.SetCurrentItem(index)
}

func addsong() {
	index := myTui.filelist.GetCurrentItem()
	if index < len(filelistFiles)-1 {
		myTui.filelist.SetCurrentItem(index + 1)
	}
	audioplayer.Addsong(filelistFiles[index])
	drawplaylist()
}

func moveUp() {
	index := myTui.playlist.GetCurrentItem()
	if index == 0 {
		return
	}
	audioplayer.MoveUp(index)
	drawplaylist()
	myTui.playlist.SetCurrentItem(index - 1)
}

func moveDown() {
	index := myTui.playlist.GetCurrentItem()
	audioplayer.MoveDown(index)
	drawplaylist()
	myTui.playlist.SetCurrentItem(index + 1)
}
