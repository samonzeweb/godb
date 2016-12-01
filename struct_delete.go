package godb

import "fmt"

// structDelete build a DELETE statement for the given object
type structDelete struct {
	Error             error
	deleteStatement   *deleteStatement
	recordDescription *recordDescription
}

// Delete initialise aa DELETE sql statement for the given object
func (db *DB) Delete(record interface{}) *structDelete {
	var err error

	sd := &structDelete{}
	sd.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		sd.Error = err
		return sd
	}

	if sd.recordDescription.isSlice {
		sd.Error = fmt.Errorf("Delete accept only a single instance, got a slice")
		return sd
	}

	quotedTableName := db.adapter.Quote(sd.recordDescription.getTableName())
	sd.deleteStatement = db.DeleteFrom(quotedTableName)
	return sd
}

// Do executes the DELETE statement for the struct given to the Delete method.
func (sd *structDelete) Do() (int64, error) {
	// Keys
	keyColumns := sd.recordDescription.structMapping.GetKeyColumnsNames()
	keyValues := sd.recordDescription.structMapping.GetKeyFieldsValues(sd.recordDescription.record)
	if len(keyColumns) == 0 {
		return 0, fmt.Errorf("The object of type %T has no key : ", sd.recordDescription.record)
	}
	for i, column := range keyColumns {
		quotedColumn := sd.deleteStatement.db.adapter.Quote(column)
		sd.deleteStatement = sd.deleteStatement.Where(quotedColumn+" = ?", keyValues[i])
	}

	// Executes the query
	rowsAffected, err := sd.deleteStatement.Do()
	return rowsAffected, err
}
