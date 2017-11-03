package godb

import (
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewCache(t *testing.T) {
	Convey("New create a new cache", t, func() {
		c := newStmtCache()

		Convey("With default settings", func() {
			So(c.IsEnabled(), ShouldBeTrue)
			So(c.GetSize(), ShouldEqual, DefaultStmtCacheSize)
		})

		Convey("With zero as LRU counter", func() {
			So(c.lastUse, ShouldBeZeroValue)
		})

		Convey("With a valid content map", func() {
			So(c.content, ShouldNotBeNil)
		})
	})
}

func TestEnabling(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()

		Convey("Enable enables the cache", func() {
			c.isEnabled = false
			c.Enable()
			So(c.isEnabled, ShouldBeTrue)
			So(c.IsEnabled(), ShouldBeTrue)
		})

		Convey("Disable disables the cache", func() {
			c.isEnabled = true
			c.Disable()
			So(c.isEnabled, ShouldBeFalse)
			So(c.IsEnabled(), ShouldBeFalse)
		})
	})
}

func TestSize(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()

		Convey("SetSize changes the cache size", func() {
			newSize := 2 * DefaultStmtCacheSize
			c.SetSize(newSize)
			So(c.maxSize, ShouldEqual, newSize)
		})

		Convey("SetSize reduces the cache size and close statements if needed", func() {
			db := fixturesSetup(t)
			defer db.Close()

			c.SetSize(4)
			query := "select * from dummies where id="
			for i := 0; i < 4; i++ {
				iQuery := query + strconv.Itoa(i)
				stmt, _ := db.CurrentDB().Prepare(iQuery)
				c.add(iQuery, stmt)
			}
			c.SetSize(2)
			So(len(c.content), ShouldEqual, 2)
		})

		Convey("GetSize returns the cache size", func() {
			So(c.GetSize(), ShouldEqual, c.maxSize)
		})
	})
}

func TestAdd(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()
		db := fixturesSetup(t)
		defer db.Close()

		Convey("add adds a stmt into the cache", func() {
			query := "select * from dummies"
			stmt, _ := db.CurrentDB().Prepare(query)
			c.add(query, stmt)
			So(len(c.content), ShouldEqual, 1)
		})

		Convey("add keeps cache size at it's max allowed value", func() {
			c.SetSize(2)
			query := "select * from dummies where id="
			for i := 1; i <= 3; i++ {
				iQuery := query + strconv.Itoa(i)
				stmt, _ := db.CurrentDB().Prepare(iQuery)
				c.add(iQuery, stmt)
			}
			So(len(c.content), ShouldEqual, 2)
		})
	})
}

func TestRemoveLeastRecentlyUsed(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()
		db := fixturesSetup(t)
		defer db.Close()

		Convey("removeLeastRecentlyUsed remove the least recently used item from the cache", func() {
			query := "select * from dummies where id="
			// add fixtures
			for i := 0; i < 10; i++ {
				iQuery := query + strconv.Itoa(i)
				stmt, _ := db.CurrentDB().Prepare(iQuery)
				c.add(iQuery, stmt)
			}
			// ensure stmt with 'id=3' is the looser
			for i := 0; i < 10; i++ {
				if i != 3 {
					iQuery := query + strconv.Itoa(i)
					_ = c.get(iQuery)
				}
			}

			c.removeLeastRecentlyUsed()
			removedQuery := query + strconv.Itoa(3)
			So(c.get(removedQuery), ShouldBeNil)
		})
	})
}

func TestGet(t *testing.T) {

	Convey("Given a cache", t, func() {
		c := newStmtCache()
		db := fixturesSetup(t)
		defer db.Close()

		Convey("get returns the stmt from the cache corresponding to the given query", func() {
			query := "select * from dummies"
			stmt, _ := db.CurrentDB().Prepare(query)
			c.add(query, stmt)
			So(c.get(query), ShouldEqual, stmt)
		})

		Convey("get returns nil if the query was not found", func() {
			query := "select * from dummies"
			So(c.get(query), ShouldBeNil)
		})
	})
}

func TestClear(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Clears close the stmt and remove all entries from the cache", func() {
			query := "select * from dummies"
			stmt, _ := db.CurrentDB().Prepare(query)
			c.add(query, stmt)
			c.Clear()
			So(len(c.content), ShouldEqual, 0)
			_, err := stmt.Query()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestClearWithoutClosingStmt(t *testing.T) {
	Convey("Given a cache", t, func() {
		c := newStmtCache()
		db := fixturesSetup(t)
		defer db.Close()

		Convey("Clears remove all entries from the cache but does not close stmt", func() {
			query := "select * from dummies"
			stmt, _ := db.CurrentDB().Prepare(query)
			c.add(query, stmt)
			c.clearWithoutClosingStmt()
			So(len(c.content), ShouldEqual, 0)
			_, err := stmt.Query()
			So(err, ShouldBeNil)
		})
	})
}
