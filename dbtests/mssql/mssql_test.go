package mssql_test

import (
	"os"
	"testing"

	"gitlab.com/samonzeweb/godb"
	"gitlab.com/samonzeweb/godb/adapters/mssql"
	"gitlab.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetup(t *testing.T) (*godb.DB, func()) {
	if os.Getenv("GODB_MSSQL") == "" {
		t.Skip("Don't run SQL Server test, GODB_MSSQL not set")
	}

	db, err := godb.Open(mssql.Adapter, os.Getenv("GODB_MSSQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create table books (
		id 						int identity,
		title     		nvarchar(128) not null,
		author    	  nvarchar(128) not null,
		published			datetime2 not null);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table books"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestMySQL(t *testing.T) {
	Convey("A DB for a SQL Server database", t, func() {
		db, teardown := fixturesSetup(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.MainTest(db, t)
		})
	})
}
