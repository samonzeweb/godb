package godb

import "fmt"

// structUpdate build an UPDATE statement for the given object
type structUpdate struct {
	Error             error
	updateStatement   *updateStatement
	recordDescription *recordDescription
}

// Insert initialise an insert sql statement for the given object
func (db *DB) Update(record interface{}) *structUpdate {
	var err error

	su := &structUpdate{}
	su.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		su.Error = err
		return su
	}

	if su.recordDescription.isSlice {
		su.Error = fmt.Errorf("Update accept only a single instance, got a slice")
		return su
	}

	quotedTableName := db.adapter.Quote(su.recordDescription.getTableName())
	su.updateStatement = db.UpdateTable(quotedTableName)
	return su
}

// Do executes the UPDATE statement for the struct given to the Update method.
func (su *structUpdate) Do() (int64, error) {
	// Find non key columns => SET
	// Find key coluns => WHERE

}
