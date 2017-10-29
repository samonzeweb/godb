package common

import (
	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters"
)

func getReturningBuilder(db *godb.DB) adapters.ReturningBuilder {
	returningBuilder, _ := db.Adapter().(adapters.ReturningBuilder)
	return returningBuilder
}
