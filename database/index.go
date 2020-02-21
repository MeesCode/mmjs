// Package database manages everything that has to do with communicating with the database.
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

	// this is needed but i don't know why it is blank
	// probably just does some stuff in the background
	_ "github.com/go-sql-driver/mysql"
)

// Index indexes every folder and playable file that is contained within the
// specified root folder. It ignores hidden folders entirely.
func Index() {

	db := getConnection()

	// Prepare statement for inserting a folder
	folderIns, err := db.Prepare("INSERT IGNORE INTO Folders(Path, ParentID) VALUES(?, ?)")
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer folderIns.Close()

	// Prepare statement for inserting a file
	fileIns, err := db.Prepare(`INSERT IGNORE INTO Tracks(Path, FolderID, Title, 
		Album, Artist, Genre, Year) VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer fileIns.Close()

	// Prepare statement for finding parent folder
	parentOut, err := db.Prepare("SELECT FolderID FROM Folders WHERE Path = ?")
	if err != nil {
		log.Fatalln("could not prepare statement", err)
	}
	defer parentOut.Close()

	err = filepath.Walk(globals.Root,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// make sure it's a playable file
			if info.IsDir() ||
				globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file))) {

				// skip hidden folders
				if strings.Contains(file, "/.") {
					return filepath.SkipDir
				}

				// relative path
				rpath := path.Clean(file[len(globals.Root):])

				fmt.Println(file)

				var isRoot = globals.Root == file

				var parentID = 0

				if !isRoot {
					err = parentOut.QueryRow(path.Dir(rpath)).Scan(&parentID)
					if err != nil {
						log.Println("Could not perform query, or query returned empty. query: ", path.Dir(rpath), err)
					}
				}

				// if it's a folder
				if info.IsDir() {
					if isRoot {
						// special case for when it's the root folder
						_, err = folderIns.Exec("/", parentID)
						if err != nil {
							log.Println("Could not add root to the database", err)
						}
					} else {
						_, err = folderIns.Exec(rpath, parentID)
						if err != nil {
							log.Println("Could not add folder to the database", err)
						}
					}

					// if it's a file
				} else {

					// read metadata
					f, _ := os.Open(file)
					m, err := tag.ReadFrom(f)

					// if no tags were found default to nil
					if err != nil {
						_, err = fileIns.Exec(rpath, parentID, nil, nil, nil, nil, nil)
					} else {
						_, err = fileIns.Exec(
							rpath,
							parentID,
							StringToSQLNullableString(m.Title()),
							StringToSQLNullableString(m.Album()),
							StringToSQLNullableString(m.Artist()),
							StringToSQLNullableString(m.Genre()),
							IntToSQLNullableInt(m.Year()))
					}

				}
			}

			return nil
		})
	if err != nil {
		log.Fatalln("Could not walk the filesystem at the given location", err)
	}

	return
}
