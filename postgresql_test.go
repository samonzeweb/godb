package godb_test

import (
	"os"
	"testing"

	"gitlab.com/samonzeweb/godb"
	"gitlab.com/samonzeweb/godb/adapters/postgresql"

	. "github.com/smartystreets/goconvey/convey"
)

// GODB_POSTGRESQL="host=xxxx user=xxxx password=xxxx dbname=xxxx"
// see https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func skipUnlessPostgresql(t *testing.T) {
	if os.Getenv("GODB_POSTGRESQL") == "" {
		t.Skip("Don't run PostgreSQL test, GODB_POSTGRESQL not set")
	}
}

func fixturesSetup() (*godb.DB, func()) {
	db, err := godb.Open(postgresql.Adapter, os.Getenv("GODB_POSTGRESQL"))
	if err != nil {
		panic(err)
	}

	createTable :=
		`create temporary table if not exists dummies (
		id 						serial primary key,
		a_text     		varchar(255) not null,
		another_text	varchar(255) not null,
		an_integer 		integer not null);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		panic(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists dummies"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			panic(err)
		}
	}

	return db, fixturesTeardown
}
func TestPostgresql(t *testing.T) {
	skipUnlessPostgresql(t)
	Convey("Given a Postgresql database", t, func() {
		db, teardown := fixturesSetup()
		defer teardown()

		//TODO
		_ = db

	})

}
