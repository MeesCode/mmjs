package database

import (
	"database/sql"
	"mp3bak2/globals"
)

func connect() *sql.DB {
	db, err := sql.Open("mysql",
		globals.DatabaseCredentials.Username+":"+
			globals.DatabaseCredentials.Password+"@/"+
			globals.DatabaseCredentials.Database)

	if err != nil {
		panic("cannot open database")
	}

	err = db.Ping()
	if err != nil {
		panic("connection with database could not be established")
	}

	return db
}
