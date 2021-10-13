// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"
)

// changedirDatabase changes the current directory (when in database mode) to
// the one that is selected.
func changedirDatabase() {
	myTui.filelist.SetTitle(" Current directory ")
	var base = directorylistFolders[myTui.directorylist.GetCurrentItem()]

	// add files
	filelistFiles = database.GetTracksByFolderID(base.ID)

	// only add parent folder when we are not in the root directory
	var isRoot = base.ID == 1
	if !isRoot {
		directorylistFolders = []globals.Folder{database.GetFolderByID(base.ParentID)}
	} else {
		directorylistFolders = nil
	}

	//add the rest of the folders
	directorylistFolders = append(directorylistFolders, database.GetFoldersByParentID(base.ID)...)

	drawdirectorylist(changedirDatabase, isRoot)
	drawfilelist()
}

// get 100 most popular tracks
func getPopular(){
	filelistFiles = database.GetPopularTracks(100)
	drawfilelistWithPlays()
}

// searchDatabase searches (while in database mode) for the tracks that match on
// either the title, album or artist. It uses the text that is currently entered in the searchbox.
func searchDatabase() {
	var term = myTui.searchinput.GetText()
	filelistFiles = database.GetSearchResults(term)
	closeSearch(true)
}

func searchDatabaseQuery(query string) {
	openSearch()
	filelistFiles = database.GetSearchResults(query)
	closeSearch(true)
}

// addFolderDatabaseRec is a recursive function that takes a folder and add all
// containing tracks to the playlist, after which it will call itself for every
// child folder. This function should be called when in database mode.
func addFolderDatabaseRec(folder globals.Folder) {
	// add tracks from current folder
	tracks := database.GetTracksByFolderID(folder.ID)
	audioplayer.Playlist = append(audioplayer.Playlist, tracks...)

	// add children recusively
	folders := database.GetFoldersByParentID(folder.ID)
	for _, folder := range folders {
		addFolderDatabaseRec(folder)
	}
}

// addFolderDatabase adds all tracks inside the currently selected folder to the playlist.
// This includes all tracks inside child folders.
func addFolderDatabase() {
	addFolderDatabaseRec(directorylistFolders[myTui.directorylist.GetCurrentItem()])
	drawplaylist()
}

func savePlaylist() {
	var name = myTui.playlistinput.GetText()
	if name == "" {
		closePlaylist()
		return
	}
	database.SavePlaylist(name, audioplayer.Playlist)
	closePlaylist()
}

// openSearch removes the keybinds box and replaces it with the search box.
func openPlaylistInput() {
	if myTui.searchinput.HasFocus() || myTui.playlistinput.HasFocus() {
		return
	}
	myTui.pages.AddPage("playlist", myTui.playlistbox, true, true)
	myTui.playlistinput.SetText("")
	focusWithColor(myTui.playlistinput)
}

// closeSearch removes the search box and replaces it with the keybinds box.
func closePlaylist() {
	myTui.pages.RemovePage("playlist")
	focusWithColor(myTui.filelist)
	drawfilelist()
}

func insertPlaylist() {
	audioplayer.Stop()
	audioplayer.Songindex = 0
	pl := filelistFiles[myTui.filelist.GetCurrentItem()]
	audioplayer.Playlist = database.GetPlaylistTracks(pl.ID)
	drawplaylist()
}

func showPlaylists() {
	myTui.filelist.SetTitle(" Playlists ")
	filelistFiles = database.GetPlaylists()
	myTui.filelist.Clear()
	for _, track := range filelistFiles {
		myTui.filelist.AddItem(trackToDisplayText(track), "", 0, insertPlaylist)
	}
	focusWithColor(myTui.filelist)
}
