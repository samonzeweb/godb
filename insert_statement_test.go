package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInsertInto(t *testing.T) {
	Convey("TestInsertInto creates an insert into statement", t, func() {
		db := &DB{}
		q := db.InsertInto("dummies")
		So(q.intoTable, ShouldEqual, "dummies")
	})
}

func TestInsertColumns(t *testing.T) {
	Convey("Given an insert statement", t, func() {
		db := &DB{}
		q := db.InsertInto("dummies")

		Convey("Columns add columns after the existing list", func() {
			q.Columns("foo")
			q.Columns("bar", "baz")
			So(len(q.columns), ShouldEqual, 3)
			So(q.columns[0], ShouldEqual, "foo")
			So(q.columns[1], ShouldEqual, "bar")
		})
	})
}

func TestInsertSuffix(t *testing.T) {
	Convey("Given an insert statement", t, func() {
		db := &DB{}
		q := db.InsertInto("dummies")

		Convey("Calling Suffix will add the given string to the suffixes list", func() {
			suffix := "RETURNING id"
			q.Suffix(suffix)
			So(len(q.suffixes), ShouldEqual, 1)
			So(q.suffixes[0], ShouldEqual, suffix)
		})
	})
}

func TestInsertToSQL(t *testing.T) {
	Convey("Given a valid insert statement with table, columns and values", t, func() {
		db := &DB{}
		q := db.InsertInto("dummies")
		q.Columns("foo", "bar", "baz")
		q.Values(1, 2, 3)

		Convey("ToSQL create a SQL request", func() {
			sql, _, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "INSERT INTO dummies (foo, bar, baz) VALUES (?, ?, ?)")
		})

		Convey("Calling Values multiple times create a SQL with more values", func() {
			q.Values(4, 5, 6)
			sql, _, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "INSERT INTO dummies (foo, bar, baz) VALUES (?, ?, ?), (?, ?, ?)")
		})

		Convey("Given values are in returned arguments", func() {
			q.Values(4, 5, 6)
			_, args, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(len(args), ShouldEqual, 6)
		})

		Convey("Calling Suffix will add the given clause to SQL", func() {
			q.Suffix("RETURNING id")
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, " RETURNING id")
		})

	})
}

func TestInsertToSQLErrors(t *testing.T) {
	db := &DB{}

	Convey("Table name is mandatory", t, func() {
		q := db.InsertInto("").Columns("foo").Values(1, 2, 3)
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})

	Convey("Columns are mandatory", t, func() {
		q := db.InsertInto("dummies").Values(1, 2, 3)
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})

	Convey("Values are mandatory", t, func() {
		q := db.InsertInto("dummies").Columns("foo")
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})

	Convey("The count of values have to match the columns count", t, func() {
		q := db.InsertInto("dummies").
			Columns("foo", "bar", "baz").
			Values(1, 2)
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})
}

func TestDoInsert(t *testing.T) {
	Convey("Given a test database", t, func() {
		db := fixturesSetup(t)

		Convey("Do execute the query and return the Id", func() {
			lastID, err := db.InsertInto("dummies").
				Columns("a_text", "another_text", "an_integer").
				Values("Foo", "Bar", 123).Do()
			So(err, ShouldBeNil)
			So(lastID, ShouldBeGreaterThan, 0)

			Convey("The data are in the database", func() {
				dummy := Dummy{}
				db.Select(&dummy).Where("id = ?", lastID).Do()
				So(dummy.ID, ShouldEqual, lastID)
			})
		})
	})
}
