package sqlite_test

import (
	"testing"

	"gitlab.com/samonzeweb/godb"
	"gitlab.com/samonzeweb/godb/adapters/sqlite"
	"gitlab.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetup(t *testing.T) *godb.DB {
	db, err := godb.Open(sqlite.Adapter, ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create table books (
		id 						integer not null primary key autoincrement,
		title     		text not null,
		author    	   text not null);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		panic(err)
	}

	return db
}

func TestSQLite(t *testing.T) {
	Convey("A DB for a SQLite database", t, func() {
		db := fixturesSetup(t)

		Convey("The common tests must pass", func() {
			common.MainTest(db, t)
		})
	})
}
