package tui

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mp3bak2/database"
	"mp3bak2/globals"
	"path"
	"strings"
	"time"

	"github.com/gdamore/tcell"
)

// play the song currently selected on the playlist
func playsong() {
	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	globals.Playfile <- playlistFiles[myTui.playlist.GetCurrentItem()]
}

// draw the playlist
func drawplaylist() {
	myTui.playlist.Clear()
	for index, track := range playlistFiles {
		if songindex == index {
			myTui.playlist.AddItem(track.Title.String, "", '>', playsong)
		} else {
			myTui.playlist.AddItem(track.Title.String, "", 0, playsong)
		}
	}
	myTui.playlist.SetCurrentItem(songindex)
	myTui.app.Draw()
}

// draw the file list
func drawfilelist() {
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(track.Title.String, "", 0, addsong)
	}
	myTui.app.Draw()
}

// draw the directory list
func drawdirectorylist(parentFunc func(), isRoot bool) {
	myTui.directorylist.Clear()
	for index, folder := range directorylistFolders {

		// first folder shows .. instead of the folder name
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
	if len(playlistFiles) > songindex+1 {
		songindex++
		drawplaylist()
		globals.Playfile <- playlistFiles[songindex]
	}
}

// go to the previous song (if available)
func previoussong() {
	if songindex > 0 {
		songindex--
		drawplaylist()
		globals.Playfile <- playlistFiles[songindex]
	}
}

// add a song to the playlist
func addsong() {
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
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}
	fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneBlock)
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

}

// shuffle the playlist
func shuffle() {
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

	files, _ := ioutil.ReadDir(root)

	directorylistFolders = nil
	filelistFiles = nil

	// add parent folder
	var folder = globals.Folder{
		Id:       -1,
		Path:     path.Clean(path.Join(root, "..")),
		ParentID: -1}

	directorylistFolders = append(directorylistFolders, folder)

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
	drawdirectorylist(changedirFilesystem, false)
	drawfilelist()
}
