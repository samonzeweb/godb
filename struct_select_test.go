package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDoWithStruct(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup()

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
