package godb

import "fmt"

// structSelect build an INSERT statement for the given object
type structInsert struct {
	Error             error
	insertStatement   *insertStatement
	recordDescription *recordDescription
}

// Insert initialise an insert sql statement for the given object
func (db *DB) Insert(record interface{}) *structInsert {
	var err error

	si := &structInsert{}
	si.recordDescription, err = buildRecordDescription(record)
	if err != nil {
		si.Error = err
		return si
	}

	if si.recordDescription.isSlice {
		si.Error = fmt.Errorf("Insert accept only a single instance, got a slice")
		return si
	}

	si.insertStatement = db.InsertInto(si.recordDescription.getTableName())
	return si
}

// Do executes the insert statement
func (si *structInsert) Do() error {
	if si.Error != nil {
		return si.Error
	}

	// Columns names
	columns := si.recordDescription.structMapping.GetNonAutoColumnsNames()
	si.insertStatement = si.insertStatement.Columns(si.insertStatement.db.quoteAll(columns)...)

	// Values
	values := si.recordDescription.structMapping.GetNonAutoFieldsValues(si.recordDescription.record)
	si.insertStatement.Values(values...)

	// Run
	insertedId, err := si.insertStatement.Do()
	if err != nil {
		return err
	}

	// Get the Id
	// TODO : postgresql : get all auto fields
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
