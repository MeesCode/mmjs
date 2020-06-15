// Package globals exists because go does not suppot circular dependencies,
// therefore this is it's own package.
package globals

import (
	"database/sql"
)

// Root is the root folder where the player is initialized
var Root string

// Port is the port on which the webserver runs
var Port = "8080"

// DatabaseConnection is a struct that holds all database connection info.
type DatabaseConnection struct {
	Username string
	Password string
	Database string
	Host     string
	Port     string
}

// Folder is a struct that holds all folder info, this correlates directly
// to what is in the database. It is also used in filesystem mode but only to
// hold the path.
type Folder struct {
	ID       int
	Path     string
	ParentID int
}

// Track is a struct that holds all folder info, this correlates directly
// to what is in the database. It is also used in filesystem mode but only to
// hold the meta tags.
type Track struct {
	ID       int
	Path     string
	FolderID int
	Title    sql.NullString
	Album    sql.NullString
	Artist   sql.NullString
	Genre    sql.NullString
	Year     sql.NullInt64
}

// GetSupportedFormats returns an array of strings with the file extentions
// that are supported by the program.
func GetSupportedFormats() []string {
	return []string{".wav", ".mp3", ".ogg", ".flac"}
}

// Contains is helper function to check if an array cointains a specific string.
// Often used in correlation with GetSupportedFormats().
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
