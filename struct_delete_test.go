package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)

		Convey("Delete delete a record", func() {
			dummy := &Dummy{}
			err := db.Select(dummy).Where("an_integer = ?", 11).Do()
			So(err, ShouldBeNil)
			count, err := db.Delete(dummy).Do()
			So(err, ShouldBeNil)

			Convey("The data isn't in database", func() {
				found, err := db.SelectFrom("dummies").Where("id = ?", dummy.ID).Count()
				So(err, ShouldBeNil)
				So(found, ShouldEqual, 0)
			})

			Convey("Delete returns the count of affected rows", func() {
				So(count, ShouldEqual, 1)
			})
		})
	})
}
