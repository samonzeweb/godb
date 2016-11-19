package godb

import (
	"testing"

	"gitlab.com/samonzeweb/godb/adapters/sqlite"
)

func checkToSQL(t *testing.T, sqlExpected string, sqlProduced string, err error) {
	if err != nil {
		t.Fatal("ToSQL produces error :", err)
	}

	t.Log("SQL expected :", sqlExpected)
	t.Log("SQL produced :", sqlProduced)
	if sqlProduced != sqlExpected {
		t.Fatal("ToSQL produces incorrect SQL")
	}
}

func createInMemoryConnection() *DB {
	db, err := Open(sqlite.Adapter, ":memory:")
	if err != nil {
		panic(err)
	}

	return db
}

// Fixtures

type Dummy struct {
	ID          int    `db:"id,key,auto"`
	AText       string `db:"a_text"`
	AnotherText string `db:"another_text"`
	AnInteger   int    `db:"an_integer"`
}

func (*Dummy) TableName() string {
	return "dummies"
}

func fixturesSetup() *DB {
	db := createInMemoryConnection()

	createTable :=
		`create table dummies (
		id 						integer not null primary key autoincrement,
		a_text     		text not null,
		another_text	text not null,
		an_integer 		integer not null);
	`
	_, err := db.sqlDB.Exec(createTable)
	if err != nil {
		panic(err)
	}

	insertRows :=
		`insert into dummies
		(a_text, another_text, an_integer)
		values
		("First", "Premier", 11),
		("Second", "Second", 12),
		("Third", "Troisième", 13);
	`
	_, err = db.sqlDB.Exec(insertRows)
	if err != nil {
		panic(err)
	}

	return db
}
