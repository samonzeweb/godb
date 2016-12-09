/*
Package godb is a simple ORM allowing go code to execute sql queries with go
structs.

godb does not manage relationships like Active Record or Entity Framework, it's
a lighter library. Its goal is to be more production than manually doing mapping
between Go structs and databases tables. godb also have a sql builder.

godb needs adapters to use databases, some are packaged with godb for :

  * SQLite
  * PostgreSQL
  * MySQL
  * SQL Server

Start with an adapter, and the Open method which returns a godb.DB pointer :

  import (
  	"gitlab.com/samonzeweb/godb"
  	"gitlab.com/samonzeweb/godb/adapters/sqlite"
  )

  func main() {
    db, err := godb.Open(sqlite.Adapter, "./library.db")
    if err != nil {
      log.Fatal(err)
    }
    …
  }

There are two main tools family in godb : the statements tools and the structs
tools. Statements tools allow to write and executes row SQL select, insert,
update and delete. Structs tools looks more 'orm-ish' as they're take instances
if objects or slices to run select, insert, update and delete.

The statements tools are based on types :

  * SelectStatement : initialize it with db.SelectFrom
  * InsertStatement : initialize it with with db.InsertInto
  * UpdateStatement : initialize it with with db.UpdateTable
  * DeleteStatement : initialize it with with db.DeleteFrom

Example :

  type CountByAuthor struct {
    Author string `db:"author"`
    Count  int    `db:"count"`
  }
  …

  count, err := db.SelectFrom("books").Count()
  …

  err = db.SelectFrom("books").
		Columns("author", "count(*) as count").
		GroupBy("author").
		Having("count(*) > 3").
		Do(&countByAuthor)
  …

  newId, err := db.InsertInto("dummies")
    .Columns("foo", "bar", "baz")
    .Values(1, 2, 3)
    .Do()
  …

The structs tools are based on types :

  * StructSelect : initialize it with db.Select
  * StructInsert : initialize it with db.Insert or db.BulkInsert
  * StructUpdate : initialize it with db.Update
  * StructDelete : initialize it with db.Delete

Examples :

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

  …

  bookTheHobbit := Book{
    Title:     "The Hobbit",
    Author:    authorTolkien,
    Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
  }

  err = db.Insert(&bookTheHobbit).Do()
  …

  singleBook := Book{}
  err = db.Select(&singleBook).
		Where("title = ?", bookTheHobbit.Title).
		Do()
  …

  multipleBooks := make([]Book, 0, 0)
  err = db.Select(&multipleBooks).Do()
  …


  // TODO : struct reflexion

  // TODO : an example after each Struct (or its builder bethod)

*/
package godb
