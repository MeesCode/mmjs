package database

import (
	"database/sql"
	"mp3bak2/globals"
)

var dbc *sql.DB

func Warmup() *sql.DB {
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

	dbc = db
	return dbc
}

func getConnection() *sql.DB {
	return dbc
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
