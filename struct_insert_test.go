package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInsertDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup()

		Convey("Given an object to insert", func() {
			dummy := Dummy{
				AText:       "Foo Bar",
				AnotherText: "Baz",
				AnInteger:   1234,
			}

			Convey("Do execute the query and fill the auto key", func() {
				err := db.Insert(&dummy).Do()

				So(err, ShouldBeNil)
				So(dummy.ID, ShouldBeGreaterThan, 0)

				Convey("The data are in the database", func() {
					retrieveddummy := Dummy{}
					db.Select(&retrieveddummy).Where("id = ?", dummy.ID).Do()
					So(retrieveddummy.ID, ShouldEqual, dummy.ID)
					So(retrieveddummy.AText, ShouldEqual, dummy.AText)
					So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
					So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
				})
			})
		})
	})
}
