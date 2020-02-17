package tui

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"mp3bak2/audioplayer"
	"mp3bak2/database"
	"mp3bak2/globals"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dhowden/tag"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

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

func stringOrUnknown(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "unknown"
}

func audioStateUpdater() {
	for {
		select {
		case data := <-globals.Audiostate:
			myTui.app.QueueUpdate(func() {
				updateInfoBox(data.Track, myTui.infobox)
				drawprogressbar(time.Duration(0), data.Length)
			})

		case data := <-globals.DurationState:
			myTui.app.QueueUpdate(func() {
				drawprogressbar(data.Playtime, data.Length)
				if data.Playtime == data.Length {
					myTui.app.QueueUpdate(nextsong)
				}
			})

		}
	}
}

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

// draw the playlist
func drawplaylist() {
	myTui.playlist.Clear()
	for index, track := range playlistFiles {
		if songindex == index {
			myTui.playlist.AddItem(trackToDisplayText(track), "", '>', func() { myTui.app.QueueUpdate(playsong) })
		} else {
			myTui.playlist.AddItem(trackToDisplayText(track), "", 0, func() { myTui.app.QueueUpdate(playsong) })
		}
	}
	myTui.playlist.SetCurrentItem(songindex)
	myTui.app.Draw()
}

// draw the file list
func drawfilelist() {
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(trackToDisplayText(track), "", 0, func() { myTui.app.QueueUpdate(addsong) })
	}
	myTui.app.Draw()
}

// draw the directory list
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
	myTui.app.Draw()
}

// play the song currently selected on the playlist
func playsong() {
	if len(playlistFiles) == 0 || songindex > len(playlistFiles) {
		return
	}
	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	go audioplayer.Play(playlistFiles[myTui.playlist.GetCurrentItem()])
}

// go to the next song (if available)
func nextsong() {
	if len(playlistFiles) > songindex+1 {
		songindex++
		drawplaylist()
		go audioplayer.Play(playlistFiles[songindex])
	}
}

// go to the previous song (if available)
func previoussong() {
	if songindex > 0 {
		songindex--
		drawplaylist()
		go audioplayer.Play(playlistFiles[songindex])
	}
}

// add a song to the playlist
func addsong() {
	track := filelistFiles[myTui.filelist.GetCurrentItem()]
	playlistFiles = append(playlistFiles, track)
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// insert a song into the playlist
func insertsong() {
	track := filelistFiles[myTui.filelist.GetCurrentItem()]
	playlistFiles = append(playlistFiles[:songindex+1], append([]globals.Track{track}, playlistFiles[songindex+1:]...)...)
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// insert a song into the playlist
func deletesong() {

	// if list is empty do nothing
	if len(playlistFiles) == 0 {
		return
	}

	// remove selected song from the list
	var i = myTui.playlist.GetCurrentItem()
	playlistFiles = append(playlistFiles[:i], playlistFiles[i+1:]...)

	// if after deleting an item the list is empty make sure the
	// songindex is set to 0 and redraw
	if len(playlistFiles) == 0 {
		songindex = 0
		drawplaylist()
		return
	}

	// stop the music when last song is deleted
	if len(playlistFiles) == songindex && i == songindex {
		go audioplayer.Stop()
		songindex--
		drawplaylist()
		return
	}

	// play the next song when the current song is deleted
	// but there is a next song on the list
	if i == songindex {
		go audioplayer.Play(playlistFiles[songindex])
	}

	// if we delete a song that is before the current one
	// match the songindex to the new list
	if i < songindex {
		songindex--
	}

	drawplaylist()
	myTui.playlist.SetCurrentItem(i)
}

// draw the progressbar
func drawprogressbar(playtime time.Duration, length time.Duration) {
	myTui.progressbar.Clear()
	_, _, width, _ := myTui.progressbar.GetInnerRect()
	fill := width * int(playtime) / int(length)
	for i := 0; i < fill-1; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneCkBoard)
	}
	fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneBlock)
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

	ph, pm, ps := int64(playtime.Hours()), int64(playtime.Minutes()), int64(playtime.Seconds())
	myTui.playtime.Clear()
	fmt.Fprintf(myTui.playtime, "%02d:%02d:%02d", ph, pm-ph*60, ps-pm*60)

	th, tm, ts := int64(length.Hours()), int64(length.Minutes()), int64(length.Seconds())
	myTui.totaltime.Clear()
	fmt.Fprintf(myTui.totaltime, "%02d:%02d:%02d", th, tm-th*60, ts-tm*60)

	myTui.app.Draw()
}

// shuffle the playlist
func shuffle() {
	if len(playlistFiles) == 0 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(playlistFiles), func(i, j int) {
		playlistFiles[i], playlistFiles[j] = playlistFiles[j], playlistFiles[i]
	})
	songindex = 0
	go audioplayer.Play(playlistFiles[songindex])
	drawplaylist()
}

