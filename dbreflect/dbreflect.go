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
const optionOptimisticLocking = "oplock"

// StructMapping contains the relation between a struct and database columns
type StructMapping struct {
	Name             string
	fieldsMapping    []fieldMapping
	subStructMapping []subStructMapping
}

// fieldMapping contains the relation between a field and a database column
type fieldMapping struct {
	name     string
	sqlName  string
	isKey    bool
	isAuto   bool
	isOpLock bool
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

		if fieldInfo.Type.Kind() == reflect.Struct {
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
	_, fieldMapping.isOpLock = options["oplock"]

	return fieldMapping, nil
}

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

//
func (sm *StructMapping) GetAllColumnsNames() []string {
	columns := make([]string, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		columns = append(columns, fieldMapping.sqlName)
	}
	return columns
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

//
func (sm *StructMapping) GetNonAutoColumnsNames() []string {
	columns := make([]string, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		if fieldMapping.isAuto == false {
			columns = append(columns, fieldMapping.sqlName)
		}
	}
	return columns
}

//
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

//
func (sm *StructMapping) GetAllFieldsValues(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	values := make([]interface{}, 0, len(sm.fieldsMapping))
	for _, fieldMapping := range sm.fieldsMapping {
		fieldValue := v.FieldByName(fieldMapping.name)
		values = append(values, fieldValue.Interface())
	}
	return values
}

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

func (sm *StructMapping) GetNonAutoColumnsNamesAndValues(s interface{}) map[string]interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	m := make(map[string]interface{})
	for _, fieldMapping := range sm.fieldsMapping {
		if fieldMapping.isAuto == false {
			m[fieldMapping.sqlName] = v.FieldByName(fieldMapping.name).Interface()
		}
	}
	return m
}

//
func (sm *StructMapping) GetPointersForColumns(s interface{}, columns ...string) ([]interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	// Find pointers
	pointersMap := make(map[string]interface{})
	f := func(fullName string, value *reflect.Value) (stop bool, err error) {
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

type treeExplorer func(fullName string, value *reflect.Value) (stop bool, err error)

// TODO
func (sm *StructMapping) traverseTree(prefix string, startValue *reflect.Value, f treeExplorer) (bool, error) {
	var stopped bool
	var err error

	// Fields in in current StructMapping
	for _, fm := range sm.fieldsMapping {
		fullName := prefix + fm.sqlName
		if startValue != nil {
			fieldValue := startValue.FieldByName(fm.name)
			stopped, err = f(fullName, &fieldValue)
		} else {
			stopped, err = f(fullName, nil)
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
