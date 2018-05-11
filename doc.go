/*
Package godb is query builder and struct mapper.

godb does not manage relationships like Active Record or Entity Framework, it's
not a full-featured ORM. Its goal is to be more productive than manually doing mapping
between Go structs and databases tables.

godb needs adapters to use databases, some are packaged with godb for :

	* SQLite
	* PostgreSQL
	* MySQL
	* SQL Server

Start with an adapter, and the Open method which returns a godb.DB pointer :

	import (
		"github.com/samonzeweb/godb"
		"github.com/samonzeweb/godb/adapters/sqlite"
	)

	func main() {
		db, err := godb.Open(sqlite.Adapter, "./library.db")
		if err != nil {
			log.Fatal(err)
		}
		…
	}

There are three ways to executes SQL with godb :

	* the statements tools
	* the structs tools
	* and raw queries

Using raw queries you can execute any SQL queries and get the results into
a slice of structs (or single struct) using the automatic mapping.

Structs tools looks more 'orm-ish' as they're take instances
of objects or slices to run select, insert, update and delete.

Statements tools stand between raw queries and structs tools. It's easier to
use than raw queries, but are limited to simplier cases.


Statements tools


The statements tools are based on types :

	* SelectStatement : initialize it with db.SelectFrom
	* InsertStatement : initialize it with db.InsertInto
	* UpdateStatement : initialize it with db.UpdateTable
	* DeleteStatement : initialize it with db.DeleteFrom

Example :

	type CountByAuthor struct {
		Author string `db:"author"`
		Count  int    `db:"count"`
	}
	…

	count, err := db.SelectFrom("books").Count()
	…

	countByAuthor := make([]CountByAuthor, 0, 0)
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

The SelectStatement type could also build a query using columns from a structs. It facilitates the build of queries returning values from multiple table (or views). See struct mapping explanations, in particular the `rel` part.

Example :

	type Book struct {
		Id        int       `db:"id,key,auto"`
		Title     string    `db:"title"`
		Author    string    `db:"author"`
		Published time.Time `db:"published"`
		Version   int       `db:"version,oplock"`
	}

	type InventoryPart struct {
		Id       sql.NullInt64 `db:"id"`
		Counting sql.NullInt64 `db:"counting"`
	}

	type BooksWithInventories struct {
		Book          `db:",rel=books"`
		InventoryPart `db:",rel=inventories"`
	}
	…

	booksWithInventories := make([]BooksWithInventories, 0, 0)
	err = db.SelectFrom("books").
		ColumnsFromStruct(&booksWithInventories).
		LeftJoin("inventories", "inventories", godb.Q("inventories.book_id = books.id")).
		Do(&booksWithInventories)


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


Raw queries

Raw queries are executed using the RawSQL type.

The query could be a simple hand-written string, or something complex builded
using SQLBuffer and Conditions.

Example :

	books := make([]Book, 0, 0)
	err = db.RawSQL("select * from books where author = ?", authorAssimov).Do(&books)


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

Structs could be nested. A nested struct is mapped only if has the 'db' tag. The tag value is a columns prefix applied to all fields columns of the struct. The prefix is not mandatory, a blank string is allowed (no prefix).

A nested struct could also have an optionnal `rel` attribute of the form `rel=relationname`. It's useful to build a select query using multiples relations (table, view, ...). See the example using the BooksWithInventories type.

Example

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
time.Time, or NullString, ... You can register a custom struct with the
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


SQLBuffer


The SQLBuffer exists to ease the build of complex raw queries. It's also used
internaly by godb. Its use and purpose are simple : concatenate sql parts
(accompagned by their arguments) in an efficient way.

Example :

	// see NewSQLBuffer for details about sizes
	subQuery := godb.NewSQLBuffer(32, 0).
		Write("select author ").
		Write("from books ").
		WriteCondition(godb.Q("where title = ?", bookFoundation.Title))

	queryBuffer := godb.NewSQLBuffer(64, 0).
		Write("select * ").
		Write("from books ").
		Write("where author in (").
		Append(subQuery).
		Write(")")

	if queryBuffer.Err() != nil {
		…
	}

	err = db.RawSQL(queryBuffer.SQL(), queryBuffer.Arguments()...).Do(&books)
	if err != nil {
		…
	}


Optimistic Locking

For all databases, structs updates and deletes manage optimistic locking when a dedicated integer row
is present. Simply tags it with `oplock` :

	type KeyStruct struct {
		...
		Version    int    `db:"version,oplock"`
		...
	}

When an update or delete operation fails, Do() returns the `ErrOpLock` error.

With PostgreSQL and SQL Server, godb manages optimistic locking with automatic fields.
Just add a dedicated field in the struct and tag it with `auto,oplock`.

With PostgreSQL you can use the `xmin` system column like this :


	type KeyStruct struct {
		...
		Version    int    `db:"xmin,auto,oplock"`
		...
	}

For more informations about `xmin` see
https://www.postgresql.org/docs/10/static/ddl-system-columns.html

With SQL Server you can use a `rowversion` field with the `mssql.Rowversion` type like this :


	type KeyStruct struct {
		...
		Version   mssql.Rowversion `db:"version,auto,oplock"`
		...
	}

For more informations about the `rowversion` data type see
https://docs.microsoft.com/en-us/sql/t-sql/data-types/rowversion-transact-sql


Consumed Time


godb keep track of time consumed while executing queries. You can reset it and
get the time consumed since Open or the previous reset :

	fmt.Prinln("Consumed time : %v", db.ConsumedTime())
	db.ResetConsumedTime()


Logger


You can log all executed queried and details of condumed time. Simply add a
logger :

	db.SetLogger(log.New(os.Stderr, "", 0))


RETURNING and OUTPUT Clauses


godb takes advantage of PostgreSQL RETURNING clause, and SQL Server OUTPUT clause.

With statements tools you have to add a RETURNING clause with the Suffix method
and call DoWithReturning method instead of Do(). It's optionnal.

With StructInsert it's transparent, the RETURNING or OUTPUT clause is added
for all 'auto' columns and it's managed for you. One of the big advantage is
with BulkInsert : for others databases the rows are inserted but the new keys
are unkonwns. With PostgreSQL and SQL Server the slice is updated for all inserted
rows.

It also enables optimistic locking with *automatic* columns.


Prepared statements cache


godb has two prepared statements caches, one to use during transactions, and
one to use outside of a transaction. Both use a LRU algorithm.

The transaction cache is enabled by default, but not the other. A transaction
(sql.Tx) isn't shared between goroutines, using prepared statement with it has a
predictable behavious. But without transaction a prepared statement could have
to be reprepared on a different connection if needed, leading to unpredictable
performances in high concurrency scenario.

Enabling the non transaction cache could improve performances with single
goroutine batch. With multiple goroutines accessing the same database : it
depends ! A benchmark would be wise.


Iterator

Using statements tools and structs tools you can execute select queries and get an
iterator instead of filling a slice of struct instances. This could be useful if the
request's result is big and you don't want to allocate too much memory. On the other
side you will write almost as much code as with the `sql` package, but with an automatic
struct mapping, and a request builder.

Iterators are also available with raw queries. In this cas you cas executes any kind of
sql code, not just select queries.

To get an interator simply use the `DoWithIterator` method instead of `Do`. The iterator
usage is similar to the standard `sql.Rows` type. Don't forget to check that there are
no errors with the `Err` method, and don't forget to call `Close` when the iterator is no
longer useful, especialy if you don't scan all the resultset.

	iter, err := db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		OrderBy("author").OrderBy("title").
		DoWithIterator()
	if err != nil {
		...
	}
	defer iter.Close()

	for iter.Next() {
		book := Book{}
		if err := iter.Scan(&book); err != nil {
			...
		}
		// do something with the book
		...
		}
	}

	if iter.Err() != nil {
		t.Fatal(err)
	}

*/
package godb
