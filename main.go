package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// dbSchema contains SQL statements needed to initialize database on first start
//go:embed schema.sql
var dbSchema string

// osaPolicy contains permission policies defined in external file
//go:embed authorization.polar
var osoPolicy string

func main() {
	// prepare OSO
	authManager, err := NewAuthorizer(osoPolicy)
	if err != nil {
		panic(err)
	}

	// prepare DB
	dbName := os.Getenv("EXPENSES_DB")
	if dbName == "" {
		dbName = "expenses.sqlite"
	}
	db, err := NewDBManager(dbName)
	if err != nil {
		panic(err)
	}
	if err := db.rawExec(dbSchema); err != nil {
		panic(err)
	}

	// prepare HTTP server
	webApp := NewHTTPHandler(db, authManager)

	// run server
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", webApp))

}
