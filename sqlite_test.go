package godb_test

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

		create table inventories (
		id             integer not null primary key autoincrement,
		book_id			   int not null,
		last_inventory date not null,
		counting       int not null default 0);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists books; drop table if exists inventories;"
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

		Convey("The common statements tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsSQLite(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db, teardown := fixturesSetupSQLite(t)
		defer teardown()

		Convey("The common structs tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}
func TestRawSQLite(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db, teardown := fixturesSetupSQLite(t)
		defer teardown()

		Convey("The common raw tests must pass", func() {
			common.RawSQLTests(db, t)
		})
	})
}
