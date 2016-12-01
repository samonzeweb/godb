package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDelete(t *testing.T) {
	Convey("Create a delete statement", t, func() {
		db := &DB{}
		Convey("With table name", func() {
			q := db.DeleteFrom("dummies")
			Convey("The table name is defined", func() {
				So(q.fromTable, ShouldEqual, "dummies")
			})
		})
	})
}

func TestDeleteWhere(t *testing.T) {
	Convey("Given a delete statement", t, func() {
		db := &DB{}
		q := db.DeleteFrom("dummies")

		Convey("Call Where will add a new condition", func() {
			sql := "id = ?"
			q.Where(sql, 123)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0].sql, ShouldEqual, sql)
		})
	})
}

func TestDeleteWhereQ(t *testing.T) {
	Convey("Given a delete statement", t, func() {
		db := &DB{}
		q := db.DeleteFrom("dummies")

		Convey("Call WhereQ will add the given condition", func() {
			qc := Q("id = ?", 123)
			q.WhereQ(qc)
			So(len(q.where), ShouldEqual, 1)
			So(q.where[0], ShouldEqual, qc)
		})
	})
}

func TestDeleteSuffix(t *testing.T) {
	Convey("Given a delete statement", t, func() {
		db := &DB{}
		q := db.DeleteFrom("dummies")

		Convey("Calling Suffix will add the given string to the suffixes list", func() {
			suffix := "RETURNING id"
			q.Suffix(suffix)
			So(len(q.suffixes), ShouldEqual, 1)
			So(q.suffixes[0], ShouldEqual, suffix)
		})
	})
}

func TestDeleteToSQL(t *testing.T) {
	Convey("Given a valid delete statement", t, func() {
		db := &DB{}
		q := db.DeleteFrom("dummies")

		Convey("ToSQL create a SQL request", func() {
			sql, _, err := q.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "DELETE FROM dummies")
		})

		Convey("Calling Where multiple times", func() {
			q.Where("id = ?", 123).Where("is_deleted = ?", 0)
			sql, args, _ := q.ToSQL()

			Convey("will add all the specified conditions clause to SQL using the 'AND' conjunction", func() {
				So(sql, ShouldEndWith, "WHERE id = ? AND is_deleted = ?")
			})

			Convey("will add given arguments in the correct order", func() {
				So(len(args), ShouldEqual, 2)
				So(args[0].(int), ShouldEqual, 123)
				So(args[1].(int), ShouldEqual, 0)
			})
		})

		Convey("Calling Suffix will add the given clause to SQL", func() {
			q.Suffix("RETURNING id")
			sql, _, _ := q.ToSQL()
			So(sql, ShouldEndWith, " RETURNING id")
		})

	})
}

func TestDeleteToSQLErrors(t *testing.T) {
	Convey("Table name is mandatory", t, func() {
		db := &DB{}
		q := db.DeleteFrom("")
		_, _, err := q.ToSQL()
		So(err, ShouldNotBeNil)
	})
}
