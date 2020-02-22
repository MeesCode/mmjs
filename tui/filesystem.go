// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"io/ioutil"
	"log"
	"mmjs/globals"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// changedirFilesystem changes the current directory (when in filesystem mode) to
// the one that is selected.
func changedirFilesystem() {
	myTui.filelist.SetTitle(" Current directory ")
	var base = directorylistFolders[myTui.directorylist.GetCurrentItem()]
	var isRoot = base.Path == globals.Root

	files, err := ioutil.ReadDir(base.Path)

	if err != nil {
		log.Println("could not find directory to change into", err)
		return
	}

	directorylistFolders = nil
	filelistFiles = nil

	if !isRoot {
		// add parent folder
		var folder = globals.Folder{
			ID:       -1,
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
				ID:       -1,
				Path:     path.Join(base.Path, file.Name()),
				ParentID: -1}

			directorylistFolders = append(directorylistFolders, folder)
		} else {

			// if we've encountered a playable file, add it to the file list
			if !globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file.Name()))) {
				continue
			}

			var track = parseTrack(path.Join(base.Path, file.Name()))
			filelistFiles = append(filelistFiles, track)

		}
	}
	drawdirectorylist(changedirFilesystem, isRoot)
	drawfilelist()
}

// searchFilesystem searches (while in filesystem mode) for the tracks that match on
// either the title, album or artist. It uses the text that is currently entered in the searchbox.
func searchFilesystem() {
	var term = myTui.searchinput.GetText()
	filelistFiles = nil
	err := filepath.Walk(globals.Root,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip hidden folders
			if strings.Contains(file, "/.") {
				return filepath.SkipDir
			}

			if !info.IsDir() ||
				!globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file))) {
				// read metadata
				track := parseTrack(file)

				if strings.HasPrefix(strings.ToLower(track.Artist.String), strings.ToLower(term)) ||
					strings.HasPrefix(strings.ToLower(track.Album.String), strings.ToLower(term)) ||
					strings.HasPrefix(strings.ToLower(track.Title.String), strings.ToLower(term)) {

					filelistFiles = append(filelistFiles, track)
				}
			}

			return nil
		})
	if err != nil {
		log.Println("Could not walk the filesystem at the given location", err)
	}
	closeSearch()
}

// addFolderFilesystem is a recursive function that takes a folder and add all
// containing tracks to the playlist, after which it will call itself for every
// child folder. This function should be called when in filesystem mode.
func addFolderFilesystem() {

	folder := directorylistFolders[myTui.directorylist.GetCurrentItem()]

	err := filepath.Walk(folder.Path,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip hidden folders
			if strings.Contains(file, "/.") {
				return filepath.SkipDir
			}

			if !info.IsDir() && globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file))) {
				playlistFiles = append(playlistFiles, parseTrack(file))
			}

			return nil
		})
	if err != nil {
		log.Println("Could not walk the filesystem at the given location", err)
	}
	drawplaylist()
}
