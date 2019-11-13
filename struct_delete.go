package godb

import "fmt"

// StructDelete builds a DELETE statement for the given object.
//
// Example (book is a struct instance):
//
// 	 count, err := db.Delete(&book).Do()
//
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

	quotedTableName := db.quote(db.defaultTableNamer(sd.recordDescription.getTableName()))
	sd.deleteStatement = db.DeleteFrom(quotedTableName)
	return sd
}

// Do executes the DELETE statement for the struct given to the Delete method,
// and returns the count of deleted rows and an error.
func (sd *StructDelete) Do() (int64, error) {
	if sd.error != nil {
		return 0, sd.error
	}

	// Keys
	keyColumns := sd.recordDescription.structMapping.GetKeyColumnsNames()
	keyValues := sd.recordDescription.structMapping.GetKeyFieldsValues(sd.recordDescription.record)
	if len(keyColumns) == 0 {
		return 0, fmt.Errorf("the object of type %T has no key : ", sd.recordDescription.record)
	}
	for i, column := range keyColumns {
		quotedColumn := sd.deleteStatement.db.quote(column)
		sd.deleteStatement = sd.deleteStatement.Where(quotedColumn+" = ?", keyValues[i])
	}

	// Optimistic Locking
	opLockColumn := sd.recordDescription.structMapping.GetOpLockSQLFieldName()
	if opLockColumn != "" {
		opLockValue, err := sd.recordDescription.structMapping.GetAndUpdateOpLockFieldValue(sd.recordDescription.record)
		if err != nil {
			return 0, err
		}
		sd.deleteStatement = sd.deleteStatement.Where(opLockColumn+" = ?", opLockValue)
	}

	// Executes the query
	rowsAffected, err := sd.deleteStatement.Do()

	if opLockColumn != "" && rowsAffected == 0 {
		err = ErrOpLock
	}

	return rowsAffected, err
}
