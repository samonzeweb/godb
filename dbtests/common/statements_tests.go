package common

import (
	"strings"
	"testing"
	"time"

	"github.com/samonzeweb/godb"
)

func StatementsTests(db *godb.DB, t *testing.T) {
	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	// Check experimental prepared statement cache for sql.DB
	// db.StmtCacheDB().Enable()

	statementInsertTest(db, t)
	statementSelectTest(db, t)
	statementUpdateTest(db, t)
	statementDeleteTest(db, t)
}

func statementInsertTest(db *godb.DB, t *testing.T) {
	returningBuilder := getReturningBuilder(db)

	// Simple insert
	query := db.InsertInto("books").
		Columns("title", "author", "published").
		Values(bookTheHobbit.Title, bookTheHobbit.Author, bookTheHobbit.Published)

	id, err := query.Do()
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 && returningBuilder == nil {
		t.Fatal("Id was not returned.")
	}

	// Multiple insert (without returning clause)
	booksToInsert := setTheLordOfTheRing[:]
	if returningBuilder == nil {
		booksToInsert = append(booksToInsert, setFoundation...)
	}
	query = db.InsertInto("books").
		Columns("title", "author", "published")
	for _, book := range booksToInsert {
		query.Values(book.Title, book.Author, book.Published)
	}
	_, err = query.Do()
	if err != nil {
		t.Fatal(err)
	}

	// Multiple insert with returning clause (if implemented)
	// As the given slice isn't empty, it is filled with the values returned
	// by the database. Of course order of values and slice content must match !
	if returningBuilder != nil {
		booksToInsert = setFoundation[:]

		query = db.InsertInto("books").
			Columns("title", "author", "published").
			Returning(returningBuilder.FormatForNewValues([]string{"id"})...)
		for _, book := range booksToInsert {
			query.Values(book.Title, book.Author, book.Published)
		}
		_, err = query.DoWithReturning(&booksToInsert)
		if err != nil {
			t.Fatal(err)
		}
		for _, book := range booksToInsert {
			if book.Id == 0 {
				t.Fatalf("Id not set, fail go get returning values : %v", book)
			}
		}
	}
}

func statementSelectTest(db *godb.DB, t *testing.T) {
	// Count books
	count := CountBooks(t, db)
	if count != 7 {
		t.Fatalf("Wrong books count : %v", count)
	}

	// Select a single row
	book := Book{}
	err := db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		Where("title = ?", bookTheHobbit.Title).Do(&book)
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != bookTheHobbit.Title {
		t.Fatalf("Book not filled : %v", book)
	}

	// Select multiple rows with order
	allBooks := make([]Book, 0, 0)
	err = db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		OrderBy("author").OrderBy("title").
		Do(&allBooks)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(allBooks)) != count {
		t.Fatalf("Wrong books count : %v", len(allBooks))
	}
	if allBooks[0].Title != bookFoundation.Title {
		t.Fatalf("Wrong book order first is : %v", allBooks[0])
	}

	// Select with group by and having
	countByAuthor := make([]CountByAuthor, 0, 0)
	err = db.SelectFrom("books").
		Columns("author", "count(*) as count").
		GroupBy("author").
		Having("count(*) > 3").
		Do(&countByAuthor)
	if err != nil {
		t.Fatal(err)
	}
	if len(countByAuthor) != 1 {
		t.Fatalf("Wrong count by author, total rows is : %v", len(countByAuthor))
	}
	if countByAuthor[0].Author != authorTolkien ||
		countByAuthor[0].Count != 4 {
		t.Fatalf("Wrong result : %v", countByAuthor[0])
	}

	// Select with complex condition
	titles := []string{
		bookFoundation.Title,
		bookFoundationAndEmpire.Title,
	}
	q := godb.And(
		godb.Q("author = ?", authorAssimov),
		godb.Q("title in (?)", titles),
	)
	twoBooks := make([]Book, 0, 0)
	err = db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		WhereQ(q).
		Do(&twoBooks)
	if err != nil {
		t.Fatal(err)
	}
	if len(twoBooks) != 2 {
		t.Fatalf("Wrong result, books count : %v", len(twoBooks))
	}

	// Select using a struct to build columns names
	// Add more fixtures : inventories for Tolkien's books
	tolkiensBooks := make([]Book, 0, 0)
	err = db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		Where("author = ?", authorTolkien).
		OrderBy("published").
		Do(&tolkiensBooks)
	if err != nil {
		t.Fatal(err)
	}
	for _, book := range tolkiensBooks {
		inventory := Inventory{
			BookId:        book.Id,
			LastInventory: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			Counting:      book.Published.Year(), // easy to test later
		}
		err := db.Insert(&inventory).Do()
		if err != nil {
			t.Fatal(err)
		}
	}
	// Find all books with their inventories (not all books have one)
	booksWithInventories := make([]BooksWithInventories, 0, 0)
	err = db.SelectFrom("books").
		ColumnsFromStruct(&booksWithInventories).
		LeftJoin("inventories", "inventories", godb.Q("inventories.book_id = books.id")).
		Do(&booksWithInventories)
	if err != nil {
		t.Fatal(err)
	}
	if len(booksWithInventories) != len(allBooks) {
		t.Fatalf("Wrong books+inventories count : %v", len(booksWithInventories))
	}
	for _, bookWithIventory := range booksWithInventories {
		switch bookWithIventory.Author {
		case authorTolkien:
			if !bookWithIventory.Counting.Valid ||
				bookWithIventory.Counting.Int64 != int64(bookWithIventory.Published.Year()) {
				t.Fatalf("Wrong counting %v for book %v", bookWithIventory.Counting, bookWithIventory.Title)
			}
		case authorAssimov:
			// no inventory
			if bookWithIventory.Counting.Valid {
				t.Fatalf("Wrong counting %v for book %v", bookWithIventory.Counting, bookWithIventory.Title)
			}
		default:
			t.Fatalf("Wrong author in inventory %v", bookWithIventory.Author)
		}
	}

	// Select with an iterator
	iter, err := db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		OrderBy("author").OrderBy("title").
		DoWithIterator()
	if err != nil {
		t.Fatal(err)
	}
	count = 0
	for iter.Next() {
		count++
		book := Book{}
		if err := iter.Scan(&book); err != nil {
			t.Fatal(err)
		}
		if count == 1 {
			if book.Author != bookFoundation.Author {
				t.Fatalf("Book isn't filled with iterator")
			}
			if book.Title != bookFoundation.Title {
				t.Fatalf("Book isn't filled with iterator")
			}
		}
	}
	if iter.Err() != nil {
		t.Fatal(err)
	}
	if iter.Close() != nil {
		t.Fatal(err)
	}
	if count != 7 {
		t.Fatalf("Wrong books count found with an iterator : %v", count)
	}

	// Select with transaction and nested iterators.
	// Some drivers will cause troubles with nested queries during a
	// transaction. This test ensures that DoWithIterator() does not use
	// a current transaction to avoid troubles.
	db.Begin()
	// Should not have effect on the next two select statement
	db.DeleteFrom("books").Do()

	iter, err = db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		OrderBy("author").OrderBy("title").
		DoWithIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	iter2, err := db.SelectFrom("books").
		Columns("id", "title", "author", "published").
		DoWithIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter2.Close()

	count = 0
	for iter.Next() {
		count++
	}
	if count != 7 {
		t.Fatalf("Wrong books count found with an iterator : %v", count)
	}

	count = 0
	for iter2.Next() {
		count++
	}
	if count != 7 {
		t.Fatalf("Wrong books count found with an iterator : %v", count)
	}

	db.Rollback()
}

