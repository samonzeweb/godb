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

		testCache := func(cache *StmtCache) {
			Convey("getQueryable returns a wrapper if the cache is disabled", func() {
				cache.Disable()
				q, err := db.getQueryable(sqlQuery)
				So(err, ShouldBeNil)
				So(q, ShouldHaveSameTypeAs, &queryWrapper{})
			})

			Convey("getQueryable returns a prepared statement if the cache is enabled", func() {
				cache.Enable()
				q, err := db.getQueryable(sqlQuery)
				So(err, ShouldBeNil)
				So(q, ShouldHaveSameTypeAs, &sql.Stmt{})

				Convey("getQueryable returns cached prepared statements", func() {
					q2, err := db.getQueryable(sqlQuery)
					So(err, ShouldBeNil)
					So(q2, ShouldEqual, q)
				})

				Convey("getQueryable returns a new prepared statements after a clear cache", func() {
					cache.Clear()
					q2, err := db.getQueryable(sqlQuery)
					So(err, ShouldBeNil)
					So(q2, ShouldNotEqual, q)
				})
			})
		}

		Convey("Without Tx", func() {
			testCache(db.StmtCacheDB())
		})

		Convey("With Tx", func() {
			db.Begin()
			testCache(db.StmtCacheTx())
			db.Rollback()
		})
	})
}
