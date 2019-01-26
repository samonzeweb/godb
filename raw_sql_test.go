package godb

import (
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRawSQLDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Do execute the raw query and fills a given instance", func() {
			singleDummy := Dummy{}
			err := db.RawSQL("select * from dummies where an_integer = ?", 12).Do(&singleDummy)
			So(err, ShouldBeNil)
			So(singleDummy.ID, ShouldBeGreaterThan, 0)
			So(singleDummy.AText, ShouldEqual, "Second")
			So(singleDummy.AnotherText, ShouldEqual, "Second")
			So(singleDummy.AnInteger, ShouldEqual, 12)
		})

		Convey("Do returns sql.sql.ErrNoRows if a single instance if requested but not found", func() {
			dummy := Dummy{}
			err := db.RawSQL("select * from dummies where an_integer = 123").Do(&dummy)
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("Do execute the query and fills a slice", func() {
			dummiesSlice := make([]Dummy, 0)
			err := db.RawSQL("select * from dummies").Do(&dummiesSlice)
			So(err, ShouldBeNil)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].ID, ShouldBeGreaterThan, 0)
			So(dummiesSlice[0].AText, ShouldEqual, "First")
			So(dummiesSlice[0].AnotherText, ShouldEqual, "Premier")
			So(dummiesSlice[0].AnInteger, ShouldEqual, 11)
			So(dummiesSlice[1].AnInteger, ShouldEqual, 12)
			So(dummiesSlice[2].AnInteger, ShouldEqual, 13)
		})
	})
}

func TestRawSQLDoWithIterator(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("DoWithIterator executes the query and returns an iterator", func() {
			iter, err := db.RawSQL("select * from dummies order by an_integer").DoWithIterator()
			So(err, ShouldBeNil)
			defer iter.Close()

			count := 0
			for iter.Next() {
				count++
				singleDummy := Dummy{}
				err := iter.Scan(&singleDummy)
				So(err, ShouldBeNil)

				if count == 1 {
					So(singleDummy.ID, ShouldBeGreaterThan, 0)
					So(singleDummy.AText, ShouldEqual, "First")
					So(singleDummy.AnotherText, ShouldEqual, "Premier")
					So(singleDummy.AnInteger, ShouldEqual, 11)
				}

			}
			So(count, ShouldEqual, 3)
			So(iter.Err(), ShouldBeNil)
		})
	})
}
