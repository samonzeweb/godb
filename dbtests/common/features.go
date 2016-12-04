package common

import "gitlab.com/samonzeweb/godb"

func hasReturning(db *godb.DB) bool {
	return db.Adapter().DriverName() == "postgres"
}
