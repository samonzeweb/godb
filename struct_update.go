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
	whiteList         []string
	blackList         []string
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

	quotedTableName := db.adapter.Quote(db.defaultTableNamer(su.recordDescription.getTableName()))
	su.updateStatement = db.UpdateTable(quotedTableName)
	return su
}

// WhiteList saves columns to be updated from struct
func (su *StructUpdate) WhiteList(columns ...string) *StructUpdate {
	su.whiteList = append(su.whiteList, columns...)
	return su
}

// WhiteListReset resets whiteList
func (su *StructUpdate) WhiteListReset() *StructUpdate {
	su.whiteList = nil
	return su
}

// BlackList saves columns not to be updated from struct
// It adds columns to list each time it is called. If a column defined in whitelist is
// also given in black list than that column will be blacklisted.
func (su *StructUpdate) BlackList(columns ...string) *StructUpdate {
	su.blackList = append(su.blackList, columns...)
	return su
}

// BlackListReset resets blacklist
func (su *StructUpdate) BlackListReset() *StructUpdate {
	su.blackList = nil
	return su
}

// Do executes the UPDATE statement for the struct given to the Update method.
func (su *StructUpdate) Do() error {
	if su.error != nil {
		return su.error
	}

	// Which columns to update ?
	var columnsToUpdate []string
	if len(su.whiteList) > 0 {
		columnsToUpdate = su.whiteList
	} else {
		columnsToUpdate = su.recordDescription.structMapping.GetNonAutoColumnsNames()
	}
	// Filter black listed columns
	i := 0
	for _, c := range su.blackList {
		i = 0
		for _, a := range columnsToUpdate {
			if a != c {
				columnsToUpdate[i] = a
				i++
			}
		}
		columnsToUpdate = columnsToUpdate[:i]
	}

	values := su.recordDescription.structMapping.GetNonAutoFieldsValuesFiltered(su.recordDescription.record, columnsToUpdate)
	for _, column := range columnsToUpdate {
		quotedColumn := su.updateStatement.db.adapter.Quote(column)
		su.updateStatement = su.updateStatement.Set(quotedColumn, values[column])
	}

	// On which keys
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

	// Use a RETURNING (or similar) clause ?
	returningBuilder, ok := su.updateStatement.db.adapter.(adapters.ReturningBuilder)
	if ok {
		autoColumns := su.recordDescription.structMapping.GetAutoColumnsNames()
		su.updateStatement.Returning(returningBuilder.FormatForNewValues(autoColumns)...)
	}

	var rowsAffected int64
	var err error

	if returningBuilder != nil {
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
