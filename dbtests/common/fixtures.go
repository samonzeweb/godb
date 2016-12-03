package common

import "time"

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
