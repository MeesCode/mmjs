// Package database manages everything that has to do with communicating with the database.
package database

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/MeesCode/mmjs/globals"
)

// this stuct holds all prepared statements
var pst preparedStatements

// list of prepared statements
type preparedStatements struct {
	insertFolder         *sql.Stmt
	insertTrack          *sql.Stmt
	findSubFolders       *sql.Stmt
	findFolder           *sql.Stmt
	findFolderByPath     *sql.Stmt
	findTracksInFolder   *sql.Stmt
	searchTracks         *sql.Stmt
	insertPlaylistTrack  *sql.Stmt
	insertPlaylist       *sql.Stmt
	findTracksInPlaylist *sql.Stmt
	findPlaylists        *sql.Stmt
	incrementCounter     *sql.Stmt
}

// Warmup the mysql connection pool
func Warmup() *sql.DB {
	db, err := sql.Open("mysql",
		globals.Config.Database.User+":"+
			globals.Config.Database.Password+"@("+
			globals.Config.Database.Host+":"+
			strconv.Itoa(globals.Config.Database.Port)+")/"+
			globals.Config.Database.Database)

	if err != nil {
		log.Fatalln("connection with database could not be established", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln("connection with database could not be pinged", err)
	}

	prepareStatements(db)
	return db
}

// prepare all statements ahead of time
func prepareStatements(db *sql.DB) {
	var err error
	pst.insertFolder, err = db.Prepare("INSERT IGNORE INTO Folders(Path, ParentID) VALUES(?, ?)")
	pst.insertTrack, err = db.Prepare(`INSERT IGNORE INTO Tracks(Path, FolderID, Title, Album, Artist, Genre, Year) VALUES(?, ?, ?, ?, ?, ?, ?)`)
	pst.findSubFolders, err = db.Prepare(`SELECT FolderId, Path, ParentId FROM 
		Folders WHERE ParentID = ? ORDER BY Path`)
	pst.findFolder, err = db.Prepare(`SELECT FolderId, Path, ParentId FROM 
		Folders WHERE FolderID = ?`)
	pst.findFolderByPath, err = db.Prepare("SELECT FolderID FROM Folders WHERE Path = ?")
	pst.findTracksInFolder, err = db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
		Genre, Year FROM Tracks WHERE FolderID = ?`)
	pst.searchTracks, err = db.Prepare(`SELECT TrackID, Path, FolderID, Title, Album, Artist, 
		Genre, Year FROM Tracks WHERE Artist LIKE ? OR Title LIKE ? OR Path LIKE ? OR Album LIKE ? ORDER BY Album`)
	pst.insertPlaylist, err = db.Prepare(`INSERT INTO Playlists (Name) VALUES (?)`)
	pst.insertPlaylistTrack, err = db.Prepare(`INSERT INTO PlaylistEntries (TrackID, PlaylistID) VALUES (?, ?)`)
	pst.findTracksInPlaylist, err = db.Prepare(`SELECT Tracks.TrackID, Tracks.Path, Tracks.FolderID, 
		Tracks.Title, Tracks.Album, Tracks.Artist, Tracks.Genre, Tracks.Year 
		FROM Tracks 
		JOIN PlaylistEntries ON Tracks.TrackID = PlaylistEntries.TrackID 
		JOIN Playlists ON Playlists.PlaylistID = PlaylistEntries.PlaylistID 
		WHERE Playlists.PlaylistID = ?`)
	pst.findPlaylists, err = db.Prepare(`SELECT PlaylistID, Name FROM Playlists`)
	pst.incrementCounter, err = db.Prepare(`UPDATE Tracks SET Plays = Plays + 1 WHERE TrackID = ?`)

	if err != nil {
		log.Fatalln("could not prepare statements", err)
	}
}

// StringToSQLNullableString converts a string into a nullable string.
func StringToSQLNullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{String: s, Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// IntToSQLNullableInt converts an int into a nullable int.
func IntToSQLNullableInt(s int) sql.NullInt64 {
	var i = int64(s)
	if s == 0 {
		return sql.NullInt64{Int64: i, Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
