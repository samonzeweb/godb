package godb

import "fmt"

// StructUpdate builds an UPDATE statement for the given object.
type StructUpdate struct {
	error             error
	updateStatement   *UpdateStatement
	recordDescription *recordDescription
}

// Update initializes an UPDATE sql statement for the given object.
func (db *DB) Update(record interface{}) *StructUpdate {
	var err error

	su := &StructUpdate{}
	su.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		su.error = err
		return su
	}

	if su.recordDescription.isSlice {
		su.error = fmt.Errorf("Update accept only a single instance, got a slice")
		return su
	}

	quotedTableName := db.adapter.Quote(su.recordDescription.getTableName())
	su.updateStatement = db.UpdateTable(quotedTableName)
	return su
}

// Do executes the UPDATE statement for the struct given to the Update method.
func (su *StructUpdate) Do() (int64, error) {
	if su.error != nil {
		return 0, su.error
	}

	// Which columns to update ?
	columnsToUpdate := su.recordDescription.structMapping.GetNonAutoColumnsNames()
	values := su.recordDescription.structMapping.GetNonAutoFieldsValues(su.recordDescription.record)
	for i, column := range columnsToUpdate {
		quotedColumn := su.updateStatement.db.adapter.Quote(column)
		su.updateStatement = su.updateStatement.Set(quotedColumn, values[i])
	}

	// On wich keys
	keyColumns := su.recordDescription.structMapping.GetKeyColumnsNames()
	keyValues := su.recordDescription.structMapping.GetKeyFieldsValues(su.recordDescription.record)
	if len(keyColumns) == 0 {
		return 0, fmt.Errorf("The object of type %T has no key : ", su.recordDescription.record)
	}
	for i, column := range keyColumns {
		quotedColumn := su.updateStatement.db.adapter.Quote(column)
		su.updateStatement = su.updateStatement.Where(quotedColumn+" = ?", keyValues[i])
	}

	// Executes the query
	rowsAffected, err := su.updateStatement.Do()
	return rowsAffected, err
}
