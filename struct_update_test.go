package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)

		Convey("Update update a record", func() {
			dummy := &Dummy{}
			err := db.Select(dummy).Where("an_integer = ?", 11).Do()
			So(err, ShouldBeNil)
			dummy.AText = "New text"
			dummy.AnotherText = "Replacement text"
			dummy.AnInteger = 123
			count, err := db.Update(dummy).Do()
			So(err, ShouldBeNil)

			Convey("The data are in the database", func() {
				retrieveddummy := Dummy{}
				db.Select(&retrieveddummy).Where("id = ?", dummy.ID).Do()
				So(retrieveddummy.ID, ShouldEqual, dummy.ID)
				So(retrieveddummy.AText, ShouldEqual, dummy.AText)
				So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
				So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
			})

			Convey("Update returns the count of affected rows", func() {
				So(count, ShouldEqual, 1)
			})
		})

		Convey("Update returns error if optimistic locking fails", func() {
			Convey("With non auto oplock field", func() {
				dummy := &Dummy{}
				err := db.Select(dummy).Where("an_integer = ?", 11).Do()
				So(err, ShouldBeNil)

				// Simulate another update of the record.
				_, err = db.UpdateTable("dummies").Set("version", 1).Where("id = ?", dummy.ID).Do()
				So(err, ShouldBeNil)

				count, err := db.Update(dummy).Do()
				So(count, ShouldEqual, 0)
				So(err, ShouldEqual, ErrOpLock)
			})

			Convey("With auto oplock field", func() {
				dummy := &DummyAutoOplock{}
				err := db.Select(dummy).Where("an_integer = ?", 11).Do()
				So(err, ShouldBeNil)

				// Simulate another update of the record.
				_, err = db.UpdateTable("dummiesautooplock").Set("version", 1).Where("id = ?", dummy.ID).Do()
				So(err, ShouldBeNil)

				count, err := db.Update(dummy).Do()
				So(count, ShouldEqual, 0)
				So(err, ShouldEqual, ErrOpLock)
			})
		})
	})
}
