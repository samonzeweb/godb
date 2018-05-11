package common

import (
	"database/sql"
	"time"
)

type Book struct {
	Id        int       `db:"id,key,auto"`
	Title     string    `db:"title"`
	Author    string    `db:"author"`
	Published time.Time `db:"published"`
	Version   int       `db:"version,oplock"`
}

func (*Book) TableName() string {
	return "books"
}

type Inventory struct {
	Id            int       `db:"id,key,auto"`
	BookId        int       `db:"book_id"`
	LastInventory time.Time `db:"last_inventory"`
	Counting      int       `db:"counting"`
}

func (*Inventory) TableName() string {
	return "inventories"
}

type InventoryPart struct {
	Id       sql.NullInt64 `db:"id"`
	Counting sql.NullInt64 `db:"counting"`
}

type BooksWithInventories struct {
	Book          `db:",rel=books"`
	InventoryPart `db:",rel=inventories"`
}

type CountByAuthor struct {
	Author string `db:"author"`
	Count  int    `db:"count"`
}

var authorTolkien = "J.R.R. tolkien"

var bookTheHobbit = Book{
	Title:     "The Hobbit",
	Author:    authorTolkien,
	Published: time.Date(1937, 9, 21, 0, 0, 0, 0, time.UTC),
}

var bookTheFellowshipOfTheRing = Book{
	Title:     "The Fellowship of the Ring",
	Author:    authorTolkien,
	Published: time.Date(1954, 7, 29, 0, 0, 0, 0, time.UTC),
}

var bookTheTwoTowers = Book{
	Title:     "The Two Towers",
	Author:    authorTolkien,
	Published: time.Date(1954, 11, 11, 0, 0, 0, 0, time.UTC),
}

var bookTheReturnOfTheKing = Book{
	Title:     "The Return of the King",
	Author:    authorTolkien,
	Published: time.Date(1955, 10, 20, 0, 0, 0, 0, time.UTC),
}

var setTheLordOfTheRing = []Book{
	bookTheFellowshipOfTheRing,
	bookTheTwoTowers,
	bookTheReturnOfTheKing,
}

var authorAssimov = "Isaac Assimov"

var bookFoundation = Book{
	Title:  "Foundation",
	Author: authorAssimov,
	// Don't know the exact date
	Published: time.Date(1951, 1, 1, 0, 0, 0, 0, time.UTC),
}

var bookFoundationAndEmpire = Book{
	Title:  "Foundation and Empire",
	Author: authorAssimov,
	// Don't know the exact date
	Published: time.Date(1952, 1, 1, 0, 0, 0, 0, time.UTC),
}

var bookSecondFoundation = Book{
	Title:  "Second Foundation",
	Author: authorAssimov,
	// Don't know the exact date
	Published: time.Date(1953, 1, 1, 0, 0, 0, 0, time.UTC),
}

var setFoundation = []Book{
	bookFoundation,
	bookFoundationAndEmpire,
	bookSecondFoundation,
}

var setAllBooks = []Book{
	bookTheHobbit,
	bookTheFellowshipOfTheRing,
	bookTheTwoTowers,
	bookTheReturnOfTheKing,
	bookFoundation,
	bookFoundationAndEmpire,
	bookSecondFoundation,
}
