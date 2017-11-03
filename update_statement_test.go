package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdate(t *testing.T) {
	Convey("Create an update query", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")
		Convey("The table name is set", func() {
			So(q.updateTable, ShouldEqual, "dummies")
		})
	})
}

func TestSet(t *testing.T) {
	Convey("Create an update query", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Add a SET clause", func() {
			q.Set("foo", 1)

			Convey("The set clause is added to the sets list", func() {
				So(len(q.sets), ShouldEqual, 1)
				So(q.sets[0].column, ShouldEqual, "foo")
				So(q.sets[0].value.(int), ShouldEqual, 1)
			})
		})
	})
}

func TestSetRaw(t *testing.T) {
	Convey("Create an update query", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Add a raw SET clause", func() {
			sql := "foo = foo+1"
			q.SetRaw(sql)

			Convey("The set clause is added to the sets list", func() {
				So(len(q.sets), ShouldEqual, 1)
				So(q.sets[0].column, ShouldEqual, sql)
				So(q.sets[0].value, ShouldBeNil)
			})
		})
	})
}

func TestUpdateWhere(t *testing.T) {
	Convey("Given an update statement", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Call Where will add a new condition", func() {
			sql := "id = ?"
			q.Where(sql, 123)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0].sql, ShouldEqual, sql)
		})
	})
}

func TestUpdateWhereQ(t *testing.T) {
	Convey("Given an update statement", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Call WhereQ will add the given condition", func() {
			qc := Q("id = ?", 123)
			q.WhereQ(qc)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0], ShouldEqual, qc)
		})
	})
}

func TestUpdateSuffix(t *testing.T) {
	Convey("Given an update statement", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Calling Suffix will add the given string to the suffixes list", func() {
			suffix := "RETURNING foo"
			q.Suffix(suffix)
			So(len(q.suffixes), ShouldEqual, 1)
			So(q.suffixes[0], ShouldEqual, suffix)
		})
	})
}

func TestUpdateToSQL(t *testing.T) {
	Convey("Given a valid update statement", t, func() {
		db := &DB{}
		q := db.UpdateTable("dummies")

		Convey("Calling Set add the SET clause to SQL", func() {
			q.Set("foo", 1)
			sql, args, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "UPDATE dummies SET foo=?")
			So(len(args), ShouldEqual, 1)
			So(args[0], ShouldEqual, 1)

			Convey("Calling SetRaw add the SET clause to SQL", func() {
				rawSet := "bar = bar + 1"
				q.SetRaw(rawSet)
				sql, _, err = q.ToSQL()
				So(err, ShouldBeNil)
				So(sql, ShouldContainSubstring, rawSet)
			})

			Convey("Calling Where add a condition", func() {
				q.Where("id = ?", 123)
				sql, _, err = q.ToSQL()
				So(err, ShouldBeNil)
				So(sql, ShouldContainSubstring, "WHERE id = ?")
			})

			Convey("Calling Suffix will add the given clause to SQL", func() {
				q.Suffix("RETURNING bar")
				sql, _, _ := q.ToSQL()
				So(sql, ShouldEndWith, " RETURNING bar")
			})
		})
	})
}

func TestUpdateToSQLErrors(t *testing.T) {
	Convey("Table name is mandatory", t, func() {
		db := &DB{}
		q := db.UpdateTable("")
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})
}

func TestDoUpdate(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Do execute the query and return the count of affected rows", func() {
			rowsAffected, err := db.UpdateTable("dummies").
				Set("another_text", "New text").
				Where("an_integer >= ?", 12).
				Do()

			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			Convey("The database is up-to-date", func() {
				dummies := make([]Dummy, 0, 0)
				_ = db.Select(&dummies).Where("an_integer >= ?", 12).Do()
				for _, dummy := range dummies {
					So(dummy.AnotherText, ShouldEqual, "New text")
				}
			})
		})
	})
}
