package dbreflect

import (
	"fmt"
	"reflect"
	"strings"
)

const tagName = "db"
const contentSeparator = ","

const optionPrefix = "prefix"
const optionKey = "key"
const optionAuto = "auto"

// StructMapping contains the relation between a struct and database columns
// TODO : change structs to have a root with cache info (like # fields, ...)
type StructMapping struct {
	Name             string
	fieldsMapping    []fieldMapping
	subStructMapping []subStructMapping
}

// fieldMapping contains the relation between a field and a database column
type fieldMapping struct {
	name    string
	sqlName string
	isKey   bool
	isAuto  bool
}

// subStructMapping contrains nested structs
type subStructMapping struct {
	name          string
	prefix        string
	structMapping StructMapping
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
		// Only non pointers are mapped
		if fieldInfo.Type.Kind() == reflect.Ptr {
			continue
		}
		// No tag, no mapping
		if _, ok := fieldInfo.Tag.Lookup(tagName); !ok {
			continue
		}

		// Some structs are scannable, like time.Time, or other registered types.
		// See RegisterScannableStruct.
		if fieldInfo.Type.Kind() == reflect.Struct && !isStructScannable(fieldInfo.Type.Name()) {
			// Map a sub struct
			subStructMapping, err := structMapping.newSubStructMapping(fieldInfo)
			if err != nil {
				return nil, err
			}
			structMapping.subStructMapping = append(structMapping.subStructMapping, *subStructMapping)
		} else {
			// Map a field
			fieldMapping, err := structMapping.newFieldMapping(fieldInfo)
			if err != nil {
				return nil, err
			}
			structMapping.fieldsMapping = append(structMapping.fieldsMapping, *fieldMapping)
		}
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
	var options map[string]bool
	fieldMapping.sqlName, options = sm.tagData(structField.Tag)
	if len(fieldMapping.sqlName) < 1 {
		return nil, fmt.Errorf("Empty tag name for %s.%s", sm.Name, fieldMapping.name)
	}

	_, fieldMapping.isAuto = options["auto"]
	_, fieldMapping.isKey = options["key"]

	return fieldMapping, nil
}

// newSubStructMapping build nested structs mapping
func (sm *StructMapping) newSubStructMapping(structField reflect.StructField) (*subStructMapping, error) {
	structInfo := structField.Type

	// Mapping
	structMapping, err := NewStructMapping(structInfo)
	if err != nil {
		return nil, err
	}

	subStructMapping := &subStructMapping{
		name:          structField.Name,
		structMapping: *structMapping,
	}

	// Optionnal prefix
	subStructMapping.prefix, _ = sm.tagData(structField.Tag)

	return subStructMapping, nil
}

// tagData extract tag data :
// * the first value is returned as is (column name or prefix)
// * others values are used to build a key,value map (options)
func (*StructMapping) tagData(tag reflect.StructTag) (string, map[string]bool) {
	tagMaps := make(map[string]bool)
	tagContent := strings.Split(tag.Get(tagName), contentSeparator)
	isFirstData := true
	var firstValue string
	for _, tagOption := range tagContent {
		value := strings.TrimSpace(tagOption)
		if isFirstData {
			isFirstData = false
			firstValue = value
			continue
		}
		// Options
		tagMaps[value] = true
	}

	return firstValue, tagMaps
}

// GetAllColumnsNames returns the names of all columns
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetAllColumnsNames() []string {
	columns := make([]string, 0, 0)

	f := func(fullName string, _ *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		columns = append(columns, fullName)
		return false, nil
	}
	sm.traverseTree("", nil, f)

	return columns
}

// GetNonAutoColumnsNames returns the names of non auto columns
// It is intended to be used for INSERT statements.
func (sm *StructMapping) GetNonAutoColumnsNames() []string {
	columns := make([]string, 0, 0)

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if !fieldMapping.isAuto {
			columns = append(columns, fullName)
		}
		return false, nil
	}
	sm.traverseTree("", nil, f)

	return columns
}

// GetAutoColumnsNames returns the names of auto columns
// It is intended to be used for INSERT statements with adapters
// like PostgreSQL
func (sm *StructMapping) GetAutoColumnsNames() []string {
	columns := make([]string, 0, 0)

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if fieldMapping.isAuto {
			columns = append(columns, fullName)
		}
		return false, nil
	}
	sm.traverseTree("", nil, f)

	return columns
}

