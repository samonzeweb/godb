package godb_test

import (
	"os"
	"testing"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/mysql"
	"github.com/samonzeweb/godb/dbtests/common"

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
		`create temporary table if not exists books (
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
	}

	return db, fixturesTeardown
}

func TestStatementsMySQL(t *testing.T) {
	Convey("A DB for a MySQL database", t, func() {
		db, teardown := fixturesSetupMySQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsMySQL(t *testing.T) {
	Convey("A DB for a MySQL database", t, func() {
		db, teardown := fixturesSetupMySQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}
