// Package database manages everything that has to do with communicating with the database.
package database

import (
	"database/sql"
	"log"

	"github.com/MeesCode/mmjs/globals"
)

////////////////////////////////////////////////
// root folder always has id 1 and parentid 0 //
////////////////////////////////////////////////

// GetFoldersByParentID returns the folders with the provided ParentID.
func GetFoldersByParentID(parentid int) []globals.Folder {
	folders := make([]globals.Folder, 0)

	rows, err := db.Query(stmts.findSubFolders, parentid)
	if err != nil {
		log.Println("could not find folder", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var folder globals.Folder
		err = rows.Scan(&folder.ID, &folder.Path, &folder.ParentID)

		if err != nil {
			log.Println("Could not find folder", err)
		} else {
			folders = append(folders, folder)
		}

	}

	return folders

}

// GetFolderByID returns the folder with the provided ID.
func GetFolderByID(folderid int) globals.Folder {
	var folder globals.Folder

	err := db.QueryRow(stmts.findFolder, folderid).Scan(&folder.ID, &folder.Path, &folder.ParentID)
	if err != nil {
		log.Fatalln("could not prepare statement. Did you forget to run index mode first?", err)
	}

	return folder

}

// GetTracksByFolderID returns all tracks that are in a given folder.
func GetTracksByFolderID(folderid int) []globals.Track {
	tracks := make([]globals.Track, 0)

	rows, err := db.Query(stmts.findTracksInFolder, folderid)
	if err != nil {
		log.Println("Could not find folder", err)
		return nil
	}
	defer rows.Close()

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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// GetSearchResults searches the database for a specific term and
// return the results. The results are found by checking if the given search term matches
// the beginning of either the Title, Artist or Album name. Results are ordered by album.
// Tries to remove duplicates.
func GetSearchResults(term string) []globals.Track {

	tracks := make([]globals.Track, 0)

	rows, err := db.Query(stmts.searchTracks, term+"%", term+"%", "%"+term+"%", term+"%")
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
	}
	defer rows.Close()

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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {

			//check for duplicates
			dup := false
			for _, t := range tracks {
				if track.Artist == t.Artist && track.Title == t.Title {
					dup = true
					break
				}
			}

			if !dup {
				tracks = append(tracks, track)
			}

		}

	}

	return tracks

}

// GetRandomTracks get n random tracks from the database
func GetRandomTracks(n int) []globals.Track {

	if n < 1 {
		return nil
	}

	tracks := make([]globals.Track, 0)

	rows, err := db.Query(stmts.randomTracks, n)
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
	}
	defer rows.Close()

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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// GetPopularTracks get n popular tracks from the database
func GetPopularTracks(n int) []globals.Track {

	if n < 1 {
		return nil
	}

	tracks := make([]globals.Track, 0)

	rows, err := db.Query(stmts.popularTracks, n)
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
	}
	defer rows.Close()

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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// SavePlaylist saves aplaylist to the database
func SavePlaylist(name string, tracks []globals.Track) {
	res, err := db.Exec(stmts.insertPlaylist, name)
	id, err2 := res.LastInsertId()
	if err != nil || err2 != nil {
		log.Fatalln("could not create playlist", err, err2)
	}

	for _, track := range tracks {
		db.Exec(stmts.insertPlaylistTrack, track.ID, id)
	}

}

// GetPlaylistTracks return all tracks in a playlist
func GetPlaylistTracks(playlistid int) []globals.Track {
	tracks := make([]globals.Track, 0)

	rows, err := db.Query(stmts.findTracksInPlaylist, playlistid)
	if err != nil {
		log.Println("Could not perform query", err)
		return nil
	}
	defer rows.Close()

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
			log.Println("Could not find track in database", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// GetPlaylists searches the database for all playlists and return them desguised as tracks
func GetPlaylists() []globals.Track {
	playlists := make([]globals.Track, 0)

	rows, err := db.Query(stmts.findPlaylists)
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var playlist globals.Track
		err = rows.Scan(
			&playlist.ID,
			&playlist.Title)

		playlist.Album = sql.NullString{String: "playlist", Valid: true}
		playlist.Artist = sql.NullString{String: "playlist", Valid: true}
		playlist.FolderID = -1
		playlist.Genre = sql.NullString{String: "playlist", Valid: true}
		playlist.Path = "not applicable"
		playlist.Year = sql.NullInt64{Int64: 0, Valid: false}

		if err != nil {
			log.Println("Could not find playlist", err)
		} else {
			playlists = append(playlists, playlist)
		}

	}

	return playlists

}

// IncrementPlayCounter increments the play counter of a given track by one
func IncrementPlayCounter(track_id int) {
	_, err := db.Exec(stmts.incrementCounter, track_id)
	if err != nil {
		log.Println("Could not increment the play counter", err)
	}
}
