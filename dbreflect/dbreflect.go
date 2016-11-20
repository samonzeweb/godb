package dbreflect

import (
	"fmt"
	"reflect"
	"strings"
)

const tagName = "db"
const contentSeparator = ","

const optionKey = "key"
const optionAuto = "auto"
const optionOptimisticLocking = "oplock"

// StructMapping contains the relation between a struct and database columns
type StructMapping struct {
	Name          string
	fieldsMapping []fieldMapping
}

// fieldMapping contains the relation between a field and a database column
type fieldMapping struct {
	name     string
	sqlName  string
	isKey    bool
	isAuto   bool
	isOpLock bool
}

// NewStructMapping build a StructMapping with a given reflect.Type
func NewStructMapping(structInfo reflect.Type) (*StructMapping, error) {
	if structInfo.Kind() == reflect.Ptr {
		structInfo = structInfo.Elem()
	}
	if structInfo.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Invalid argument, need a struct, got a %s", structInfo.Kind())
	}

	structMapping := &StructMapping{
		Name:          structInfo.PkgPath() + "." + structInfo.Name(),
		fieldsMapping: make([]fieldMapping, 0, structInfo.NumField()),
	}

	for i := 0; i < structInfo.NumField(); i++ {
		fieldInfo := structInfo.Field(i)
		tag := fieldInfo.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		fieldMapping, err := structMapping.newFieldMapping(fieldInfo)
		if err != nil {
			return nil, err
		}
		structMapping.fieldsMapping = append(structMapping.fieldsMapping, *fieldMapping)
	}

	return structMapping, nil
}

// newFieldMapping build a fieldMapping parsing tag content
func (sm *StructMapping) newFieldMapping(structField reflect.StructField) (*fieldMapping, error) {
	fieldMapping := &fieldMapping{name: structField.Name}

	tag := structField.Tag.Get(tagName)
	tagContent := strings.Split(tag, contentSeparator)
	for i, tagValue := range tagContent {
		tagContent[i] = strings.TrimSpace(tagValue)
	}

	// First value is always the sql column name
	fieldMapping.sqlName = tagContent[0]
	if len(fieldMapping.sqlName) < 1 {
		return nil, fmt.Errorf("Empty tag name for %s.%s", sm.Name, fieldMapping.name)
	}

	// Parse options
	tagContent = tagContent[1:]
	fieldMapping.isAuto = isOptionPresent(tagContent, "auto")
	fieldMapping.isKey = isOptionPresent(tagContent, "key")
	fieldMapping.isOpLock = isOptionPresent(tagContent, "oplock")

	return fieldMapping, nil
}

// isOptionPresent check the presente of an option (string) in a list of options (string)
func isOptionPresent(tagValues []string, optionName string) bool {
	for _, option := range tagValues {
		if option == optionName {
			return true
		}
	}
	return false
}

// GetAllColumnsNames returns the names of all columns
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetAllColumnsNames() []string {
	columns := make([]string, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		columns = append(columns, fieldMapping.sqlName)
	}
	return columns
}

// GetNonAutoColumnsNames returns the names of non auto columns
// It is intended to be used for INSERT statements.
// TODO : manage oplock later
func (sm *StructMapping) GetNonAutoColumnsNames() []string {
	columns := make([]string, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		if fieldMapping.isAuto == false {
			columns = append(columns, fieldMapping.sqlName)
		}
	}
	return columns
}

// GetAllFieldsPointers returns pointers for all fields, in the same order
// as GetAllColumnsNames.
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetAllFieldsPointers(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointers := make([]interface{}, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		fieldValue := v.FieldByName(fieldMapping.name)
		pointers = append(pointers, fieldValue.Addr().Interface())
	}
	return pointers
}

// GetNonAutoFieldsValues returns values of non auto fiels, in the same order
// as GetNonAutoColumnsNames.
// It is intended to be used for INSERT statements.
func (sm *StructMapping) GetNonAutoFieldsValues(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	values := make([]interface{}, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		if fieldMapping.isAuto == false {
			fieldValue := v.FieldByName(fieldMapping.name)
			values = append(values, fieldValue.Interface())
		}
	}
	return values
}

// GetPointersForColumns returns pointers for the given instance and columns
// names.
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetPointersForColumns(s interface{}, columns ...string) ([]interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)
	pointers := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		fieldMapping, err := sm.findFieldMapping(column)
		if err != nil {
			return nil, err
		}
		fieldValue := v.FieldByName(fieldMapping.name)
		pointers = append(pointers, fieldValue.Addr().Interface())
	}
	return pointers, nil
}

// GetAutoKeyPointer returns a pointer for a key and auto columns.
// It will return nil if there is no such column, but no error.
// It will return an error if there is more than one auto and key column.
// It is intended to be used for INSERT statements.
func (sm *StructMapping) GetAutoKeyPointer(s interface{}) (interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	var pointer interface{}
	for _, fieldMapping := range sm.fieldsMapping {
		if fieldMapping.isAuto && fieldMapping.isKey {
			if pointer != nil {
				return nil, fmt.Errorf("Multiple auto+key fields for %s", sm.Name)
			}
			fieldValue := v.FieldByName(fieldMapping.name)
			pointer = fieldValue.Addr().Interface()
		}
	}

	return pointer, nil
}

// findFieldMapping find the fieldMapping instance for the given column name
func (sm *StructMapping) findFieldMapping(columnName string) (*fieldMapping, error) {
	for _, fm := range sm.fieldsMapping {
		if fm.sqlName == columnName {
			return &fm, nil
		}
	}

	return nil, fmt.Errorf("No field mapping for column %s", columnName)
}
