package godb

import (
	"fmt"

	"github.com/samonzeweb/godb/adapters"
)

// StructUpdate builds an UPDATE statement for the given object.
//
// Example (book is a struct instance):
//
// 	 count, err := db.Update(&book).Do()
//
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
func (su *StructUpdate) Do() error {
	if su.error != nil {
		return su.error
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
		return fmt.Errorf("The object of type %T has no key : ", su.recordDescription.record)
	}
	for i, column := range keyColumns {
		quotedColumn := su.updateStatement.db.adapter.Quote(column)
		su.updateStatement = su.updateStatement.Where(quotedColumn+" = ?", keyValues[i])
	}

	// Optimistic Locking
	opLockColumn := su.recordDescription.structMapping.GetOpLockSQLFieldName()
	if opLockColumn != "" {
		opLockValue, err := su.recordDescription.structMapping.GetAndUpdateOpLockFieldValue(su.recordDescription.record)
		if err != nil {
			return err
		}
		su.updateStatement = su.updateStatement.Where(opLockColumn+" = ?", opLockValue)
	}

	// Specifig suffix needed ?
	suffixer, ok := su.updateStatement.db.adapter.(adapters.ReturningSuffixer)
	if ok {
		autoColumns := su.recordDescription.structMapping.GetAutoColumnsNames()
		su.updateStatement.Suffix(suffixer.ReturningSuffix(autoColumns))
	}

	var rowsAffected int64
	var err error

	if suffixer != nil {
		// the function which will return the pointers according to the given columns
		f := func(record interface{}, columns []string) ([]interface{}, error) {
			pointers, err := su.recordDescription.structMapping.GetAutoFieldsPointers(record)
			return pointers, err
		}
		// Case for adapters implenting ReturningSuffix()
		rowsAffected, err = su.updateStatement.doWithReturning(su.recordDescription, f)
	} else {
		// Case for adapters not implenting ReturningSuffix()
		rowsAffected, err = su.updateStatement.Do()
		if err != nil {
			return err
		}
	}

	if opLockColumn != "" && rowsAffected == 0 {
		err = ErrOpLock
	}

	return err
}
