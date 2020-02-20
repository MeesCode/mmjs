// Package database manages everything that has to do with communicating with the database.
package database

import (
	"fmt"
	"mp3bak2/globals"
)

////////////////////////////////////////////////
// root folder always has id 1 and parentid 0 //
////////////////////////////////////////////////

// GetFoldersByParentID returns the folder with the provided ParentID.
func GetFoldersByParentID(parentid int) []globals.Folder {
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT FolderId, Path, ParentId FROM 
	Folders WHERE ParentID = ? ORDER BY Path`)
	if err != nil {
		panic(err.Error())
	}
	defer folderOut.Close()

	folders := make([]globals.Folder, 0)

	rows, err := folderOut.Query(parentid)
	if err != nil {
		panic(err.Error())
	}

	for rows.Next() {
		var folder globals.Folder
		err = rows.Scan(&folder.ID, &folder.Path, &folder.ParentID)
		if err != nil {
			panic(err.Error())
		}

		folders = append(folders, folder)

	}

	return folders

}

// GetFolderByID returns the folder with the provided ID.
func GetFolderByID(folderid int) globals.Folder {
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT FolderId, Path, ParentId FROM 
	Folders WHERE FolderID = ?`)
	if err != nil {
		panic(err.Error())
	}
	defer folderOut.Close()

	var folder globals.Folder

	err = folderOut.QueryRow(folderid).Scan(&folder.ID, &folder.Path, &folder.ParentID)
	if err != nil {
		fmt.Println("did you forget to run index mode first?")
		panic(err.Error())
	}

	return folder

}

// GetTracksByFolderID returns all tracks that are in a given folder.
func GetTracksByFolderID(folderid int) []globals.Track {
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year FROM Tracks WHERE FolderID = ?`)
	if err != nil {
		panic(err.Error())
	}
	defer folderOut.Close()

	tracks := make([]globals.Track, 0)

	rows, err := folderOut.Query(folderid)
	if err != nil {
		panic(err.Error())
	}

	for rows.Next() {
		var track globals.Track
		err = rows.Scan(
			&track.ID,
			&track.Path,
			&track.FolderID,
			&track.Title,
			&track.Album,
			&track.Artist,
			&track.Genre,
			&track.Year)
		if err != nil {
			panic(err.Error())
		}

		tracks = append(tracks, track)

	}

	return tracks

}

// GetSearchResults searches the database for a specific term and
// return the results. The results a found by checking if the given search term matches
// the beginning of either the Title, Artist or Album name. Results are ordered by album.
func GetSearchResults(term string) []globals.Track {
	db := getConnection()

	trackOut, err := db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year FROM Tracks WHERE Artist LIKE ? OR Title LIKE ? OR Album LIKE ? ORDER BY Album`)
	if err != nil {
		panic(err.Error())
	}
	defer trackOut.Close()

	tracks := make([]globals.Track, 0)

	rows, err := trackOut.Query(term+"%", term+"%", term+"%")
	if err != nil {
		panic(err.Error())
	}

	for rows.Next() {
		var track globals.Track
		err = rows.Scan(
			&track.ID,
			&track.Path,
			&track.FolderID,
			&track.Title,
			&track.Album,
			&track.Artist,
			&track.Genre,
			&track.Year)
		if err != nil {
			panic(err.Error())
		}

		tracks = append(tracks, track)

	}

	return tracks

}
