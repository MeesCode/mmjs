// Package database manages everything that has to do with communicating with the database.
package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/MeesCode/mmjs/globals"
	"github.com/dhowden/tag"
)

// Reindex goes over all indexed folders starting at the root
// if the amount of tracks has changed the new tracks will be added to the database
func Reindex() {

	// get all paths
	folders, err := pst.findAllFolders.Query()
	if err != nil {
		log.Fatalln("folders could not be found", err)
	}

	// iterate over all folders
	for folders.Next() {
		var folderID int
		var relPath string
		folders.Scan(&folderID, &relPath)

		absPath := path.Clean(path.Join(globals.Root, relPath))

		// open folder
		files, err := ioutil.ReadDir(absPath)
		if err != nil {
			log.Fatalln("folder could not be opened", err)
		}

		// count files in folder
		dirFiles := []os.FileInfo{}
		for _, file := range files {
			if !file.IsDir() {
				dirFiles = append(dirFiles, file)
			}
		}

		// get the amount of files that the db thinks are in the folder
		dbFiles, err := pst.countTracksInFolder.Query(folderID)
		var fileCountDb int
		dbFiles.Scan(&fileCountDb)

		// if the number of files in a given folder differ,
		// reindex the folder
		if fileCountDb != len(dirFiles) {

			// empty folder in database
			pst.emptyFolder.Query(folderID)

			for _, file := range files {

				//ignore hidden files
				if file.Name()[0] == '.' {
					continue
				}

				// if we've encountered a playable file, add it to the file list
				if globals.Contains(globals.GetSupportedFormats(), strings.ToLower(path.Ext(file.Name()))) {

					absFileName := path.Join(absPath, file.Name())
					relFileName := path.Join(relPath, file.Name())

					// read metadata
					f, _ := os.Open(absFileName)
					m, err := tag.ReadFrom(f)

					// if no tags were found default to nil
					if err != nil {
						_, err = pst.insertTrack.Exec(relFileName, folderID, nil, nil, nil, nil, nil)
					} else {

						fmt.Println(relFileName)

						_, err = pst.insertTrack.Exec(
							relFileName,
							folderID,
							StringToSQLNullableString(m.Title()),
							StringToSQLNullableString(m.Album()),
							StringToSQLNullableString(m.Artist()),
							StringToSQLNullableString(m.Genre()),
							IntToSQLNullableInt(m.Year()))
					}
				}

			}
		}
	}
}
