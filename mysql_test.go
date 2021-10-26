package godb_test

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/mysql"
	"github.com/samonzeweb/godb/dbtests/common"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetupMySQL(t *testing.T) (*godb.DB, func()) {
	if os.Getenv("GODB_MYSQL") == "" {
		t.Skip("Don't run MySQL test, GODB_MYSQL not set")
	}

	db, err := godb.Open(mysql.Adapter, os.Getenv("GODB_MYSQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create table if not exists books (
		id 						int auto_increment primary key,
		title     		varchar(128) not null,
		author    	  varchar(128) not null,
		published			date not null,
		version       int not null default 0);
		`

	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	createTable = `create table if not exists inventories (
		id             int auto_increment primary key,
		book_id			   int not null,
		last_inventory date not null,
		counting       int not null default 0);
		`

	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists books;"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
		dropTable = "drop table if exists inventories;"
		_, err = db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestStatementsMySQL(t *testing.T) {
	Convey("A DB for a MySQL database", t, func() {
		db, teardown := fixturesSetupMySQL(t)
		defer teardown()

		Convey("The common statements tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsMySQL(t *testing.T) {
	Convey("A DB for a MySQL database", t, func() {
		db, teardown := fixturesSetupMySQL(t)
		defer teardown()

		Convey("The common structs tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}

func TestRawMySQL(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db, teardown := fixturesSetupMySQL(t)
		defer teardown()

		Convey("The common raw tests must pass", func() {
			common.RawSQLTests(db, t)
		})
	})
}
