package database

import (
	"fmt"
	"log"
	"mp3bak2/globals"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

// Index the specified root folder
func Index(root string) {

	db := getConnection()

	// Prepare statement for inserting a folder
	folderIns, err := db.Prepare("INSERT IGNORE INTO Folders(Path, ParentID) VALUES(?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer folderIns.Close()

	// Prepare statement for inserting a file
	fileIns, err := db.Prepare("INSERT IGNORE INTO Tracks(Path, FolderID, Title, Album, Artist, Genre, Year) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer fileIns.Close()

	// Prepare statement for finding parent folder
	parentOut, err := db.Prepare("SELECT FolderID FROM Folders WHERE Path = ?")
	if err != nil {
		panic(err.Error())
	}
	defer parentOut.Close()

	err = filepath.Walk(root,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// make sure it's a playable file
			if info.IsDir() || globals.Contains(globals.Formats, strings.ToLower(path.Ext(file))) {

				fmt.Println(file)

				var isRoot = root == file

				var parentID = 0

				if !isRoot {
					err = parentOut.QueryRow(path.Dir(file)).Scan(&parentID)
					if err != nil {
						panic(err.Error())
					}
				}

				// if it's a folder
				if info.IsDir() {
					_, err = folderIns.Exec(file, parentID)
					if err != nil {
						panic(err.Error())
					}

					// if it's a file
				} else {

					// read metadata
					f, _ := os.Open(file)
					m, err := tag.ReadFrom(f)

					// if no tags were found default to nil
					if err != nil {
						_, err = fileIns.Exec(file, parentID, nil, nil, nil, nil, nil)
					} else {
						_, err = fileIns.Exec(
							file,
							parentID,
							StringToSQLNullableString(m.Title()),
							StringToSQLNullableString(m.Album()),
							StringToSQLNullableString(m.Artist()),
							StringToSQLNullableString(m.Genre()),
							IntToSQLNullableInt(m.Year()))
					}

					if err != nil {
						panic(err.Error())
					}
				}
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return
}
