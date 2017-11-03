package godb_test

import (
	"os"
	"testing"
	"time"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/postgresql"
	"github.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetupPostgreSQL(t *testing.T) (*godb.DB, func()) {
	if os.Getenv("GODB_POSTGRESQL") == "" {
		t.Skip("Don't run PostgreSQL test, GODB_POSTGRESQL not set")
	}

	db, err := godb.Open(postgresql.Adapter, os.Getenv("GODB_POSTGRESQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create temporary table if not exists books (
		id 						serial primary key,
		title     		varchar(128) not null,
		author    	  varchar(128) not null,
		published			date not null,
		version       int not null default 0);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists books"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestStatementsPostgreSQL(t *testing.T) {
	Convey("A DB for a PostgreSQL database", t, func() {
		db, teardown := fixturesSetupPostgreSQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsPostgreSQL(t *testing.T) {
	Convey("A DB for a PostgreSQL database", t, func() {
		db, teardown := fixturesSetupPostgreSQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}

type bookWithXmin struct {
	Id        int       `db:"id,key,auto"`
	Title     string    `db:"title"`
	Author    string    `db:"author"`
	Published time.Time `db:"published"`
	Version   int       `db:"version"`
	Xmin      int       `db:"xmin,auto,oplock"`
}

func (bookWithXmin) TableName() string {
	return "books"
}
func TestReturningClausePostgreSQL(t *testing.T) {
	Convey("A DB for a PostgreSQL database", t, func() {
		db, teardown := fixturesSetupPostgreSQL(t)
		defer teardown()

		Convey("The auto columns are set after a struct insert", func() {
			book := bookWithXmin{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
				Version:   1,
			}
			db.Insert(&book).Do()
			So(book.Id, ShouldBeGreaterThan, 0)
			So(book.Xmin, ShouldBeGreaterThan, 0)
		})

		Convey("The auto columns are updates after a struct update", func() {
			book := bookWithXmin{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
				Version:   1,
			}
			db.Insert(&book).Do()
			previousXmin := book.Xmin
			db.Update(&book).Do()
			So(book.Xmin, ShouldBeGreaterThan, previousXmin)
		})
	})
}

func TestAutomaticOptimisticLockingPostgreSQL(t *testing.T) {
	Convey("A DB for a PostgreSQL database", t, func() {
		db, teardown := fixturesSetupPostgreSQL(t)
		defer teardown()

		Convey("A row inserted in database", func() {
			book := bookWithXmin{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
				Version:   1,
			}
			db.Insert(&book).Do()

			Convey("Another player read the row", func() {
				var sameBook bookWithXmin
				db.Select(&sameBook).Where("id = ? ", book.Id).Do()

				Convey("Both instances has same value for the oplock field", func() {
					So(book.Xmin, ShouldEqual, sameBook.Xmin)

					Convey("Both players update the row, but the second fails because oplock", func() {
						err := db.Update(&book).Do()
						So(err, ShouldBeNil)
						err = db.Update(&sameBook).Do()
						So(err, ShouldEqual, godb.ErrOpLock)
					})
				})
			})
		})

	})
}
