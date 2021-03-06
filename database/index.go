// Package database manages everything that has to do with communicating with the database.
package database

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MeesCode/mmjs/globals"

	"github.com/dhowden/tag"

	// this is needed but i don't know why it is blank
	// probably just does some stuff in the background
	_ "github.com/go-sql-driver/mysql"
)

// Index indexes every folder and playable file that is contained within the
// specified root folder. It ignores hidden folders entirely.
func Index() {

	var err error
	findFolderByPath, err := db.Prepare(stmts.findFolderByPath)
	if err != nil {
		log.Fatalln("could not prepare statements")
	}

	insertFolder, err := db.Prepare(stmts.insertFolder)
	if err != nil {
		log.Fatalln("could not prepare statements")
	}

	insertTrack, err := db.Prepare(stmts.insertTrack)
	if err != nil {
		log.Fatalln("could not prepare statements")
	}

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
					err = findFolderByPath.QueryRow(path.Dir(rpath)).Scan(&parentID)
					if err != nil {
						log.Println("Could not perform query, or query returned empty. query: ", path.Dir(rpath), err)
					}
				}

				// if it's a folder
				if info.IsDir() {
					if isRoot {
						// special case for when it's the root folder
						_, err = insertFolder.Exec("/", parentID)
						if err != nil {
							log.Println("Could not add root to the database", err)
						}
					} else {
						_, err = insertFolder.Exec(rpath, parentID)
						if err != nil {
							log.Println("Could not add folder to the database", err)
						}
					}

					// if it's a file
				} else {

					// if we've encountered a playable file, add it to the file list
					if globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file))) {

						// read metadata
						f, _ := os.Open(file)
						m, err := tag.ReadFrom(f)

						// if no tags were found default to nil
						if err != nil {
							_, err = insertTrack.Exec(rpath, parentID, nil, nil, nil, nil, nil)
						} else {
							_, err = insertTrack.Exec(
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
			}

			return nil
		})
	if err != nil {
		log.Fatalln("Could not walk the filesystem at the given location", err)
	}

	return
}
