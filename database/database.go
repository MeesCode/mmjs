// Package database manages everything that has to do with communicating with the database.
package database

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/MeesCode/mmjs/globals"
)

// this stuct holds all defined statements
var stmts definedStatements
var db *sql.DB

// list of defined statements
type definedStatements struct {
	insertFolder         string
	insertTrack          string
	findSubFolders       string
	findFolder           string
	findFolderByPath     string
	findTracksInFolder   string
	searchTracks         string
	insertPlaylistTrack  string
	insertPlaylist       string
	findTracksInPlaylist string
	findPlaylists        string
	incrementCounter     string
	randomTracks         string
	popularTracks        string
}

// Warmup the mysql connection pool
func Warmup() *sql.DB {
	defineStatements()

	dbc, err := sql.Open("mysql",
		globals.Config.Database.User+":"+
			globals.Config.Database.Password+"@("+
			globals.Config.Database.Host+":"+
			strconv.Itoa(globals.Config.Database.Port)+")/"+
			globals.Config.Database.Database)

	dbc.SetConnMaxLifetime(time.Minute * 5)

	if err != nil {
		log.Fatalln("connection with database could not be established", err)
	}

	err = dbc.Ping()
	if err != nil {
		log.Fatalln("connection with database could not be pinged", err)
	}

	db = dbc

	return dbc
}

// define all statements ahead of time
func defineStatements() {
	stmts.insertFolder = "INSERT IGNORE INTO Folders(Path, ParentID) VALUES(?, ?)"
	stmts.insertTrack = `INSERT IGNORE INTO Tracks(Path, FolderID, Title, Album, Artist, Genre, Year) VALUES(?, ?, ?, ?, ?, ?, ?)`
	stmts.findSubFolders = `SELECT FolderId, Path, ParentId FROM 
		Folders WHERE ParentID = ? ORDER BY Path`
	stmts.findFolder = `SELECT FolderId, Path, ParentId FROM 
		Folders WHERE FolderID = ?`
	stmts.findFolderByPath = "SELECT FolderID FROM Folders WHERE Path = ?"
	stmts.findTracksInFolder = `SELECT TrackID, Path, FolderID, Title, Album, Artist, 
		Genre, Year FROM Tracks WHERE FolderID = ?`
	stmts.searchTracks = `SELECT TrackID, Path, FolderID, Title, Album, Artist, 
		Genre, Year FROM Tracks WHERE Artist LIKE ? OR Title LIKE ? OR Path LIKE ? OR Album LIKE ? ORDER BY Album`
	stmts.insertPlaylist = `INSERT INTO Playlists (Name) VALUES (?)`
	stmts.insertPlaylistTrack = `INSERT INTO PlaylistEntries (TrackID, PlaylistID) VALUES (?, ?)`
	stmts.findTracksInPlaylist = `SELECT Tracks.TrackID, Tracks.Path, Tracks.FolderID, 
		Tracks.Title, Tracks.Album, Tracks.Artist, Tracks.Genre, Tracks.Year 
		FROM Tracks 
		JOIN PlaylistEntries ON Tracks.TrackID = PlaylistEntries.TrackID 
		JOIN Playlists ON Playlists.PlaylistID = PlaylistEntries.PlaylistID 
		WHERE Playlists.PlaylistID = ?`
	stmts.findPlaylists = `SELECT PlaylistID, Name FROM Playlists`
	stmts.incrementCounter = `UPDATE Tracks SET Plays = Plays + 1 WHERE TrackID = ?`
	stmts.randomTracks = `SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year FROM Tracks ORDER BY RAND() LIMIT ?`
	stmts.popularTracks = `SELECT TrackID, Path, FolderID, Title, Album, Artist, 
	Genre, Year, Plays FROM Tracks ORDER BY Plays DESC LIMIT ?`
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
