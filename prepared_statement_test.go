package godb

import (
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetQueryable(t *testing.T) {
	Convey("Given a connection to a database", t, func() {
		db := fixturesSetup(t)
		sqlQuery := "SELECT * FROM dummies"

		Convey("getQueryable returns a wrapper if there is no Tx", func() {
			q, err := db.getQueryable(sqlQuery)
			So(err, ShouldBeNil)
			So(q, ShouldHaveSameTypeAs, &queryable{})
		})

		Convey("getQueryable returns a prepared statement during a Tx", func() {
			db.Begin()
			q, err := db.getQueryable(sqlQuery)
			So(err, ShouldBeNil)
			So(q, ShouldHaveSameTypeAs, &sql.Stmt{})

			Convey("getQueryable returns cached prepared statements", func() {
				q2, err := db.getQueryable(sqlQuery)
				So(err, ShouldBeNil)
				So(q2, ShouldEqual, q)
			})
		})

	})
}
