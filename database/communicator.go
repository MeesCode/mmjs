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

// GetFoldersByParentID returns the folder with the provided ParentID.
func GetFoldersByParentID(parentid int) []globals.Folder {
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT FolderId, Path, ParentId FROM 
	Folders WHERE ParentID = ? ORDER BY Path`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer folderOut.Close()

	folders := make([]globals.Folder, 0)

	rows, err := folderOut.Query(parentid)
	if err != nil {
		log.Println("could not find folder", err)
		return nil
	}

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
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT FolderId, Path, ParentId FROM 
	Folders WHERE FolderID = ?`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer folderOut.Close()

	var folder globals.Folder

	err = folderOut.QueryRow(folderid).Scan(&folder.ID, &folder.Path, &folder.ParentID)
	if err != nil {
		log.Fatalln("could not prepare statement. Did you forget to run index mode first?", err)
	}

	return folder

}

// GetTracksByFolderID returns all tracks that are in a given folder.
func GetTracksByFolderID(folderid int) []globals.Track {
	db := getConnection()

	folderOut, err := db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year FROM Tracks WHERE FolderID = ?`)
	if err != nil {
		log.Println("could not prepare statement. Did you forget to run index mode first?", err)
	}
	defer folderOut.Close()

	tracks := make([]globals.Track, 0)

	rows, err := folderOut.Query(folderid)
	if err != nil {
		log.Println("Could not find folder", err)
		return nil
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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// GetSearchResults searches the database for a specific term and
// return the results. The results a found by checking if the given search term matches
// the beginning of either the Title, Artist or Album name. Results are ordered by album.
func GetSearchResults(term string) []globals.Track {
	db := getConnection()

	trackOut, err := db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year FROM Tracks WHERE Artist LIKE ? OR Title LIKE ? OR Path LIKE ? OR Album LIKE ? ORDER BY Album`)
	if err != nil {
		log.Println("could not prepare statement. Did you forget to run index mode first?", err)
	}
	defer trackOut.Close()

	tracks := make([]globals.Track, 0)

	rows, err := trackOut.Query(term+"%", term+"%", "%"+term+"%", term+"%")
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
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
			log.Println("Could not find metadata, file corrupt?", err)
		} else {

			//check for duplicates
			dup := false
			for _, t := range tracks {
				if track.Artist == t.Artist && track.Title == t.Title {
					dup = true
					break;
				}
			}

			if !dup {
				tracks = append(tracks, track)
			}
			
		}

	}

	return tracks

}

// SavePlaylist saves aplaylist to the database
func SavePlaylist(name string, tracks []globals.Track) {
	db := getConnection()

	// Prepare statement for creating a playlist
	plIns, err := db.Prepare(`INSERT INTO Playlists (Name) VALUES (?)`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer plIns.Close()

	// Prepare statement for adding a track to a playlist
	plEntryIns, err := db.Prepare(`INSERT INTO PlaylistEntries (TrackID, PlaylistID) VALUES (?, ?)`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer plEntryIns.Close()

	res, err := plIns.Exec(name)
	id, err2 := res.LastInsertId()
	if err != nil || err2 != nil {
		log.Fatalln("could not create playlist", err, err2)
	}

	for _, track := range tracks {
		plEntryIns.Exec(track.ID, id)
	}

}

// GetPlaylistTracks return all tracks in a playlist
func GetPlaylistTracks(playlistid int) []globals.Track {
	db := getConnection()

	plOut, err := db.Prepare(`SELECT Tracks.TrackID, Tracks.Path, Tracks.FolderID, 
	Tracks.Title, Tracks.Album, Tracks.Artist, Tracks.Genre, Tracks.Year 
	FROM Tracks 
	JOIN PlaylistEntries ON Tracks.TrackID = PlaylistEntries.TrackID 
	JOIN Playlists ON Playlists.PlaylistID = PlaylistEntries.PlaylistID 
	WHERE Playlists.PlaylistID = ?`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer plOut.Close()

	tracks := make([]globals.Track, 0)

	rows, err := plOut.Query(playlistid)
	if err != nil {
		log.Println("Could not perform query", err)
		return nil
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
			log.Println("Could not find track in database", err)
		} else {
			tracks = append(tracks, track)
		}

	}

	return tracks

}

// GetPlaylists searches the database for all playlists and return them desguised as tracks
func GetPlaylists() []globals.Track {
	db := getConnection()

	plOut, err := db.Prepare(`SELECT PlaylistID, Name FROM Playlists`)
	if err != nil {
		log.Println("could not prepare statement.", err)
	}
	defer plOut.Close()

	playlists := make([]globals.Track, 0)

	rows, err := plOut.Query()
	if err != nil {
		log.Println("Could not perform search query", err)
		return nil
	}

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
