package mysql_test

import (
	"log"
	"os"
	"testing"

	"gitlab.com/samonzeweb/godb"
	"gitlab.com/samonzeweb/godb/adapters/mysql"
	"gitlab.com/samonzeweb/godb/dbtests/common"

	. "github.com/smartystreets/goconvey/convey"
)

func fixturesSetup(t *testing.T) (*godb.DB, func()) {
	if os.Getenv("GODB_MYSQL") == "" {
		t.Fatal("Don't run MySQL test, GODB_MYSQL not set")
	}

	db, err := godb.Open(mysql.Adapter, os.Getenv("GODB_MYSQL"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	db.SetLogger(log.New(os.Stderr, "", 0))

	createTable :=
		`create temporary table if not exists books (
		id 						int auto_increment primary key,
		title     		varchar(128) not null,
		author    	  varchar(128) not null,
		published			date not null);
	`
	_, err = db.CurrentDB().Exec(createTable)
	if err != nil {
		panic(err)
	}

	fixturesTeardown := func() {
		dropTable := "drop table if exists books"
		_, err := db.CurrentDB().Exec(dropTable)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, fixturesTeardown
}

func TestMySQL(t *testing.T) {
	Convey("A DB for a MySQL database", t, func() {
		db, teardown := fixturesSetup(t)
		defer teardown()

		Convey("The common tests must pass", func() {
			common.MainTest(db, t)
		})
	})
}
