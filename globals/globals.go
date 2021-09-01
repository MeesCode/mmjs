// Package globals exists because go does not suppot circular dependencies,
// therefore this is it's own package.
package globals

import (
	"database/sql"
)

// Root is the root folder where the player is initialized
var Root string

// ConfigFile is a struct that holds all database connection info.
type ConfigFile struct {
	Mode     string `json:"mode"`
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"database"`
	Highlight    string `json:"highlight"`
	Quiet        bool   `json:"quiet"`
	Logging      bool   `json:"logging"`
	DisableSound bool   `json:"disableSound"`
	Webserver    struct {
		Enable bool `json:"enable"`
		Port   int  `json:"port"`
	} `json:"webserver"`
	Webinterface    struct {
		Enable bool `json:"enable"`
		Port   int  `json:"port"`
	} `json:"webinterface"`
	Serial    struct {
		Enable bool   `json:"enable"`
		Port   string `json:"port"`
	} `json:"serial"`
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
	Plays    int
}

// Config is the variable that holder the config file
var Config ConfigFile

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