func statementUpdateTest(db *godb.DB, t *testing.T) {
	returningBuilder := getReturningBuilder(db)

	db.Begin()

	// Update books
	gandalf := "Gandalf the Grey"
	updated, err := db.UpdateTable("books").
		Set("author", gandalf).
		SetRaw("title = 'book by Gandalf'").
		Where("author = ?", authorTolkien).
		Do()
	if err != nil {
		t.Fatal(err)
	}
	if updated != 4 {
		t.Fatalf("Wrong count of updated books : %v", updated)
	}

	// Count changed
	count, err := db.SelectFrom("books").
		Where("author = ?", gandalf).
		Where("title = 'book by Gandalf'").
		Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != updated {
		t.Fatalf("Wrong books count : %v", count)
	}

	db.Rollback()

	if returningBuilder != nil {
		db.Begin()
		// Update books and get back all updated books
		// As the given slice is empty, it will add instances filled by the values
		// returned by the database.
		updatedBooks := make([]Book, 0, 0)
		hari := "Hari Seldon"
		_, err = db.UpdateTable("books").
			Set("author", hari).
			Where("author = ?", authorAssimov).
			Returning(returningBuilder.FormatForNewValues([]string{"id", "title", "author", "published"})...).
			DoWithReturning(&updatedBooks)
		if err != nil {
			t.Fatal(err)
		}
		for _, book := range updatedBooks {
			if book.Id == 0 || book.Author != hari || !strings.Contains(book.Title, "Foundation") {
				t.Fatalf("Fields not set in update statement with returning clause : %v", book)
			}
		}
		db.Rollback()
	}
}

func statementDeleteTest(db *godb.DB, t *testing.T) {
	returningBuilder := getReturningBuilder(db)

	db.Begin()

	deleted, err := db.DeleteFrom("books").
		Where("author = ?", authorAssimov).
		Do()
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 3 {
		t.Fatalf("Wrong count of deleted books : %v", deleted)
	}

	count := CountBooks(t, db)
	if count != 4 {
		t.Fatalf("Wrong books count : %v", count)
	}

	db.Rollback()

	if returningBuilder != nil {
		// A little hack as SQL Server need a prefix
		var returningColumns []string
		if db.Adapter().DriverName() == "mssql" {
			returningColumns = []string{"deleted.id", "deleted.title", "deleted.author", "deleted.published"}
		} else {
			returningColumns = []string{"id", "title", "author", "published"}
		}
		deletedBooks := make([]Book, 0, 0)
		_, err = db.DeleteFrom("books").
			Where("author = ?", authorAssimov).
			Returning(returningColumns...).
			DoWithReturning(&deletedBooks)
		if err != nil {
			t.Fatal(err)
		}
		for _, book := range deletedBooks {
			if book.Id == 0 || book.Author != authorAssimov || !strings.Contains(book.Title, "Foundation") {
				t.Fatalf("Fields not set in delete statement with returning clause : %v", book)
			}
		}
		count := CountBooks(t, db)
		if count != 4 {
			t.Fatalf("Wrong books count : %v", count)
		}

		db.Rollback()
	}
}
