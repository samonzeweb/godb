# godb - a simple Go ORM

[![Build Status](https://travis-ci.org/samonzeweb/godb.svg?branch=master)](https://travis-ci.org/samonzeweb/godb) [![GoDoc](https://godoc.org/github.com/samonzeweb/godb?status.svg)](https://godoc.org/github.com/samonzeweb/godb)

godb is a simple Go ORM. It contains a simple SQL query builder and manages mapping between SQL and structs.

Initially godb was a learning project. The purpose was to improve my Go skills by doing real and usable stuff. But it could be useful for somebody else.

The documentation is as young as the code, but it exists ! You have an example (see below) showing you the main features, and more details in GoDoc : https://godoc.org/github.com/samonzeweb/godb

WARNING : it is still a young project and the public API could change.

## Features

* Queries builder.
* Mapping between structs and tables.
* Mapping with nested structs.
* Execution of custom SELECT, INSERT, UPDATE and DELETE queries with structs and slices.
* SQL queries and durations logs.
* Two adjustable prepared statements caches (with/without transaction).
* `RETURNING` support for PostgreSQL.
* Could by used with
  * SQLite
  * PostgreSQL
  * MySQL / MariaDB
  * MS SQL Server
  * other compatible database if you write an adapter.

godb does not manage relationship.

I made some tests of godb on Windows and Linux, ARM (Cortex A7) and Intel x64.

## Install

```
go get github.com/samonzeweb/godb
```

Install the required driver (see tests). You cas use multiple databases if needed.

## Tests

godb tests use GoConvey and at least SQLite :

```
go get github.com/smartystreets/goconvey
go get github.com/mattn/go-sqlite3
```

To run tests, go into the godb directory and executes `go test ./...`

SQLite tests are done with in memory database, it's fast. You can run tests with others databases, see below.

With the exception of SQLite, all drivers are *pure Go* code, and does not require external dependencies.

### Tests with PostgreSQL

Install the driver and set the `GODB_POSTGRESQL` environment variable with the PostgreSQL connection string.

```
go get github.com/lib/pq
GODB_POSTGRESQL="your connection string" go test ./...
```

### Tests with MySQL / MariaDB

Install the driver and set the `GODB_MYSQL` environment variable with the MySQL connection string.

```
go get github.com/go-sql-driver/mysql
GODB_MYSQL="your connection string" go test ./...
```

### Tests with MS SQL Server

Install the driver and set the `GODB_MSSQL` environment variable with the SQL Server connection string.

```
go get github.com/denisenkom/go-mssqldb
GODB_MSSQL="your connection string" go test ./...
```

## Example

The example below illustrates the main features of godb.

You can copy the code into an `example.go` file and run it. You need to create the database and the `books` table as explained in the code.


```
package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
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

// Optionnal, default if the struct name (Book)
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

	// Single insert (id will be updated)
	err = db.Insert(&bookTheHobbit).Do()
	panicIfErr(err)

	// Multiple insert
	// Warning : BulkInsert only update ids with PostgreSQL !
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

	// Select multiple objects
	multipleBooks := make([]Book, 0, 0)
	err = db.Select(&multipleBooks).Do()
	panicIfErr(err)
	fmt.Println("Books found : ", len(multipleBooks))

	// Update and transactions
	err = db.Begin()
	panicIfErr(err)

	updated, err := db.UpdateTable("books").Set("author", "Tolkien").Do()
	panicIfErr(err)
	fmt.Println("Books updated : ", updated)

	bookTheHobbit.Author = "Tolkien"
	updated, err = db.Update(&bookTheHobbit).Do()
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

// It's just an example, what did you expect ?
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

```

# Licence

Released under the MIT License, see LICENSE.txt for more informations.
