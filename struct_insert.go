package godb

import (
	"database/sql"
	"fmt"

	"github.com/samonzeweb/godb/adapters"
	"github.com/samonzeweb/godb/types"
)

// StructInsert builds an INSERT statement for the given object.
//
// Example (book is a struct instance, books a slice) :
//
// 	 err := db.Insert(&book).Do()
//
// 	 err = db.BulkInsert(&books).Do()
type StructInsert struct {
	error             error
	insertStatement   *InsertStatement
	recordDescription *recordDescription
	whiteList         []string
	blackList         []string
}

// Insert initializes an INSERT sql statement for the given object.
func (db *DB) Insert(record interface{}) *StructInsert {
	si := db.buildInsert(record)

	if si.recordDescription.isSlice {
		si.error = fmt.Errorf("Insert accepts only a single instance, got a slice")
	}

	return si
}

// BulkInsert initializes an INSERT sql statement for a slice.
//
// Warning : not all databases are able to update the auto columns in the
// case of insert with multiple rows. Only adapters implementing the
// InsertReturningSuffix interface will have auto columns updated.
func (db *DB) BulkInsert(record interface{}) *StructInsert {
	si := db.buildInsert(record)

	if !si.recordDescription.isSlice {
		si.error = fmt.Errorf("BulkInsert accepts only a slice")
	}

	return si
}

// buildInsert initializes an insert sql statement for the given object, either
// a slice or a single instance.
// For internal use only.
func (db *DB) buildInsert(record interface{}) *StructInsert {
	var err error

	si := &StructInsert{}
	si.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		si.error = err
		return si
	}

	quotedTableName := db.adapter.Quote(db.defaultTableNamer(si.recordDescription.getTableName()))
	si.insertStatement = db.InsertInto(quotedTableName)
	return si
}

// Whitelist saves columns to be inserted from struct
// It adds columns to list each time it is called
// whitelist should not include auto key tagged columns
func (si *StructInsert) Whitelist(columns ...string) *StructInsert {
	si.whiteList = append(si.whiteList, columns...)
	return si
}

// WhitelistReset resets whiteList
func (si *StructInsert) WhitelistReset() *StructInsert {
	si.whiteList = nil
	return si
}

// Blacklist saves columns not to be inserted from struct
// It adds columns to list each time it is called. If a column defined in whitelist is
// also given in black list than that column will be blacklisted.
func (si *StructInsert) Blacklist(columns ...string) *StructInsert {
	si.blackList = append(si.blackList, columns...)
	return si
}

// BlacklistReset resets blacklist
func (si *StructInsert) BlacklistReset() *StructInsert {
	si.blackList = nil
	return si
}

// Do executes the insert statement.
//
// The behavior differs according to the adapter. If it implements the
// InsertReturningSuffixer interface it will use it and fill all auto fields
// of the given struct. Otherwise it only fills the key with LastInsertId.
//
// With BulkInsert the behavior changeq according to the adapter, see
// BulkInsert documentation for more information.
func (si *StructInsert) Do() error {
	if si.error != nil {
		return si.error
	}

	// Columns names
	var columns []string
	if len(si.whiteList) > 0 {
		columns = si.whiteList
	} else {
		columns = si.recordDescription.structMapping.GetNonAutoColumnsNames()
	}
	// Filter black listed columns
	i := 0
	for _, c := range si.blackList {
		i = 0
		for _, a := range columns {
			if a != c {
				columns[i] = a
				i++
			}
		}
		columns = columns[:i]
	}

	hasWB := (len(si.whiteList) + len(si.blackList)) > 0
	if !hasWB {
		si.insertStatement = si.insertStatement.Columns(si.insertStatement.db.quoteAll(columns)...)
	}
	// Values
	var values []interface{}
	len := si.recordDescription.len()
	wbColsSet := false
	for i := 0; i < len; i++ {
		currentRecord := si.recordDescription.index(i)
		if hasWB {
			if !wbColsSet { // order of old columns list and current values list may not be same so, set here:
				columns, values = si.recordDescription.structMapping.GetNonAutoFieldsValuesFiltered(currentRecord, columns, false)
				si.insertStatement = si.insertStatement.Columns(si.insertStatement.db.quoteAll(columns)...)
				wbColsSet = true
			} else {
				// as columns are already ordered, just get values in same order
				_, values = si.recordDescription.structMapping.GetNonAutoFieldsValuesFiltered(currentRecord, columns, true)
			}
		} else {
			values = si.recordDescription.structMapping.GetNonAutoFieldsValues(currentRecord)
		}
		si.insertStatement.Values(values...)
	}

	// Use a RETURNING (or similar) clause ?
	returningBuilder, ok := si.insertStatement.db.adapter.(adapters.ReturningBuilder)
	if ok {
		autoColumns := si.recordDescription.structMapping.GetAutoColumnsNames()
		si.insertStatement.Returning(returningBuilder.FormatForNewValues(autoColumns)...)
	}

	// Run
	if returningBuilder != nil {
		// the function which will return the pointers according to the given columns
		f := func(record interface{}, columns []string) ([]interface{}, error) {
			pointers, err := si.recordDescription.structMapping.GetAutoFieldsPointers(record)
			return pointers, err
		}
		_, err := si.insertStatement.doWithReturning(si.recordDescription, f)
		return err
	}

	// Case for adapters not implenting ReturningSuffix(), we use the
	// value given by LastInsertId() (through Do method)
	insertedID, err := si.insertStatement.Do()
	if err != nil {
		return err
	}

	// Bulk insert don't update ids with this adater, the insert was done,
	// without error, but the new ids are unknown.
	if si.recordDescription.isSlice {
		return nil
	}

	// Get the Id
	pointerToID, err := si.recordDescription.structMapping.GetAutoKeyPointer(si.recordDescription.record)
	if err != nil {
		return err
	}

	if pointerToID != nil {
		switch t := pointerToID.(type) {
		default:
			return fmt.Errorf("Not implemented type for key : %T", pointerToID)
		case *int:
			*t = int(insertedID)
		case *int8:
			*t = int8(insertedID)
		case *int16:
			*t = int16(insertedID)
		case *int32:
			*t = int32(insertedID)
		case *int64:
			*t = int64(insertedID)
		case *uint:
			*t = uint(insertedID)
		case *uint8:
			*t = uint8(insertedID)
		case *uint16:
			*t = uint16(insertedID)
		case *uint32:
			*t = uint32(insertedID)
		case *uint64:
			*t = uint64(insertedID)
		case *types.NullInt64:
			*t = types.ToNullInt64(insertedID)
		case *sql.NullInt64:
			*t = sql.NullInt64{Int64: insertedID, Valid: true}
		}
	}

	return nil
}
