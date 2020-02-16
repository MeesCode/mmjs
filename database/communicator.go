package database

import (
	"mp3bak2/globals"
)

////////////////////////////////////////////////
// root folder always has id 1 and parentid 0 //
////////////////////////////////////////////////

func GetFoldersByParentID(parentid int) []globals.Folder {
	db := getConnection()

	folderOut, err := db.Prepare("SELECT FolderId, Path, ParentId FROM Folders WHERE ParentID = ? ORDER BY Path")
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
		err = rows.Scan(&folder.Id, &folder.Path, &folder.ParentID)
		if err != nil {
			panic(err.Error())
		}

		folders = append(folders, folder)

	}

	return folders

}

func GetFolderByID(folderid int) globals.Folder {
	db := getConnection()

	folderOut, err := db.Prepare("SELECT FolderId, Path, ParentId FROM Folders WHERE FolderID = ?")
	if err != nil {
		panic(err.Error())
	}
	defer folderOut.Close()

	var folder globals.Folder

	err = folderOut.QueryRow(folderid).Scan(&folder.Id, &folder.Path, &folder.ParentID)
	if err != nil {
		panic(err.Error())
	}

	return folder

}

func GetTracksByFolderID(folderid int) []globals.Track {
	db := getConnection()

	folderOut, err := db.Prepare("SELECT TrackID, Path, FolderID, Title, Album, Artist, Genre, Year FROM Tracks WHERE FolderID = ?")
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
			&track.Id,
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

func GetSearchResults(term string) []globals.Track {
	db := getConnection()

	trackOut, err := db.Prepare("SELECT TrackID, Path, FolderID, Title, Album, Artist, Genre, Year FROM Tracks WHERE Artist LIKE ? OR Title LIKE ? OR Album LIKE ? ORDER BY Album")
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
			&track.Id,
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
