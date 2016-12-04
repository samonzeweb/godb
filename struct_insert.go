package godb

import (
	"fmt"

	"gitlab.com/samonzeweb/godb/adapters"
)

// structInsert builds an INSERT statement for the given object.
type structInsert struct {
	Error             error
	insertStatement   *insertStatement
	recordDescription *recordDescription
}

// Insert initializes an insert sql statement for the given object.
func (db *DB) Insert(record interface{}) *structInsert {
	si := db.buildInsert(record)

	if si.recordDescription.isSlice {
		si.Error = fmt.Errorf("Insert accepts only a single instance, got a slice")
	}

	return si
}

// BuklInsert initializes an insert sql statement for a slice.
// Warning : not all databases are able to update the auto columns in the
//           case of insert with multiple rows. Only adapters implementing the
//           InsertReturningSuffix interface will have auto columns updated.
func (db *DB) BulkInsert(record interface{}) *structInsert {
	si := db.buildInsert(record)

	if !si.recordDescription.isSlice {
		si.Error = fmt.Errorf("BulkInsert accepts only a slice")
	}

	return si
}

// buildInsert initializes an insert sql statement for the given object, either
// a slice or a single instance.
// For internal use only.
func (db *DB) buildInsert(record interface{}) *structInsert {
	var err error

	si := &structInsert{}
	si.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		si.Error = err
		return si
	}

	quotedTableName := db.adapter.Quote(si.recordDescription.getTableName())
	si.insertStatement = db.InsertInto(quotedTableName)
	return si
}

// Do executes the insert statement.
//
// The behaviour differs according to the adapter. If it implements the
// InsertReturningSuffixer interface it will use it and fill all auto fields
// of the given struct. Otherwise it only fills the key with LastInsertId.
func (si *structInsert) Do() error {
	if si.Error != nil {
		return si.Error
	}

	// Columns names
	columns := si.recordDescription.structMapping.GetNonAutoColumnsNames()
	si.insertStatement = si.insertStatement.Columns(si.insertStatement.db.quoteAll(columns)...)

	// Values
	len := si.recordDescription.len()
	for i := 0; i < len; i++ {
		currentRecord := si.recordDescription.index(i)
		values := si.recordDescription.structMapping.GetNonAutoFieldsValues(currentRecord)
		si.insertStatement.Values(values...)
	}

	// Specifig suffix needed ?
	suffixer, ok := si.insertStatement.db.adapter.(adapters.InsertReturningSuffixer)
	if ok {
		autoColumns := si.recordDescription.structMapping.GetAutoColumnsNames()
		si.insertStatement.Suffix(suffixer.InsertReturningSuffix(autoColumns))
	}

	// Run
	if suffixer != nil {
		err := si.insertStatement.doWithReturning(si.recordDescription.record)
		return err
	}

	// Case for adapters not implenting InsertReturningSuffix(), we use the
	// value given by LastInsertId() (through Do method)
	insertedId, err := si.insertStatement.Do()
	if err != nil {
		return err
	}

	// Bulk insert don't update ids with this adater, the insert was done,
	// without error, but the new ids are unkonwn.
	if si.recordDescription.isSlice {
		return nil
	}

	// Get the Id
	pointerToId, err := si.recordDescription.structMapping.GetAutoKeyPointer(si.recordDescription.record)
	if err != nil {
		return err
	}

	if pointerToId != nil {
		switch t := pointerToId.(type) {
		default:
			return fmt.Errorf("Not implemented type for key : %T", pointerToId)
		case *int:
			*t = int(insertedId)
		case *int8:
			*t = int8(insertedId)
		case *int16:
			*t = int16(insertedId)
		case *int32:
			*t = int32(insertedId)
		case *int64:
			*t = int64(insertedId)
		case *uint:
			*t = uint(insertedId)
		case *uint8:
			*t = uint8(insertedId)
		case *uint16:
			*t = uint16(insertedId)
		case *uint32:
			*t = uint32(insertedId)
		case *uint64:
			*t = uint64(insertedId)
		}
	}

	return nil
}
