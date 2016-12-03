package godb

import (
	"log"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClone(t *testing.T) {
	Convey("Given an existing DB", t, func() {
		db := createInMemoryConnection(t)
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
