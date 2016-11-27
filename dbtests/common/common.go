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
	// Insert
	firstBook := &Book{
		Title:     "The Hobbit",
		Author:    "J.R.R. tolkien",
		Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
	}

	err := db.Insert(firstBook).Do()
	if err != nil {
		t.Fatal(err)
	}
	if firstBook.Id == 0 {
		t.Fatal("Id not set after an insert")
	}

	otherBooks := []Book{
		Book{
			Title:     "The Fellowship of the Ring",
			Published: time.Date(1954, 7, 29, 0, 0, 0, 0, time.UTC),
		},
		Book{
			Title:     "The Two Towers",
			Published: time.Date(1954, 11, 11, 0, 0, 0, 0, time.UTC),
		},
		Book{
			Title:     "The Return of the King",
			Published: time.Date(1955, 10, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, otherBook := range otherBooks {
		otherBook.Author = "J.R.R. tolkien"
		err = db.Insert(&otherBook).Do()
		if err != nil {
			t.Fatal(err)
		}
	}

	// Count
	howManyBooks, err := db.SelectFrom("books").Count()
	if err != nil {
		t.Fatal(err)
	}
	if howManyBooks != 4 {
		t.Fatal("Wrong books count : ", howManyBooks)
	}

	// Select (one)
	theHobbit := &Book{}
	err = db.Select(theHobbit).Where("title = ?", "The Hobbit").Do()
	if err != nil {
		t.Fatal(err)
	}
	if theHobbit.Title != "The Hobbit" {
		t.Fatal("Wrong books found : ", theHobbit.Title)
	}
	if firstBook.Published.Year() != 1937 {
		t.Fatalf("Wrong published time : %v", firstBook.Published)
	}

	// Select (many)
	theLordOfTheRingsBooks := make([]Book, 0, 0)
	err = db.Select(&theLordOfTheRingsBooks).Where("title <> ?", "The Hobbit").Do()
	if err != nil {
		t.Fatal(err)
	}
	if len(theLordOfTheRingsBooks) != 3 {
		t.Fatal("Wrong books count : ", howManyBooks)
	}

	// Select during a Tx (prepared statement will be used and cached)
	db.Begin()
	titleToFind := []string{
		"The Fellowship of the Ring",
		"The Two Towers",
		"The Return of the King",
	}
	for _, title := range titleToFind {
		var book Book
		err = db.Select(&book).Where("title = ?", title).Do()
		if err != nil {
			t.Fatal(err)
		}
		if book.Title != title {
			t.Fatal("Wrong books found : ", book.Title)
		}
	}
	db.Commit()
}
