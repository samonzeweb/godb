package godb

import (
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInsertDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

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

		Convey("Given an object to insert with whitelist/blacklist", func() {
			dummy := Dummy{
				AText:           "Foo Bar2",
				AnotherText:     "Baz2",
				AnInteger:       12345,
				ANullableString: sql.NullString{String: "Void", Valid: true},
			}

			Convey("Do execute the query whitelisted after reset", func() {
				insQ := db.Insert(&dummy)
				insQ.Whitelist("a_nullable_string")
				insQ.WhitelistReset()
				err := insQ.Whitelist("a_text", "another_text", "an_integer").Do()

				So(err, ShouldBeNil)
				So(dummy.ID, ShouldBeGreaterThan, 0)

				Convey("The data are in the database", func() {
					retrieveddummy := Dummy{}
					db.Select(&retrieveddummy).Where("id = ?", dummy.ID).Do()
					So(retrieveddummy.ID, ShouldEqual, dummy.ID)
					So(retrieveddummy.AText, ShouldEqual, dummy.AText)
					So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
					So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
					So(retrieveddummy.ANullableString.Valid, ShouldEqual, false)
				})
			})
			Convey("Do execute the query whitelisted with id", func() {
				customID := 1453
				dummy.ID = customID
				err := db.Insert(&dummy).Whitelist("id", "an_integer", "a_text", "another_text").Do()

				So(err, ShouldBeNil)
				So(dummy.ID, ShouldEqual, customID)

				Convey("The data are in the database with custom key", func() {
					retrieveddummy := Dummy{}
					db.Select(&retrieveddummy).Where("id = ?", customID).Do()
					So(retrieveddummy.ID, ShouldEqual, dummy.ID)
					So(retrieveddummy.AText, ShouldEqual, dummy.AText)
					So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
					So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
					So(retrieveddummy.ANullableString.Valid, ShouldEqual, false)
				})
			})

			Convey("Do execute the query whitlisted mixed order", func() {
				err := db.Insert(&dummy).Whitelist("another_text", "an_integer", "a_text").Do()

				So(err, ShouldBeNil)
				So(dummy.ID, ShouldBeGreaterThan, 0)

				Convey("The data are in the database", func() {
					retrieveddummy := Dummy{}
					db.Select(&retrieveddummy).Where("id = ?", dummy.ID).Do()
					So(retrieveddummy.ID, ShouldEqual, dummy.ID)
					So(retrieveddummy.AText, ShouldEqual, dummy.AText)
					So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
					So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
					So(retrieveddummy.ANullableString.Valid, ShouldEqual, false)
				})
			})
			Convey("Do execute the query blacklist after reset", func() {
				insQ := db.Insert(&dummy)
				insQ.Blacklist("a_string")
				insQ.BlacklistReset()
				err := insQ.Blacklist("a_nullable_string").Do()

				So(err, ShouldBeNil)
				So(dummy.ID, ShouldBeGreaterThan, 0)

				Convey("The data are in the database", func() {
					retrieveddummy := Dummy{}
					db.Select(&retrieveddummy).Where("id = ?", dummy.ID).Do()
					So(retrieveddummy.ID, ShouldEqual, dummy.ID)
					So(retrieveddummy.AText, ShouldEqual, dummy.AText)
					So(retrieveddummy.AnotherText, ShouldEqual, dummy.AnotherText)
					So(retrieveddummy.AnInteger, ShouldEqual, dummy.AnInteger)
					So(retrieveddummy.ANullableString.Valid, ShouldEqual, false)
				})
			})
		})
	})
}

func TestBulkInsertDo(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Given a slice of objects to insert", func() {
			slice := make([]Dummy, 0)
			for i := 1; i <= 10; i++ {
				dummy := Dummy{
					AText:       "Bulk",
					AnotherText: "Insert",
					AnInteger:   i * 100,
				}
				slice = append(slice, dummy)
			}

			Convey("Do execute the query", func() {
				err := db.BulkInsert(&slice).Do()
				So(err, ShouldBeNil)

				Convey("The data are in the database", func() {
					retrieveddummies := make([]Dummy, 0)
					db.Select(&retrieveddummies).
						Where("an_integer > 99").
						Where("a_text = ?", "Bulk").
						Do()
					So(len(retrieveddummies), ShouldEqual, 10)
				})
			})

			Convey("Do executes the query using a blacklist", func() {
				dummies := make([]Dummy, 0, 2)
				dummy1 := Dummy{
					AText:           "Dummy1",
					AnotherText:     "Foo",
					AnInteger:       12345,
					ANullableString: sql.NullString{String: "Void", Valid: true},
				}
				dummy2 := Dummy{
					AText:           "Dummy2",
					AnotherText:     "Foo",
					AnInteger:       12345,
					ANullableString: sql.NullString{String: "Void", Valid: true},
				}
				dummies = append(dummies, dummy1, dummy2)
				err := db.BulkInsert(&dummies).Blacklist("a_nullable_string", "version").Do()
				So(err, ShouldBeNil)
			})

			Convey("Do executes the query using a whitelist", func() {
				dummies := make([]Dummy, 0, 2)
				dummy1 := Dummy{
					AText:           "Dummy1",
					AnotherText:     "Foo",
					AnInteger:       12345,
					ANullableString: sql.NullString{String: "Void", Valid: true},
				}
				dummy2 := Dummy{
					AText:           "Dummy2",
					AnotherText:     "Foo",
					AnInteger:       12345,
					ANullableString: sql.NullString{String: "Void", Valid: true},
				}
				dummies = append(dummies, dummy1, dummy2)
				err := db.BulkInsert(&dummies).Whitelist("an_integer", "a_text", "another_text").Do()
				So(err, ShouldBeNil)
			})

		})
	})

}
