package common

import (
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
	//db.SetLogger(log.New(os.Stderr, "", 0))

	insertTest(db, t)
	selectTest(db, t)
	updateTest(db, t)
	deleteTest(db, t)
}

func insertTest(db *godb.DB, t *testing.T) {
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

func selectTest(db *godb.DB, t *testing.T) {
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

func updateTest(db *godb.DB, t *testing.T) {
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

func deleteTest(db *godb.DB, t *testing.T) {
	// TODO
}
