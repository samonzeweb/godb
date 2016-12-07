package godb

import "fmt"

// StructDelete builds a DELETE statement for the given object.
type StructDelete struct {
	error             error
	deleteStatement   *DeleteStatement
	recordDescription *recordDescription
}

// Delete initializes a DELETE sql statement for the given object.
func (db *DB) Delete(record interface{}) *StructDelete {
	var err error

	sd := &StructDelete{}
	sd.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		sd.error = err
		return sd
	}

	if sd.recordDescription.isSlice {
		sd.error = fmt.Errorf("Delete accept only a single instance, got a slice")
		return sd
	}

	quotedTableName := db.adapter.Quote(sd.recordDescription.getTableName())
	sd.deleteStatement = db.DeleteFrom(quotedTableName)
	return sd
}

// Do executes the DELETE statement for the struct given to the Delete method.
func (sd *StructDelete) Do() (int64, error) {
	if sd.error != nil {
		return 0, sd.error
	}

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
