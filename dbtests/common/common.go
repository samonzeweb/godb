package common

import (
	"testing"

	"gitlab.com/samonzeweb/godb"
)

type Book struct {
	Id     int    `db:"id,key,auto"`
	Title  string `db:"title"`
	Author string `db:"author"`
}

func (*Book) TableName() string {
	return "books"
}

func MainTest(db *godb.DB, t *testing.T) {
	// Insert
	firstBook := &Book{
		Title:  "The Hobbit",
		Author: "J.R.R. tolkien",
	}

	err := db.Insert(firstBook).Do()
	if err != nil {
		t.Fatal(err)
	}
	if firstBook.Id == 0 {
		t.Fatal("Id not set after an insert")
	}

	titles := []string{
		"The Fellowship of the Ring",
		"The Two Towers",
		"The Return of the King",
	}

	for _, title := range titles {
		otherBook := &Book{
			Title:  title,
			Author: firstBook.Author,
		}
		err = db.Insert(otherBook).Do()
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

	// Select (many)
	theLordOfTheRingsBooks := make([]Book, 0, 0)
	err = db.Select(&theLordOfTheRingsBooks).Where("title <> ?", "The Hobbit").Do()
	if err != nil {
		t.Fatal(err)
	}
	if len(theLordOfTheRingsBooks) != 3 {
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
}
