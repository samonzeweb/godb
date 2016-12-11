package common

import "github.com/samonzeweb/godb"

func hasReturning(db *godb.DB) bool {
	return db.Adapter().DriverName() == "postgres"
}