func changedirDatabase() {
	var base = directorylistFolders[myTui.directorylist.GetCurrentItem()]

	// add files
	filelistFiles = database.GetTracksByFolderID(base.Id)

	// only add parent folder when we are not in the root directory
	var isRoot = base.Id == 1
	if !isRoot {
		directorylistFolders = []globals.Folder{database.GetFolderByID(base.ParentID)}
	} else {
		directorylistFolders = nil
	}

	//add the rest of the folders
	directorylistFolders = append(directorylistFolders, database.GetFoldersByParentID(base.Id)...)

	drawdirectorylist(changedirDatabase, isRoot)
	drawfilelist()
}

// navigate the file manager
func changedirFilesystem() {
	var base = directorylistFolders[myTui.directorylist.GetCurrentItem()]
	var isRoot = base.Path == "/"

	files, _ := ioutil.ReadDir(base.Path)

	directorylistFolders = nil
	filelistFiles = nil

	if !isRoot {
		// add parent folder
		var folder = globals.Folder{
			Id:       -1,
			Path:     path.Clean(path.Join(base.Path, "..")),
			ParentID: -1}

		directorylistFolders = append(directorylistFolders, folder)
	}

	// loop over files
	for _, file := range files {

		//ignore hidden files
		if file.Name()[0] == '.' {
			continue
		}

		// if we've encountered a directory, add it to the directorylist
		if file.IsDir() {
			var folder = globals.Folder{
				Id:       -1,
				Path:     path.Join(base.Path, file.Name()),
				ParentID: -1}

			directorylistFolders = append(directorylistFolders, folder)
		} else {

			// if we've encountered a playable file, add it to the file list
			if !globals.Contains(globals.Formats, strings.ToLower(path.Ext(file.Name()))) {
				continue
			}

			// read metadata
			f, _ := os.Open(path.Join(base.Path, file.Name()))
			m, err := tag.ReadFrom(f)

			var track globals.Track

			// if no tags were found default to nil
			if err != nil {
				track = globals.Track{
					Id:       -1,
					Path:     path.Join(base.Path, file.Name()),
					FolderID: -1,
					Title:    sql.NullString{String: file.Name(), Valid: true},
					Album:    sql.NullString{String: "", Valid: false},
					Artist:   sql.NullString{String: "", Valid: false},
					Genre:    sql.NullString{String: "", Valid: false},
					Year:     sql.NullInt64{Int64: -1, Valid: false}}
			} else {
				track = globals.Track{
					Id:       -1,
					Path:     path.Join(base.Path, file.Name()),
					FolderID: -1,
					Title:    database.StringToSQLNullableString(m.Title()),
					Artist:   database.StringToSQLNullableString(m.Artist()),
					Album:    database.StringToSQLNullableString(m.Album()),
					Genre:    database.StringToSQLNullableString(m.Genre()),
					Year:     database.IntToSQLNullableInt(m.Year())}
			}

			filelistFiles = append(filelistFiles, track)

		}
	}
	drawdirectorylist(changedirFilesystem, isRoot)
	drawfilelist()
}

func searchDatabase() {
	var term = myTui.searchinput.GetText()
	filelistFiles = database.GetSearchResults(term)
	closeSearch()
}

func searchFilesystem() {
	var term = myTui.searchinput.GetText()
	filelistFiles = nil
	err := filepath.Walk(root,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() ||
				!globals.Contains(globals.Formats, strings.ToLower(path.Ext(file))) {
				// read metadata
				f, _ := os.Open(file)
				m, err := tag.ReadFrom(f)

				if err == nil {
					if strings.HasPrefix(strings.ToLower(m.Artist()), strings.ToLower(term)) ||
						strings.HasPrefix(strings.ToLower(m.Album()), strings.ToLower(term)) ||
						strings.HasPrefix(strings.ToLower(m.Title()), strings.ToLower(term)) {
						var track globals.Track

						track = globals.Track{
							Id:       -1,
							Path:     file,
							FolderID: -1,
							Title:    database.StringToSQLNullableString(m.Title()),
							Artist:   database.StringToSQLNullableString(m.Artist()),
							Album:    database.StringToSQLNullableString(m.Album()),
							Genre:    database.StringToSQLNullableString(m.Genre()),
							Year:     database.IntToSQLNullableInt(m.Year())}

						filelistFiles = append(filelistFiles, track)
					}
				}
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}
	closeSearch()
}

func openSearch() {
	// don't open a new seach window when one is already open
	if myTui.searchinput.HasFocus() {
		return
	}
	myTui.mainFlex.RemoveItem(myTui.keybinds)
	myTui.mainFlex.AddItem(myTui.searchinput, 3, 0, false)
	myTui.searchinput.SetText("")
	myTui.app.SetFocus(myTui.searchinput)
	drawfilelist()
}

func closeSearch() {
	myTui.mainFlex.RemoveItem(myTui.searchinput)
	myTui.mainFlex.AddItem(myTui.keybinds, 3, 0, false)
	myTui.app.SetFocus(myTui.filelist)
	drawfilelist()
}

func clear() {
	go audioplayer.Stop()
	playlistFiles = nil
	drawplaylist()
}

func jump(r rune) {
	for index, folder := range directorylistFolders {
		if unicode.ToLower(rune(path.Base(folder.Path)[0])) == unicode.ToLower(r) {
			myTui.directorylist.SetCurrentItem(index)
			return
		}
	}
}

func goback() {
	myTui.directorylist.SetCurrentItem(0)
	changedir()
}
