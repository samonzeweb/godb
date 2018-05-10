package godb

import (
	"testing"

	"github.com/samonzeweb/godb/adapters/sqlite"
	"github.com/samonzeweb/godb/types"
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

func createInMemoryConnection(t *testing.T) *DB {
	db, err := Open(sqlite.Adapter, ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	return db
}

// Fixtures

type Dummy struct {
	ID              int            `db:"id,key,auto"`
	AText           string         `db:"a_text"`
	AnotherText     string         `db:"another_text"`
	AnInteger       int            `db:"an_integer"`
	ANullableString types.NullString `db:"a_nullable_string"`
	Version         int            `db:"version,oplock"`
}

func (*Dummy) TableName() string {
	return "dummies"
}

type RelatedToDummy struct {
	ID      int    `db:"id,key,auto"`
	DummyID int    `db:"dummies_id"`
	AText   string `db:"a_text"`
}

func (*RelatedToDummy) TableName() string {
	return "relatedtodummies"
}

type FromTwoTables struct {
	Dummy          `db:",rel=dummies"`
	RelatedToDummy `db:",rel=relatedtodummies"`
}

type DummyAutoOplock struct {
	ID              int            `db:"id,key,auto"`
	AText           string         `db:"a_text"`
	AnotherText     string         `db:"another_text"`
	AnInteger       int            `db:"an_integer"`
	ANullableString types.NullString `db:"a_nullable_string"`
	Version         int            `db:"version,auto,oplock"`
}

func (*DummyAutoOplock) TableName() string {
	return "dummiesautooplock"
}

func fixturesSetup(t *testing.T) *DB {
	db := createInMemoryConnection(t)

	createTable :=
		`create table dummies (
		id                  integer not null primary key autoincrement,
		a_text              text not null,
		another_text        text not null,
		an_integer          integer not null,
		a_nullable_string   text,
		version             integet not null default(0));

		create table relatedtodummies(
		id                  integer not null primary key autoincrement,
		dummies_id          integer not null,
		a_text              text not null
		);

		create table dummiesautooplock (
		id                  integer not null primary key autoincrement,
		a_text              text not null,
		another_text        text not null,
		an_integer          integer not null,
		a_nullable_string   text,
		version             integet not null default(0));

		create trigger updateversion
		after update
		on dummiesautooplock
		begin
			update dummiesautooplock set version = (NEW.version + 1);
		end;		
	`
	_, err := db.sqlDB.Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	insertRows :=
		`insert into dummies
		(a_text, another_text, an_integer, a_nullable_string)
		values
		("First", "Premier", 11, "Not empty"),
		("Second", "Second", 12, ""),
		("Third", "Troisième", 13, NULL);

		insert into relatedtodummies
		(dummies_id, a_text)
		select id, "REL_" || a_text from dummies;

		insert into dummiesautooplock
		(a_text, another_text, an_integer, a_nullable_string)
		values
		("First", "Premier", 11, "Not empty"),
		("Second", "Second", 12, ""),
		("Third", "Troisième", 13, NULL);		
	`
	_, err = db.sqlDB.Exec(insertRows)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
