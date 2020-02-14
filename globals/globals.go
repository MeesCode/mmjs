//////////////////////////////////////////////////////////////////////////////////
// go does not suppot circular dependencies, therefore this is it's own package //
//////////////////////////////////////////////////////////////////////////////////

package globals

import (
	"database/sql"
	"time"
)

var (
	Speakercommand = make(chan string)
	Playfile       = make(chan Track)
	Audiostate     = make(chan AudioStats)
	DurationState  = make(chan DurationStats)
	Formats        = []string{".wav", ".mp3", ".ogg", ".flac"}
)

type AudioStats struct {
	Track  Track
	Length time.Duration
}

type DurationStats struct {
	Playtime time.Duration
	Length   time.Duration
}

type DatabaseConnection struct {
	Username string
	Password string
	Database string
}

type Folder struct {
	Id       int
	Path     string
	ParentID int
}

type Track struct {
	Id       int
	Path     string
	FolderID int
	Title    sql.NullString
	Album    sql.NullString
	Artist   sql.NullString
	Genre    sql.NullString
	Year     sql.NullInt64
}

// helper function to check if an array cointains a specific string
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func StringToSqlNullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{String: s, Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func IntToSqlNullableInt(s int) sql.NullInt64 {
	var i = int64(s)
	if s == 0 {
		return sql.NullInt64{Int64: i, Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
