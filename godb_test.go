package godb

import (
	"log"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/samonzeweb/godb/adapters/sqlite"
)

func TestClone(t *testing.T) {
	Convey("Given an existing DB", t, func() {
		db := createInMemoryConnection()
		db.SetLogger(log.New(os.Stderr, "", 0))

		Convey("Clone create a DB copy of an existing one", func() {
			clone := db.Clone()
			So(clone.adapter, ShouldHaveSameTypeAs, db.adapter)
			So(clone.sqlDB, ShouldEqual, db.sqlDB)
			So(clone.logger, ShouldEqual, db.logger)
		})

		Convey("Clone don't copy existing transaction", func() {
			db.Begin()
			clone := db.Clone()
			So(clone.sqlTx, ShouldBeNil)
		})
	})
}

func checkToSQL(t *testing.T, sqlExpected string, sqlProduced string, err error) {
	if err != nil {
		t.Fatal("ToSQL produces error :", err)
	}

	t.Log("SQL expected :", sqlExpected)
	t.Log("SQL produced :", sqlProduced)
	if sqlProduced != sqlExpected {
		t.Fatal("ToSQL produces incorrect SQL")
	}
}

func createInMemoryConnection() *DB {
	db, err := Open(sqlite.Adapter, ":memory:")
	if err != nil {
		panic(err)
	}

	return db
}
