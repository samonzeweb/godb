package common

import (
	"testing"

	"github.com/samonzeweb/godb"
)

func RawSQLTests(db *godb.DB, t *testing.T) {
	// Enable logger if needed
	//db.SetLogger(log.New(os.Stderr, "", 0))

	// Fixtures
	booksToInsert := setAllBooks[:]
	err := db.BulkInsert(&booksToInsert).Do()
	if err != nil {
		t.Fatal(err)
	}
	if getReturningBuilder(db) != nil {
		for _, book := range booksToInsert {
			if book.Id == 0 {
				t.Fatalf("Id was not set for the book %v", book)
			}
		}
	}

	// Test & assertions
	books := make([]Book, 0, 0)
	err = db.RawSQL("select * from books where author = ?", authorAssimov).Do(&books)
	if err != nil {
		t.Fatal(err)
	}

	if len(books) != len(setFoundation) {
		t.Fatalf("Wrong books count : %d", len(books))
	}
}
