package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testStore Store
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalln("cannot connect to db:", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalln("cannot ping db:", err)
	}

	testStore = *NewStore(testDB)
	testQueries = New(testDB)

	code := m.Run()

	_ = testDB.Close()
	os.Exit(code)
}

// func TestMain(m *testing.M) {
// 	conn, err := sql.Open(dbDriver, dbSource)
// 	if err != nil {
// 		log.Fatalln("cannot connect to db:", err)
// 	}

// 	testQueries = New(conn)

// 	os.Exit(m.Run())
// }
