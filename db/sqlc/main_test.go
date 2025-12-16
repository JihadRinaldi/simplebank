package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/JihadRinaldi/simplebank/util"
	_ "github.com/lib/pq"
)

var testStore Store
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalln("cannot connect to db:", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalln("cannot ping db:", err)
	}

	testStore = NewStore(testDB)
	testQueries = New(testDB)

	code := m.Run()

	_ = testDB.Close()
	os.Exit(code)
}