// GetAllFieldsPointers returns pointers for all fields, in the same order
// as GetAllColumnsNames.
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetAllFieldsPointers(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointers := make([]interface{}, 0, 0)

	f := func(fullName string, _ *fieldMapping, value *reflect.Value) (stop bool, err error) {
		pointers = append(pointers, value.Addr().Interface())
		return false, nil
	}
	sm.traverseTree("", &v, f)

	return pointers
}

// GetNonAutoFieldsValues returns values of non auto fiels, in the same order
// as GetNonAutoColumnsNames.
// It is intended to be used for INSERT statements.
func (sm *StructMapping) GetNonAutoFieldsValues(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	values := make([]interface{}, 0, 0)

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if !fieldMapping.isAuto {
			values = append(values, value.Interface())
		}
		return false, nil
	}
	sm.traverseTree("", &v, f)

	return values
}

// GetPointersForColumns returns pointers for the given instance and columns
// names.
// It is intended to be used for SELECT statements.
func (sm *StructMapping) GetPointersForColumns(s interface{}, columns ...string) ([]interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointersMap := make(map[string]interface{})

	f := func(fullName string, _ *fieldMapping, value *reflect.Value) (stop bool, err error) {
		for _, columnName := range columns {
			if columnName == fullName {
				pointersMap[columnName] = value.Addr().Interface()
			}
		}
		return false, nil
	}

	// Explore the struct tree
	sm.traverseTree("", &v, f)

	// Returns pointers in the same order than names
	pointers := make([]interface{}, 0, len(columns))
	for _, columnName := range columns {
		pointer, ok := pointersMap[columnName]
		if !ok {
			return nil, fmt.Errorf("Unknown column name %s in struct %s", columnName, sm.Name)
		}
		pointers = append(pointers, pointer)
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

	var autoKeyPointer interface{}

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if fieldMapping.isKey && fieldMapping.isAuto {
			if autoKeyPointer != nil {
				return true, fmt.Errorf("Multiple auto+key fields for %s", sm.Name)
			}
			autoKeyPointer = value.Addr().Interface()
		}
		return false, nil
	}

	if _, err := sm.traverseTree("", &v, f); err != nil {
		return nil, err
	}

	return autoKeyPointer, nil
}

// GetAutoKeyPointer returns pointers of all auto fields
// It is intended to be used for INSERT statements in special cases
// like PostgreSQL
func (sm *StructMapping) GetAutoFieldsPointers(s interface{}) ([]interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointers := make([]interface{}, 0, 0)

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if fieldMapping.isAuto {
			pointers = append(pointers, value.Addr().Interface())
		}
		return false, nil
	}

	if _, err := sm.traverseTree("", &v, f); err != nil {
		return nil, err
	}

	return pointers, nil
}

// treeExplorer is a callback function for traverseTree, see below
type treeExplorer func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error)

// traverseTree traverses the structure tree of the mapping, calling a callback for each field.
// The arguments are
//  * prefix : the prefix for the current StructMapping, use "".
//  * startValue : the reflect.Value of the struct to explore, or nil.
//  * f : the treeExplorer callback.
// It returns a boolean and an error. The boolean is true if a callback has stopped the walk through the tree.
//
// The callback is a treeExplorer and take 3 arguments :
// 	* fullName : the fill name of the SQL columns (using prefixes).
//  * fieldMapping : the fieldMapping of the field.
//  * value : the reflect.Value of the field (or nil if traverseTree got nil as startValue).
// The callback returns a boolean and an error. If the boolean is true, the walk is stopped.

func (sm *StructMapping) traverseTree(prefix string, startValue *reflect.Value, f treeExplorer) (bool, error) {
	var stopped bool
	var err error

	// Fields in in current StructMapping
	for _, fm := range sm.fieldsMapping {
		fullName := prefix + fm.sqlName
		if startValue != nil {
			fieldValue := startValue.FieldByName(fm.name)
			stopped, err = f(fullName, &fm, &fieldValue)
		} else {
			stopped, err = f(fullName, &fm, nil)
		}
		if stopped || err != nil {
			return stopped, err
		}
	}

	// Nested structs
	for _, sub := range sm.subStructMapping {
		if startValue != nil {
			structValue := startValue.FieldByName(sub.name)
			stopped, err = sub.structMapping.traverseTree(prefix+sub.prefix, &structValue, f)
		} else {
			stopped, err = sub.structMapping.traverseTree(prefix+sub.prefix, nil, f)
		}

		if stopped || err != nil {
			return stopped, err
		}
	}

	return false, nil
}
