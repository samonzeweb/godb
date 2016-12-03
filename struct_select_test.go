package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSelectDoWithStruct(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)

		Convey("Do execute the query and fills a given instance", func() {
			singleDummy := Dummy{}
			selectStmt := db.Select(&singleDummy).
				Where("an_integer = ?", 13)

			err := selectStmt.Do()
			So(err, ShouldBeNil)
			So(singleDummy.ID, ShouldBeGreaterThan, 0)
			So(singleDummy.AText, ShouldEqual, "Third")
			So(singleDummy.AnotherText, ShouldEqual, "Troisième")
			So(singleDummy.AnInteger, ShouldEqual, 13)
		})

		Convey("Do execute the query and fills a slice", func() {
			dummiesSlice := make([]Dummy, 0, 0)
			selectStmt := db.Select(&dummiesSlice).
				OrderBy("an_integer")

			err := selectStmt.Do()
			So(err, ShouldBeNil)
			So(len(dummiesSlice), ShouldEqual, 3)
			So(dummiesSlice[0].ID, ShouldBeGreaterThan, 0)
			So(dummiesSlice[0].AText, ShouldEqual, "First")
			So(dummiesSlice[0].AnotherText, ShouldEqual, "Premier")
			So(dummiesSlice[0].AnInteger, ShouldEqual, 11)
			So(dummiesSlice[1].AnInteger, ShouldEqual, 12)
			So(dummiesSlice[2].AnInteger, ShouldEqual, 13)
		})

		Convey("Do execute the query and fills a slice of pointers", func() {
			dummiesSlice := make([]*Dummy, 0, 0)
			selectStmt := db.SelectFrom("dummies").
				Columns("id", "a_text", "another_text", "an_integer").
				OrderBy("an_integer")

			err := selectStmt.Do(&dummiesSlice)
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

func TestCountWithStruct(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)

		Convey("Count returns the count of row mathing the request", func() {
			selectStmt := db.Select(&Dummy{})
			count, err := selectStmt.Count()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 3)

			selectStmt = db.Select(&Dummy{}).Where("an_integer = ?", 12)
			count, err = selectStmt.Count()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)

			Convey("Do compute time consumed by SQL query", func() {
				So(db.ConsumedTime(), ShouldBeGreaterThan, 0)
			})
		})
	})
}
