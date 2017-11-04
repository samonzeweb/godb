package godb_test

import (
	"os"
	"testing"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"github.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

const sqlite3testdb = "sqlite3-test.db"

func fixturesSetupSQLite(t *testing.T) (*godb.DB, func()) {
	removeDBIfExists(t)
	db, err := godb.Open(sqlite.Adapter, sqlite3testdb)
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create table books (
		id 						integer not null primary key autoincrement,
		title     		text not null,
		author    	  text not null,
        published			date not null,
		version       int not null default 0);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists books"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
		removeDBIfExists(t)
	}

	return db, fixturesTeardown
}

func removeDBIfExists(t *testing.T) {
	_, err := os.Stat(sqlite3testdb)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		// no db file
		return
	}

	err = os.Remove(sqlite3testdb)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStatementsSQLite(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db, teardown := fixturesSetupSQLite(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsSQLite(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db, teardown := fixturesSetupSQLite(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}
