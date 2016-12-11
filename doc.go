/*
Package godb is a simple ORM allowing go code to execute sql queries with go
structs.

godb does not manage relationships like Active Record or Entity Framework, it's
a lighter library. Its goal is to be more productive than manually doing mapping
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
tools. Statements tools allows to write and execute row SQL select, insert,
update and delete. Structs tools looks more 'orm-ish' as they're take instances
of objects or slices to run select, insert, update and delete.


Statements tools


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


Structs tools



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


Structs mapping


Stucts contents are mapped to databases columns with tags, like in previous
example with the Book struct. The tag is 'db' and its content is :

	* The columns name (mandatory, there is no default rule).
	* The 'key' keyword if the field/column is a part of the table key.
	* The 'auto' keyword if the field/column value is set by the database.

For autoincrement identifier simple use both 'key' and 'auto'.

Example :

	type SimpleStruct struct {
		ID    int    `db:"id,key,auto"`
		Text  string `db:"my_text"`
		// ignored
		Other string
	}

More than one field could have the 'key' keyword, but with most databases
drivers none of them could have the 'auto' keyword, because executing an insert
query only returns one value : the last inserted id : https://golang.org/pkg/database/sql/driver/#RowsAffected.LastInsertId .

With PostgreSQL you cas have multiple fields with 'key' and 'auto' options.

Structs could be nested. A nested struct is mapped only if has the 'db' tag.
The tag value is a columns prefix applied to all fields columns of the struct.
The prefix is not mandatory, a blank string is allowed (no prefix).

Exemple

	type KeyStruct struct {
		ID    int    `db:"id,key,auto"`
	}

	type ComplexStruct struct {
		KeyStruct          `db:""`
		Foobar   SubStruct `db:"nested_"`
		Ignored  SubStruct
	}

	type SubStruct struct {
		Foo string `db:"foo"`
		Bar string `db:"bar"`
	}

Databases columns are :

	* id (no prefix)
	* nested_foo
	* nested_bar

The mapping is managed by the 'dbreflect' subpackage. Normally its direct use
is not necessary, exept in one case : some structs are scannable and have to be
considered like fields, and mapped to databases columns. Common case are
time.Time, or sql.NullString, ... You can register a custom struct with the
`RegisterScannableStruct` and a struct instance, for example the time.Time is
registered like this :

	dbreflect.RegisterScannableStruct(time.Time{})

The structs statements use the struct name as table name. But you can override
this simply by simplementing a TableName method :

	func (*Book) TableName() string {
		return "books"
	}


Conditions


Statements and structs tools manage 'where' and 'group by' sql clauses. These
conditionnal clauses are build either with raw sql code, or build with the
Condition struct like this :

	q := godb.Or(godb.Q("foo is null"), godb.Q("foo > ?", 123))
	count, err := db.SelectFrom("bar").WhereQ(q).Count()

WhereQ methods take a Condition instance build by godb.Q . Where mathods take
raw SQL, but is just a syntactic sugar. These calls are equivalents :

	…WhereQ(godb.Q("id = ?", 123))…
	…Where("id = ?", 123)…

Multiple calls to Where or WhereQ are allowed, these calls are equivalents :

	…Where("id = ?", 123).Where("foo is null")…
	…WhereQ(godb.And(godb.Q("id = ?", 123), godb.Q("foo is null")))…

Slices are managed in a particular way : a single placeholder is replaced with
multiple ones. This allows code like :

	count, err := db.SelectFrom("bar").Where("foo in (?)", fooSlice).Count()


Consumed Time

godb keep track of time consumed while executing queries. You can reset it and
get the time consumed since Open or the previous reset :

	fmt.Prinln("Consumed time : %v", db.ConsumedTime())
	db.ResetConsumedTime()


Logger


You can log all executed queried and details of condumed time. Simply add a
logger :

	db.SetLogger(log.New(os.Stderr, "", 0))

*/
package godb
