package database

import (
	"database/sql"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/dhowden/tag"
	_ "github.com/go-sql-driver/mysql"
)

func Index(root string) {

	db, err := sql.Open("mysql" /* connection string here */)
	if err != nil {
		panic("cannot open database")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// Prepare statement for inserting a folder
	folderIns, err := db.Prepare("INSERT IGNORE INTO Folders(Path, ParentID, Root) VALUES(?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer folderIns.Close()

	// Prepare statement for inserting a folder
	fileIns, err := db.Prepare("INSERT INTO Tracks(Path, FolderID, Title, Album, Artist, Genre, Year) VALUES(?, ?, ?, ?, ?, ?, ?)")
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

			println(file)

			var isRoot = root == file

			var parentID = 0

			if !isRoot {
				println(path.Dir(file))
				err = parentOut.QueryRow(path.Dir(file)).Scan(&parentID)
				if err != nil {
					panic(err.Error())
				}
			}

			// if it's a folder
			if info.IsDir() {
				_, err = folderIns.Exec(file, parentID, isRoot)
				if err != nil {
					panic(err.Error())
				}
			}

			// if it's a file
			if !info.IsDir() {

				// read metadata
				f, _ := os.Open(file)
				m, err := tag.ReadFrom(f)
				if err != nil {
					log.Fatal(err)
				}

				_, err = fileIns.Exec(file, parentID, m.Title(), m.Artist(), m.Album(), m.Genre(), m.Year())
				if err != nil {
					panic(err.Error())
				}
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return
}
