package godb

import (
	"testing"

	"github.com/samonzeweb/godb/tablenamer"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClone(t *testing.T) {
	Convey("Given an existing DB", t, func() {
		db := createInMemoryConnection(t)
		defer db.Close()

		Convey("Clone create a DB copy of an existing one", func() {
			clone := db.Clone()

			So(clone.adapter, ShouldHaveSameTypeAs, db.adapter)
			So(clone.sqlDB, ShouldEqual, db.sqlDB)
			So(clone.logger, ShouldEqual, db.logger)
			So(clone.defaultTableNamer, ShouldEqual, db.defaultTableNamer)
		})

		Convey("Clone don't copy existing transaction", func() {
			db.Begin()
			clone := db.Clone()
			defer clone.Clear()
			So(clone.sqlTx, ShouldBeNil)
		})
	})
}

func TestTableNamer(t *testing.T) {
	db := createInMemoryConnection(t)
	defer db.Close()
	// Same
	Convey("Given a record descriptor, same name", t, func() {
		db.SetDefaultTableNamer(tablenamer.Same())
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "typeToDescribe")
		})
	})

	Convey("Given a record descriptor of type implementing tableNamer interface, same name", t, func() {
		db.SetDefaultTableNamer(tablenamer.Same())
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName()", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "others")
		})
	})

	// Plural
	Convey("Given a record descriptor, plural name", t, func() {
		db.SetDefaultTableNamer(tablenamer.Plural())
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name in plural form", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "typeToDescribes")
		})
	})

	Convey("Given a record descriptor of type implementing tableNamer interface, in plural form", t, func() {
		db.SetDefaultTableNamer(tablenamer.Plural())
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName() - plural", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "others")
		})
	})

	// Snake
	Convey("Given a record descriptor, snake case name", t, func() {
		db.SetDefaultTableNamer(tablenamer.Snake())
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name in snake form", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "type_to_describe")
		})
	})

	Convey("Given a record descriptor of type implementing tableNamer interface, snake name", t, func() {
		db.SetDefaultTableNamer(tablenamer.Snake())
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName() - snake", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "others")
		})
	})

	// Snake Plural
	Convey("Given a record descriptor, snake case name in plural", t, func() {
		db.SetDefaultTableNamer(tablenamer.SnakePlural())
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name in plural snake form", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "type_to_describes")
		})
	})

	Convey("Given a record descriptor of type implementing tableNamer interface, plural snake name", t, func() {
		db.SetDefaultTableNamer(tablenamer.SnakePlural())
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName() - plural snake", func() {
			tableName := db.defaultTableNamer(recordDesc.getTableName())
			So(tableName, ShouldEqual, "others")
		})
	})
}

func TestQuote(t *testing.T) {
	Convey("Given an existing DB", t, func() {
		db := createInMemoryConnection(t)
		defer db.Close()

		Convey("quote act like adapter Quote with a simple identified", func() {
			identifier := "foo"
			quotedIdentifier := db.quote(identifier)

			So(quotedIdentifier, ShouldEqual, db.adapter.Quote(identifier))
		})

		Convey("quote quotes all parts of an identifier", func() {
			identifier := "foo.bar.baz"
			quotedIdentifier := db.quote(identifier)

			expectedQuotedIdentified := db.adapter.Quote("foo") + "." +
				db.adapter.Quote("bar") + "." +
				db.adapter.Quote("baz")

			So(quotedIdentifier, ShouldEqual, expectedQuotedIdentified)
		})
	})
}
