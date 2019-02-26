# godb - a Go query builder and struct mapper

[![Build Status](https://travis-ci.org/samonzeweb/godb.svg?branch=master)](https://travis-ci.org/samonzeweb/godb) [![GoDoc](https://godoc.org/github.com/samonzeweb/godb?status.svg)](https://godoc.org/github.com/samonzeweb/godb)

godb is a simple Go query builder and struct mapper, not a full-featured ORM. godb does not manage relationships.

Initially, godb was a learning project. The goal was to improve my Go skills by doing some useful things. But more and more features have been added and godb has become a serious project that can be used by others.

godb is a project that is still young and evolving. The API is almost stable, but it can still change slightly from one version to another. Each new version is associated with a tag, so it is possible to target a particular one if necessary.

## Features

- Queries builder.
- Mapping between structs and tables (or views).
- Mapping with nested structs.
- Execution of custom SELECT, INSERT, UPDATE and DELETE queries with structs and slices.
- Optional execution of SELECT queries with an iterator to limit memory consumption if needed (e.g. batches).
- Execution of raw queries, mapping rows to structs.
- Optimistic Locking
- SQL queries and durations logs.
- Two adjustable prepared statements caches (with/without transaction).
- `RETURNING` support for PostgreSQL.
- `OUTPUT` support for SQL Server.
- Optional common db errors handling for backend databases.(`db.UseErrorParser()`)
- Define your own logger (should have `Println(...)` method)
- Define model struct name to db table naming with `db.SetDefaultTableNamer(yourFn)`. Supported types are: Plural,Snake,SnakePlural. You can also define `TableName() string` method to for your struct and return whatever table name will be.
- BlackListing or WhiteListing columns for struct based inserts and updates.
- Could by used with
  - SQLite
  - PostgreSQL
  - MySQL / MariaDB
  - MS SQL Server
  - other compatible database if you write an adapter.

I made tests of godb on differents architectures and operating systems : OSX, Windows, Linux, ARM (Cortex A7) and Intel x64.

godb is compatible from Go 1.9 to 1.12 (SQL Server driver requires at least Go 1.8, sync.Map for caching needs at least Go 1.9).

## Documentation

There are three forms of documentation :

- This README with the example presented below, which gives an overview of what godb allows.
- The tests in `dbtests/common`, which are run on the different databases supported.
- Detailed documentation on GoDoc: https://godoc.org/github.com/samonzeweb/godb

## Install

```
go get github.com/samonzeweb/godb
```

Install the required driver (see tests). You cas use multiple databases if needed.

Of course you can also use a dependency management tool like `dep`.

## Running Tests

godb tests use GoConvey and at least SQLite :

```
go get github.com/smartystreets/goconvey
go get github.com/mattn/go-sqlite3
```

To run tests, go into the godb directory and executes `go test ./...`

SQLite tests are done with in memory database, it's fast. You can run tests with others databases, see below.

With the exception of SQLite, all drivers are _pure Go_ code, and does not require external dependencies.

### Test with PostgreSQL

Install the driver and set the `GODB_POSTGRESQL` environment variable with the PostgreSQL connection string.

```
go get github.com/lib/pq
GODB_POSTGRESQL="your connection string" go test ./...
```

### Test with MySQL / MariaDB

Install the driver and set the `GODB_MYSQL` environment variable with the MySQL connection string.

```
go get github.com/go-sql-driver/mysql
GODB_MYSQL="your connection string" go test ./...
```

### Test with MS SQL Server

Install the driver and set the `GODB_MSSQL` environment variable with the SQL Server connection string.

```
go get github.com/denisenkom/go-mssqldb
GODB_MSSQL="your connection string" go test ./...
```

### Test all with Docker

Using Docker you can test with SQLite, PostgreSQL, MariaDB and SQL Server with the `testallwithdocker.sh` shell script.

SQL Server is greedy, on OSX allow at least 4Go to Docker.

If the containers are slow to start, the script could wait before accessing them. Simply add the time to wait in seconds as arguments :

```
./testallwithdocker.sh 15
```

## Example

The example below illustrates the main features of godb.

You can copy the code into an `example.go` file and run it. You need to create the database and the `books` table as explained in the code.

```go
package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"log"
	"os"
)

/*
  To run this example, initialize a SQLite3 DB called 'library.db' and add
  a 'books' table like this :

  create table books (
  	id        integer not null primary key autoincrement,
  	title     text    not null,
  	author    text    not null,
  	published date    not null);
*/

// Struct and its mapping
type Book struct {
	Id        int       `db:"id,key,auto"`
	Title     string    `db:"title"`
	Author    string    `db:"author"`
	Published time.Time `db:"published"`
}

// Optional, default if the struct name (Book)
func (*Book) TableName() string {
	return "books"
}

// See "group by" example
type CountByAuthor struct {
	Author string `db:"author"`
	Count  int    `db:"count"`
}

func main() {
	// Examples fixtures
	var authorTolkien = "J.R.R. tolkien"

	var bookTheHobbit = Book{
		Title:     "The Hobbit",
		Author:    authorTolkien,
		Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
	}

	var bookTheFellowshipOfTheRing = Book{
		Title:     "The Fellowship of the Ring",
		Author:    authorTolkien,
		Published: time.Date(1954, 7, 29, 0, 0, 0, 0, time.UTC),
	}

	var bookTheTwoTowers = Book{
		Title:     "The Two Towers",
		Author:    authorTolkien,
		Published: time.Date(1954, 11, 11, 0, 0, 0, 0, time.UTC),
	}

	var bookTheReturnOfTheKing = Book{
		Title:     "The Return of the King",
		Author:    authorTolkien,
		Published: time.Date(1955, 10, 20, 0, 0, 0, 0, time.UTC),
	}

	var setTheLordOfTheRing = []Book{
		bookTheFellowshipOfTheRing,
		bookTheTwoTowers,
		bookTheReturnOfTheKing,
	}

	// Connect to the DB
	db, err := godb.Open(sqlite.Adapter, "./library.db")
	panicIfErr(err)
	// OPTIONAL: Set logger to show SQL execution logs
	db.SetLogger(log.New(os.Stderr, "", 0))
	// OPTIONAL: Set default table name building style from struct's name(if active struct doesn't have TableName() method)
	db.SetDefaultTableNamer(tablenamer.Plural())
	// Single insert (id will be updated)
	err = db.Insert(&bookTheHobbit).Do()
	panicIfErr(err)

	// Multiple insert
	// Warning : BulkInsert only update ids with PostgreSQL and SQL Server!
	err = db.BulkInsert(&setTheLordOfTheRing).Do()
	panicIfErr(err)

	// Count
	count, err := db.SelectFrom("books").Count()
	panicIfErr(err)
	fmt.Println("Books count : ", count)

	// Custom select
	countByAuthor := make([]CountByAuthor, 0, 0)
	err = db.SelectFrom("books").
		Columns("author", "count(*) as count").
		GroupBy("author").
		Having("count(*) > 3").
		Do(&countByAuthor)
	fmt.Println("Count by authors : ", countByAuthor)

	// Select single object
	singleBook := Book{}
	err = db.Select(&singleBook).
		Where("title = ?", bookTheHobbit.Title).
		Do()
	if err == sql.ErrNoRows {
		// sql.ErrNoRows is only returned when the target is a single instance
		fmt.Println("Book not found !")
	} else {
		panicIfErr(err)
	}

	// Select single record values
	authorName := ""
	title := ""
	err = db.SelectFrom("books").
		Where("title = ?", bookTheHobbit.Title).
		Columns("author", "title").
		Scanx(&authorName, &title)
	if err == sql.ErrNoRows {
		// sql.ErrNoRows is only returned when the target is a single instance
		fmt.Println("Book not found !")
	} else {
		panicIfErr(err)
	}

	// Select multiple objects
	multipleBooks := make([]Book, 0, 0)
	err = db.Select(&multipleBooks).Do()
	panicIfErr(err)
	fmt.Println("Books found : ", len(multipleBooks))

	// Iterator
	iter, err := db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		DoWithIterator()
	panicIfErr(err)
	for iter.Next() {
		book := Book{}
		err := iter.Scan(&book)
		panicIfErr(err)
		fmt.Println(book)
	}
	panicIfErr(iter.Err())
	panicIfErr(iter.Close())

	// Raw query
	subQuery := godb.NewSQLBuffer(0, 0). // sizes are indicative
						Write("select author ").
						Write("from books ").
						WriteCondition(godb.Q("where title = ?", bookTheHobbit.Title))

	queryBuffer := godb.NewSQLBuffer(64, 0).
		Write("select * ").
		Write("from books ").
		Write("where author in (").
		Append(subQuery).
		Write(")")

	panicIfErr(queryBuffer.Err())

	books := make([]Book, 0, 0)
	err = db.RawSQL(queryBuffer.SQL(), queryBuffer.Arguments()...).Do(&books)
	panicIfErr(err)
	fmt.Printf("Raw query found %d books\n", len(books))

	// Update and transactions
	err = db.Begin()
	panicIfErr(err)

	updated, err := db.UpdateTable("books").Set("author", "Tolkien").Do()
	panicIfErr(err)
	fmt.Println("Books updated : ", updated)

	bookTheHobbit.Author = "Tolkien"
	err = db.Update(&bookTheHobbit).Do()
	panicIfErr(err)
	fmt.Println("Books updated : ", updated)

	err = db.Rollback()
	panicIfErr(err)

	// Delete
	deleted, err := db.Delete(&bookTheHobbit).Do()
	panicIfErr(err)
	fmt.Println("Books deleted : ", deleted)

	deleted, err = db.DeleteFrom("books").
		WhereQ(godb.Or(
			godb.Q("author = ?", authorTolkien),
			godb.Q("author = ?", "Georged Orwell"),
		)).
		Do()
	panicIfErr(err)
	fmt.Println("Books deleted : ", deleted)

	// Bye
	err = db.Close()
	panicIfErr(err)
}

// It's just an example, what did you expect ? (never do that in real code)
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
```

# Licence

Released under the MIT License, see LICENSE.txt for more informations.
