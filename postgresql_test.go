package godb_test

import (
	"log"
	"os"
	"testing"

	"gitlab.com/samonzeweb/godb"
	"gitlab.com/samonzeweb/godb/adapters/postgresql"

	. "github.com/smartystreets/goconvey/convey"
)

type Dummy struct {
	Id          int    `db:"id,key,auto"`
	Text        string `db:"a_text"`
	AnotherText string `db:"another_text"`
	AnInteger   int    `db:"an_integer"`
}

func (*Dummy) TableName() string {
	return "dummies"
}

// GODB_POSTGRESQL="host=xxxx user=xxxx password=xxxx dbname=xxxx"
// see https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func skipUnlessPostgresql(t *testing.T) {
	if os.Getenv("GODB_POSTGRESQL") == "" {
		t.Skip("Don't run PostgreSQL test, GODB_POSTGRESQL not set")
	}
}

func fixturesSetup(t *testing.T) (*godb.DB, func()) {
	db, err := godb.Open(postgresql.Adapter, os.Getenv("GODB_POSTGRESQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create temporary table if not exists dummies (
		id 						serial primary key,
		a_text     		varchar(255) not null,
		another_text	varchar(255) not null,
		an_integer 		integer not null);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists dummies"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestPostgresql(t *testing.T) {
	skipUnlessPostgresql(t)
	Convey("Given a Postgresql database", t, func() {
		db, teardown := fixturesSetup(t)
		defer teardown()

		Convey("I can insert a struct and get back its new id", func() {
			var dummy = &Dummy{
				Text:        "My text",
				AnotherText: "Other text",
				AnInteger:   123,
			}
			err := db.Insert(dummy).Do()
			So(err, ShouldBeNil)
			So(dummy.Id, ShouldBeGreaterThan, 0)
		})
	})

}
