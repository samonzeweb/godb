package common

import (
	"log"
	"os"
	"testing"
	"time"

	"gitlab.com/samonzeweb/godb"
)

type Book struct {
	Id        int       `db:"id,key,auto"`
	Title     string    `db:"title"`
	Author    string    `db:"author"`
	Published time.Time `db:"published"`
}

func (*Book) TableName() string {
	return "books"
}

func MainTest(db *godb.DB, t *testing.T) {
	// Enable logger if needed
	db.SetLogger(log.New(os.Stderr, "", 0))

	Insert(db, t)
	Select(db, t)
}

func Insert(db *godb.DB, t *testing.T) {
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
	if db.Adapter().DriverName() == "postgres" {
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
	if db.Adapter().DriverName() == "postgres" {
		for _, book := range booksToBulkInsert {
			if book.Id == 0 {
				t.Fatalf("Id was not set for the book %v", book)
			}
		}
	}
}

func Select(db *godb.DB, t *testing.T) {
	// Count the books
	count, err := db.SelectFrom("books").Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 7 {
		t.Fatalf("Wrong book count : %v", count)
	}

	// Fetch single book
	bilbo := Book{}
	err = db.Select(&bilbo).Where("title = ?", bookTheHobbit.Title).Do()
	if err != nil {
		t.Fatal(err)
	}
	if bilbo.Title != bookTheHobbit.Title {
		t.Fatalf("Book not found or wrong book : %v", bilbo)
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
		t.Fatalf("Wrong book count : %v", theLordOfTheRingSize)
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
