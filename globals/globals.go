//////////////////////////////////////////////////////////////////////////////////
// go does not suppot circular dependencies, therefore this is it's own package //
//////////////////////////////////////////////////////////////////////////////////

package globals

import (
	"database/sql"
)

var (
	Formats = []string{".wav", ".mp3", ".ogg", ".flac"}
)

type DatabaseConnection struct {
	Username string
	Password string
	Database string
	Host     string
	Port     string
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
