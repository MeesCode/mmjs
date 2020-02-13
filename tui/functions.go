package tui

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mp3bak2/database"
	"mp3bak2/globals"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// play the song currently selected on the playlist
func playsong() {
	interfaceLock.Lock()
	defer interfaceLock.Unlock()

	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	globals.Playfile <- playlistFiles[myTui.playlist.GetCurrentItem()]
}

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
			dir, name := path.Split(data.Track.Path)
			myTui.infobox.SetCell(0, 1, tview.NewTableCell(stringOrUnknown(data.Track.Title)))
			myTui.infobox.SetCell(1, 1, tview.NewTableCell(stringOrUnknown(data.Track.Artist)))
			myTui.infobox.SetCell(2, 1, tview.NewTableCell(stringOrUnknown(data.Track.Album)))
			myTui.infobox.SetCell(3, 1, tview.NewTableCell(stringOrUnknown(data.Track.Genre)))
			if data.Track.Year.Valid && data.Track.Year.Int64 != 0 {
				myTui.infobox.SetCell(4, 1, tview.NewTableCell(strconv.FormatInt(data.Track.Year.Int64, 10)))
			} else {
				myTui.infobox.SetCell(4, 1, tview.NewTableCell("unknown"))
			}
			myTui.infobox.SetCell(5, 1, tview.NewTableCell(name))
			myTui.infobox.SetCell(6, 1, tview.NewTableCell(dir))
			drawprogressbar(time.Duration(0), data.Length)
			myTui.app.Draw()

		case data := <-globals.DurationState:
			drawprogressbar(data.Playtime, data.Length)
			if data.Playtime == data.Length {
				nextsong()
			}
			myTui.app.Draw()

		}
	}
}

// draw the playlist
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
	myTui.app.Draw()
}

// draw the file list
func drawfilelist() {
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(trackToDisplayText(track), "", 0, addsong)
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

// go to the next song (if available)
func nextsong() {
	interfaceLock.Lock()
	defer interfaceLock.Unlock()

	if len(playlistFiles) > songindex+1 {
		songindex++
		drawplaylist()
		globals.Playfile <- playlistFiles[songindex]
	}
}

// go to the previous song (if available)
func previoussong() {
	interfaceLock.Lock()
	defer interfaceLock.Unlock()

	if songindex > 0 {
		songindex--
		drawplaylist()
		globals.Playfile <- playlistFiles[songindex]
	}
}

// add a song to the playlist
func addsong() {
	interfaceLock.Lock()
	defer interfaceLock.Unlock()

	track := filelistFiles[myTui.filelist.GetCurrentItem()]
	playlistFiles = append(playlistFiles, track)
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// draw the progressbar
func drawprogressbar(playtime time.Duration, length time.Duration) {
	myTui.progressbar.Clear()
	_, _, width, _ := myTui.progressbar.GetInnerRect()
	fill := int(float64(width) * playtime.Seconds() / float64(length.Seconds()))
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

	if length != time.Duration(0) {
		th, tm, ts := int64(length.Hours()), int64(length.Minutes()), int64(length.Seconds())
		myTui.totaltime.Clear()
		fmt.Fprintf(myTui.totaltime, "%02d:%02d:%02d", th, tm-th*60, ts-tm*60)
	}

}

// shuffle the playlist
func shuffle() {
	interfaceLock.Lock()
	defer interfaceLock.Unlock()

	if len(playlistFiles) == 0 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(playlistFiles), func(i, j int) { playlistFiles[i], playlistFiles[j] = playlistFiles[j], playlistFiles[i] })
	songindex = 0
	globals.Playfile <- playlistFiles[songindex]
	drawplaylist()
}

func changedirDatabase() {
	var root = directorylistFolders[myTui.directorylist.GetCurrentItem()]

	// add files
	filelistFiles = database.GetTracksByFolderID(root.Id)

	// only add parent folder when we are not in the root directory
	var isRoot = root.Id == 1
	if !isRoot {
		directorylistFolders = []globals.Folder{database.GetFolderByID(root.ParentID)}
	} else {
		directorylistFolders = nil
	}

	//add the rest of the folders
	directorylistFolders = append(directorylistFolders, database.GetFoldersByParentID(root.Id)...)

	drawdirectorylist(changedirDatabase, isRoot)
	drawfilelist()
}

// navigate the file manager
func changedirFilesystem() {
	var root = directorylistFolders[myTui.directorylist.GetCurrentItem()].Path
	var isRoot = root == "/"

	files, _ := ioutil.ReadDir(root)

	directorylistFolders = nil
	filelistFiles = nil

	if !isRoot {
		// add parent folder
		var folder = globals.Folder{
			Id:       -1,
			Path:     path.Clean(path.Join(root, "..")),
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
				Path:     path.Join(root, file.Name()),
				ParentID: -1}

			directorylistFolders = append(directorylistFolders, folder)
		} else {

			// if we've encountered a playable file, add it to the file list
			if globals.Contains(globals.Formats, strings.ToLower(path.Ext(file.Name()))) {
				var track = globals.Track{
					Id:       -1,
					Path:     path.Join(root, file.Name()),
					FolderID: -1,
					Title:    sql.NullString{String: file.Name(), Valid: true},
					Album:    sql.NullString{String: "", Valid: false},
					Artist:   sql.NullString{String: "", Valid: false},
					Genre:    sql.NullString{String: "", Valid: false},
					Year:     sql.NullInt64{Int64: -1, Valid: false}}

				filelistFiles = append(filelistFiles, track)
			}
		}
	}
	drawdirectorylist(changedirFilesystem, isRoot)
	drawfilelist()
}
