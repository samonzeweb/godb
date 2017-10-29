package godb_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/mssql"
	"github.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetupMSSQL(t *testing.T) (*godb.DB, func()) {
	if os.Getenv("GODB_MSSQL") == "" {
		t.Skip("Don't run SQL Server test, GODB_MSSQL not set")
	}

	db, err := godb.Open(mssql.Adapter, os.Getenv("GODB_MSSQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create table books (
		id 						int identity,
		title     		nvarchar(128) not null,
		author    	  nvarchar(128) not null,
		published			datetime2 not null,
		version       int not null default 0);

		create table bookswithrowversion (
			id 						int identity,
			title     		nvarchar(128) not null,
			author    	  nvarchar(128) not null,
			published			datetime2 not null,
			version       rowversion);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table books; drop table bookswithrowversion;"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestStatementsMSSQL(t *testing.T) {
	Convey("A DB for a SQL Server database", t, func() {
		db, teardown := fixturesSetupMSSQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StatementsTests(db, t)
		})
	})
}

func TestStructsMSSQL(t *testing.T) {
	Convey("A DB for a SQL Server database", t, func() {
		db, teardown := fixturesSetupMSSQL(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.StructsTests(db, t)
		})
	})
}

type bookWithRowversion struct {
	Id        int              `db:"id,key,auto"`
	Title     string           `db:"title"`
	Author    string           `db:"author"`
	Published time.Time        `db:"published"`
	Version   mssql.Rowversion `db:"version,auto,oplock"`
}

func (bookWithRowversion) TableName() string {
	return "bookswithrowversion"
}

func TestReturningClauseMSSQL(t *testing.T) {
	Convey("A DB for a SQL Server database", t, func() {
		db, teardown := fixturesSetupMSSQL(t)
		defer teardown()

		Convey("The auto columns are set after a struct insert", func() {
			book := bookWithRowversion{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
			}
			db.Insert(&book).Do()
			So(book.Id, ShouldBeGreaterThan, 0)
			// Expect non zeroes (or empty) slice for Version
			allZeroOrEmpty := true
			for _, v := range book.Version.Version {
				allZeroOrEmpty = allZeroOrEmpty && (v == 0)
			}
			So(allZeroOrEmpty, ShouldBeFalse)
		})

		Convey("The auto columns are updates after a struct update", func() {
			book := bookWithRowversion{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
			}
			db.Insert(&book).Do()
			previousVersion := book.Version.Copy()
			db.Update(&book).Do()
			// Expect differents values in versions
			So(bytes.Compare(book.Version.Version, previousVersion.Version), ShouldNotEqual, 0)
		})
	})
}

func TestAutomaticOptimisticLockingMSSQL(t *testing.T) {
	Convey("A DB for a SQL Server database", t, func() {
		db, teardown := fixturesSetupMSSQL(t)
		defer teardown()

		Convey("A row inserted in database", func() {
			book := bookWithRowversion{
				Title:     "The Hobbit",
				Author:    "Tolkien",
				Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
			}
			db.Insert(&book).Do()

			Convey("Another player read the row", func() {
				var sameBook bookWithRowversion
				db.Select(&sameBook).Where("id = ? ", book.Id).Do()

				Convey("Both instances has same value for the oplock field", func() {
					// Expact same values in versions
					So(bytes.Compare(book.Version.Version, sameBook.Version.Version), ShouldEqual, 0)

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
