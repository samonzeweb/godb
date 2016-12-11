package common

import (
	"testing"

	"github.com/samonzeweb/godb"
)

func CountBooks(t *testing.T, db *godb.DB) int64 {
	count, err := db.SelectFrom("books").Count()
	if err != nil {
		t.Fatal(err)
	}
	return count
}
