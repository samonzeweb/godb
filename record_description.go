package godb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/samonzeweb/godb/dbreflect"
)

// recordDescription describes the source or target of a SQL statement.
// The record (source or target) could be a struct pointer, or slice of structs,
// or a slice of pointers to structs.
type recordDescription struct {
	// record is always a pointer
	record            interface{}
	instanceType      reflect.Type
	structMapping     *dbreflect.StructMapping
	isSlice           bool
	isSliceOfPointers bool
}

// tableNamer wraps the TableName method, allowing a struct to specify a
// corresponding table name in database.
type tableNamer interface {
	TableName() string
}

// buildRecordDescription builds a recordDescription for the given object.
// Always use a pointer as argument.
func buildRecordDescription(record interface{}) (*recordDescription, error) {
	recordDesc := &recordDescription{}
	recordDesc.record = record

	recordType := reflect.TypeOf(record)
	if recordType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("Invalid argument, need a pointer, got a %s", recordType.Kind())
	}
	recordType = recordType.Elem()

	// A record could be a slice, or a single instance
	if recordType.Kind() == reflect.Slice {
		// Slice
		recordDesc.isSlice = true
		recordDesc.isSliceOfPointers = false
		recordType = recordType.Elem()
		if recordType.Kind() == reflect.Ptr {
			// Slice of pointers
			recordType = recordType.Elem()
			recordDesc.isSliceOfPointers = true
		}
	} else {
		// Single instance
		recordDesc.isSlice = false
		recordDesc.isSliceOfPointers = false
	}

	if recordType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Invalid argument, need a struct or structs slice, got a (or slice of) %s", recordType.Kind())
	}

	var err error
	recordDesc.instanceType = recordType
	recordDesc.structMapping, err = dbreflect.Cache.GetOrCreateStructMapping(recordType)
	if err != nil {
		return nil, err
	}

	return recordDesc, nil
}

// fillRecord build if needed new record instance and call the given function
// with the current record.
// If the record is a single instante it just use its pointer.
// If the recod is a slice, it creates new instances and adds it to the slice.
func (r *recordDescription) fillRecord(f func(record interface{}) error) error {
	if r.isSlice == false {
		return f(r.record)
	}

	// It's a slice
	// Create a new instance (reflect.Value of a pointer of the type needed)
	newInstancePointerValue := reflect.New(r.instanceType)
	newInstancePointer := newInstancePointerValue.Interface()
	// Call func with the struct pointer
	err := f(newInstancePointer)
	if err != nil {
		return err
	}
	// Add the new instance to the struct
	// Get the current slice (r.record is a slice pointer)
	sliceValue := reflect.ValueOf(r.record).Elem()
	// Add the new instance (or pointer to) into the slice
	instanceOrPointerValue := newInstancePointerValue
	if !r.isSliceOfPointers {
		instanceOrPointerValue = newInstancePointerValue.Elem()
	}
	newSliceValue := reflect.Append(sliceValue, instanceOrPointerValue)
	// Update the content of r.record with the new slice
	reflect.ValueOf(r.record).Elem().Set(newSliceValue)

	return nil
}

// getOneInstancePointer returns an instance pointers of the record (or record
// part) to be used for interface check and method call.
// Don't use the instance pointer for other use, don't change values,
// don't store it for later use, ...
func (r *recordDescription) getOneInstancePointer() interface{} {
	if r.isSlice == false {
		return r.record
	}

	return reflect.New(r.instanceType).Interface()
}

// len returns the len of the record.
// If it is a slice, it returns the slice length, otherwise it returns 1 (for
// a single instance).
func (r *recordDescription) len() int {
	if r.isSlice == false {
		return 1
	}
	return reflect.Indirect(reflect.ValueOf(r.record)).Len()
}

// index returns the pointer to the record having the given index.
func (r *recordDescription) index(i int) interface{} {
	if r.isSlice == false {
		return r.record
	}

	slice := reflect.Indirect(reflect.ValueOf(r.record))
	v := slice.Index(i)
	if v.Type().Kind() == reflect.Ptr {
		return v.Interface()
	}
	return v.Addr().Interface()
}

// getTableName returns the table name to use for the current record.
func (r *recordDescription) getTableName() string {
	p := r.getOneInstancePointer()
	if namer, ok := p.(tableNamer); ok {
		return namer.TableName()
	}

	typeNameParts := strings.Split(r.structMapping.Name, ".")
	return typeNameParts[len(typeNameParts)-1]
}
