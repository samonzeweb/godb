package godb

import (
	"testing"

	"gitlab.com/samonzeweb/godb/adapters/sqlite"

	. "github.com/smartystreets/goconvey/convey"
)

func createConnection() *DB {
	db, err := Open(sqlite.Adapter, ":memory:")
	if err != nil {
		panic(err)
	}

	return db
}

func TestBegin(t *testing.T) {
	Convey("Given an existing connection", t, func() {
		db := createConnection()

		Convey("Begin create a new transaction", func() {
			err := db.Begin()
			So(err, ShouldBeNil)
			So(db.sqlTx, ShouldNotBeNil)

			Convey("Begin fails is a transactions already exists", func() {
				err = db.Begin()
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestCommit(t *testing.T) {
	Convey("Given an existing connexion", t, func() {
		db := createConnection()

		Convey("Commit end an existing transaction", func() {
			db.Begin()
			err := db.Commit()
			So(err, ShouldBeNil)
		})

		Convey("Commit fails if there is no transaction", func() {
			err := db.Commit()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRollback(t *testing.T) {
	Convey("Given an existing connexion", t, func() {
		db := createConnection()

		Convey("Commit end an existing transaction", func() {
			db.Begin()
			err := db.Rollback()
			So(err, ShouldBeNil)
		})

		Convey("Commit fails if there is no transaction", func() {
			err := db.Rollback()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCurrentTx(t *testing.T) {
	Convey("Given an existing connexion", t, func() {
		db := createConnection()

		Convey("CurrentTx returns nil there is no current transaction", func() {
			So(db.CurrentTx(), ShouldBeNil)
		})

		Convey("CurrentTx returns Tx if there is a current transaction", func() {
			db.Begin()
			So(db.CurrentTx(), ShouldEqual, db.sqlTx)
		})
	})
}

func TestGetTxElseDb(t *testing.T) {
	Convey("Given an existing connexion", t, func() {
		db := createConnection()

		Convey("getTxElseDb returns DB if there is no current transaction", func() {
			So(db.getTxElseDb(), ShouldEqual, db.sqlDB)
		})

		Convey("getTxElseDb returns Tx if there is a current transaction", func() {
			db.Begin()
			So(db.getTxElseDb(), ShouldEqual, db.sqlTx)
		})
	})
}
