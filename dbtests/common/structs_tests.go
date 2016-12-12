package common

import (
	"database/sql"
	"testing"

	"github.com/samonzeweb/godb"
)

func StructsTests(db *godb.DB, t *testing.T) {
	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	// Check experimental prepared statement cache for sql.DB
	// db.StmtCacheDB().Enable()

	structsInsertTest(db, t)
	structsSelectTest(db, t)
	structsUpdateTest(db, t)
	structsDeleteTest(db, t)
}

func structsInsertTest(db *godb.DB, t *testing.T) {
	// Single
	bookToInsert := bookTheHobbit
	err := db.Insert(&bookToInsert).Do()
	if err != nil {
		t.Fatal(err)
	}
	if bookToInsert.Id == 0 {
		t.Fatalf("Id was not set for the book %v", bookToInsert)
	}

	// Bulk with slice of instances
	booksToInsert := setTheLordOfTheRing[:]
	err = db.BulkInsert(&booksToInsert).Do()
	if err != nil {
		t.Fatal(err)
	}
	if hasReturning(db) {
		for _, book := range booksToInsert {
			if book.Id == 0 {
				t.Fatalf("Id was not set for the book %v", book)
			}
		}
	}

	// Bulk with slice of pointers
	booksToBulkInsert := make([]*Book, 0, len(setFoundation))
	for _, book := range setFoundation {
		// Note : don't simply use &book, it will bit you !
		bookToInsert := &Book{}
		*bookToInsert = book
		booksToBulkInsert = append(booksToBulkInsert, bookToInsert)
	}
	err = db.BulkInsert(&booksToBulkInsert).Do()
	if err != nil {
		t.Fatal(err)
	}
	if hasReturning(db) {
		for _, book := range booksToBulkInsert {
			if book.Id == 0 {
				t.Fatalf("Id was not set for the book %v", book)
			}
		}
	}
}

func structsSelectTest(db *godb.DB, t *testing.T) {
	// Count books
	count := CountBooks(t, db)
	if count != 7 {
		t.Fatalf("Wrong books count : %v", count)
	}

	// Fetch single book
	bilbo := Book{}
	err := db.Select(&bilbo).Where("title = ?", bookTheHobbit.Title).Do()
	if err != nil {
		t.Fatal(err)
	}
	if bilbo.Title != bookTheHobbit.Title {
		t.Fatalf("Book not found or wrong book : %v", bilbo)
	}

	// Fetch nonexistant book
	nonexistant := Book{}
	err = db.Select(&nonexistant).Where("title = ?", "Dune").Do()
	if err != sql.ErrNoRows {
		t.Fatalf("Error sql.ErrNoRows awaited, got : %v", err)
	}

	// Fetch multiple books
	theLordOfTheRing := make([]Book, 0, 0)
	err = db.Select(&theLordOfTheRing).
		Where("author = ?", authorTolkien).
		Where("title <> ?", bookTheHobbit.Title).
		Do()
	if err != nil {
		t.Fatal(err)
	}
	theLordOfTheRingSize := len(theLordOfTheRing)
	if len(theLordOfTheRing) != 3 {
		t.Fatalf("Wrong books count : %v", theLordOfTheRingSize)
	}

	// Multiple select in a transaction (force use of prepared statement)
	db.Begin()
	for _, book := range setFoundation {
		retrievedBook := Book{}
		err = db.Select(&retrievedBook).
			Where("title = ?", book.Title).
			Do()
		if err != nil {
			t.Fatal(err)
		}
		if retrievedBook.Title != book.Title {
			t.Fatalf("Book not found or wrong book : %v", retrievedBook)
		}
	}
	db.Commit()
}

func structsUpdateTest(db *godb.DB, t *testing.T) {
	// All the change will be rollbacked.
	db.Begin()

	booksToUpdate := make([]*Book, 0, 0)
	err := db.Select(&booksToUpdate).
		Where("author = ?", authorTolkien).
		Do()
	if err != nil {
		t.Fatal(err)
	}

	gandalf := "Gandalf the White"
	for _, book := range booksToUpdate {
		var count int64
		book.Author = gandalf
		count, err = db.Update(book).Do()
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("Wrong count of updated books : %v (book %v)", count, book)
		}
	}

	updatedBooks := make([]Book, 0, 0)
	booksCount, err := db.Select(&updatedBooks).
		Where("author = ?", gandalf).
		Count()
	if err != nil {
		t.Fatal(err)
	}
	if booksCount != 4 {
		t.Fatalf("Wrong books count : %v", booksCount)
	}

	// Cancel all changes
	db.Rollback()

	// The changes must be lost
	booksCount, err = db.Select(&updatedBooks).
		Where("author = ?", gandalf).
		Count()
	if err != nil {
		t.Fatal(err)
	}
	if booksCount != 0 {
		t.Fatalf("Wrong books count : %v", booksCount)
	}
}

func structsDeleteTest(db *godb.DB, t *testing.T) {
	bookToDelete := Book{}

	countBefore := CountBooks(t, db)

	err := db.Select(&bookToDelete).
		Where("published = ?", bookFoundation.Published).
		Do()
	if err != nil {
		t.Fatal(err)
	}

	var count int64
	count, err = db.Delete(&bookToDelete).Do()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Wrong deleted books count (from delete): %v", count)
	}

	countAfter := CountBooks(t, db)

	if (countBefore - countAfter) != 1 {
		t.Fatalf("Wrong deleted books count : %v", (countBefore - countAfter))
	}
}
