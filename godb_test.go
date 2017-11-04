package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClone(t *testing.T) {
	Convey("Given an existing DB", t, func() {
		db := createInMemoryConnection(t)
		defer db.Close()

		Convey("Clone create a DB copy of an existing one", func() {
			clone := db.Clone()

			So(clone.adapter, ShouldHaveSameTypeAs, db.adapter)
			So(clone.sqlDB, ShouldEqual, db.sqlDB)
			So(clone.logger, ShouldEqual, db.logger)
		})

		Convey("Clone don't copy existing transaction", func() {
			db.Begin()
			clone := db.Clone()
			defer clone.Clear()
			So(clone.sqlTx, ShouldBeNil)
		})
	})
}
