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
const optionOpLock = "oplock"
const optionRelation = "rel"

// StructMapping contains the relation between a struct and database columns.
type StructMapping struct {
	Name          string
	structMapping structMappingDetails
	opLockSQLName string
	fieldCount    int
	keyCount      int
	autoCount     int
}

// innerStructMapping contains the details of a relation between a struct
// and database columns.
type structMappingDetails struct {
	name             string
	fieldsMapping    []fieldMapping
	subStructMapping []subStructMapping
}

// fieldMapping contains the relation between a field and a database column.
type fieldMapping struct {
	name     string
	kind     reflect.Kind
	sqlName  string
	isKey    bool
	isAuto   bool
	isOpLock bool
}

// subStructMapping contrains nested structs.
type subStructMapping struct {
	name          string
	prefix        string
	relation      string
	structMapping structMappingDetails
}

// NewStructMapping builds a StructMapping with a given reflect.Type.
func NewStructMapping(structInfo reflect.Type) (*StructMapping, error) {
	sm := &StructMapping{}
	var err error

	sm.structMapping, err = newStructMappingDetails(structInfo)
	if err != nil {
		return nil, err
	}

	sm.Name = sm.structMapping.name
	sm.setFieldsCount()

	err = sm.setOpLockField()
	if err != nil {
		return nil, err
	}

	return sm, nil
}

// newInnerStructMapping builds aninnerStructMapping innerStructMapping with a
// given reflect.Type.
func newStructMappingDetails(structInfo reflect.Type) (structMappingDetails, error) {
	var smd structMappingDetails

	if structInfo.Kind() == reflect.Ptr {
		structInfo = structInfo.Elem()
	}
	if structInfo.Kind() != reflect.Struct {
		return smd, fmt.Errorf("Invalid argument, need a struct, got a %s", structInfo.Kind())
	}

	smd.name = structInfo.PkgPath() + "." + structInfo.Name()
	smd.fieldsMapping = make([]fieldMapping, 0, structInfo.NumField())

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
			subStructMapping, err := smd.newSubStructMapping(fieldInfo)
			if err != nil {
				return smd, err
			}
			smd.subStructMapping = append(smd.subStructMapping, *subStructMapping)
		} else {
			// Map a field
			fieldMapping, err := smd.newFieldMapping(fieldInfo)
			if err != nil {
				return smd, err
			}
			smd.fieldsMapping = append(smd.fieldsMapping, *fieldMapping)
		}
	}

	return smd, nil
}

// newFieldMapping build a fieldMapping parsing tag content.
func (smd *structMappingDetails) newFieldMapping(structField reflect.StructField) (*fieldMapping, error) {
	fieldMapping := &fieldMapping{
		name: structField.Name,
		kind: structField.Type.Kind(),
	}

	tag := structField.Tag.Get(tagName)
	tagContent := strings.Split(tag, contentSeparator)
	for i, tagValue := range tagContent {
		tagContent[i] = strings.TrimSpace(tagValue)
	}

	// First value is always the sql column name
	var options map[string]string
	fieldMapping.sqlName, options = smd.tagData(structField.Tag)
	if len(fieldMapping.sqlName) < 1 {
		return nil, fmt.Errorf("Empty tag name for %s.%s", smd.name, fieldMapping.name)
	}

	_, fieldMapping.isAuto = options[optionAuto]
	_, fieldMapping.isKey = options[optionKey]
	_, fieldMapping.isOpLock = options[optionOpLock]

	return fieldMapping, nil
}

// newSubStructMapping build nested structs mapping.
func (smd *structMappingDetails) newSubStructMapping(structField reflect.StructField) (*subStructMapping, error) {
	structInfo := structField.Type

	// Mapping
	structMapping, err := newStructMappingDetails(structInfo)
	if err != nil {
		return nil, err
	}

	subStructMapping := &subStructMapping{
		name:          structField.Name,
		structMapping: structMapping,
	}

	// Optional prefix and relation
	var options map[string]string
	subStructMapping.prefix, options = smd.tagData(structField.Tag)
	if relation, ok := options[optionRelation]; ok {
		subStructMapping.relation = relation
	}

	return subStructMapping, nil
}

// tagData extracts tag data :
// * the first value is returned as is (column name or prefix)
// * others values are used to build a key,value map (options)
func (*structMappingDetails) tagData(tag reflect.StructTag) (string, map[string]string) {
	tagMaps := make(map[string]string)
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
		// Options (simple or key=value)
		optionsParts := strings.Split(value, "=")
		if len(optionsParts) == 1 {
			tagMaps[value] = ""
		} else {
			tagMaps[optionsParts[0]] = optionsParts[1]
		}
	}

	return firstValue, tagMaps
}

// setFieldsCount set all fields count (all, auto, keys)
func (sm *StructMapping) setFieldsCount() {
	sm.fieldCount = 0
	sm.autoCount = 0
	sm.keyCount = 0

	f := func(_ string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		sm.fieldCount++
		if fieldMapping.isAuto {
			sm.autoCount++
		}
		if fieldMapping.isKey {
			sm.keyCount++
		}
		return false, nil
	}

	sm.structMapping.traverseTree("", "", nil, f)
}

// setOpLockField searchs optimistic locking field an update the struct mapping
// with the op lock field data.
// It returns an error if there is more then one op lock field.
func (sm *StructMapping) setOpLockField() error {
	opLockFieldCount := 0

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if fieldMapping.isOpLock {
			opLockFieldCount++
			if opLockFieldCount > 1 {
				sm.opLockSQLName = ""
				return true, fmt.Errorf("There is more than one optimistic locking field in %s", sm.Name)
			}

			if !fieldMapping.isAuto && !isValidNonAutoOpLockFieldType(fieldMapping) {
				return true, fmt.Errorf("The field %s in the struct %s don't have a valid type for an non auto oplock field", fieldMapping.name, sm.Name)
			}

			sm.opLockSQLName = fullName
		}
		return false, nil
	}

	_, err := sm.structMapping.traverseTree("", "", nil, f)
	return err
}

// isValidNonAutoOpLockFieldType check if a field type (Kind) is valid for an
// optimistic locking field, non automatic.
func isValidNonAutoOpLockFieldType(fieldMapping *fieldMapping) bool {
	switch fieldMapping.kind {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return true
	}
	return false
}

// GetAllColumnsNames returns the names of all columns.
func (sm *StructMapping) GetAllColumnsNames() []string {
	columns := make([]string, 0, sm.fieldCount)

	f := func(fullName string, _ *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		columns = append(columns, fullName)
		return false, nil
	}
	sm.structMapping.traverseTree("", "", nil, f)

	return columns
}

// GetNonAutoColumnsNames returns the names of non auto columns.
func (sm *StructMapping) GetNonAutoColumnsNames() []string {
	columns := make([]string, 0, sm.fieldCount-sm.autoCount)

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if !fieldMapping.isAuto {
			columns = append(columns, fullName)
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", nil, f)

	return columns
}

// GetAutoColumnsNames returns the names of auto columns.
func (sm *StructMapping) GetAutoColumnsNames() []string {
	columns := make([]string, 0, sm.autoCount)

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if fieldMapping.isAuto {
			columns = append(columns, fullName)
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", nil, f)

	return columns
}

// GetKeyColumnsNames returns the names of key columns.
func (sm *StructMapping) GetKeyColumnsNames() []string {
	columns := make([]string, 0, sm.keyCount)

	f := func(fullName string, fieldMapping *fieldMapping, _ *reflect.Value) (stop bool, err error) {
		if fieldMapping.isKey {
			columns = append(columns, fullName)
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", nil, f)

	return columns
}

// GetAllFieldsPointers returns pointers for all fields, in the same order
// as GetAllColumnsNames.
func (sm *StructMapping) GetAllFieldsPointers(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointers := make([]interface{}, 0, sm.fieldCount)

	f := func(fullName string, _ *fieldMapping, value *reflect.Value) (stop bool, err error) {
		pointers = append(pointers, value.Addr().Interface())
		return false, nil
	}
	sm.structMapping.traverseTree("", "", &v, f)

	return pointers
}

// GetNonAutoFieldsValues returns values of non auto fields, in the same order
// as GetNonAutoColumnsNames.
func (sm *StructMapping) GetNonAutoFieldsValues(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	values := make([]interface{}, 0, sm.fieldCount-sm.autoCount)

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if !fieldMapping.isAuto {
			values = append(values, value.Interface())
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", &v, f)

	return values
}

// GetNonAutoFieldsValuesFiltered returns values of fields in filterColumns,
// if filterColumns is empty than returns values of non auto fields like `GetNonAutoFieldsValues` but
// as map
func (sm *StructMapping) GetNonAutoFieldsValuesFiltered(s interface{}, filterColumns []string) ([]string, []interface{}) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)
	ln := sm.fieldCount - sm.autoCount
	if len(filterColumns) > 0 {
		ln = len(filterColumns)
	}

	columns := make([]string, 0, ln)
	values := make([]interface{}, 0, ln)
	// Explicitly defined columns in filterColumns will be returned whether it is key column or not
	flt := func(isAuto bool, colName string) bool {
		for _, c := range filterColumns {
			if c == colName {
				return true
			}
		}
		if len(filterColumns) > 0 {
			return false
		}
		return !isAuto
	}
	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if flt(fieldMapping.isAuto, fieldMapping.sqlName) {
			columns = append(columns, fieldMapping.sqlName)
			values = append(values, value.Interface())
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", &v, f)

	return columns, values
}

// GetKeyFieldsValues returns values of key fields, in the same order
// as TestGetKeyColumnsNames.
func (sm *StructMapping) GetKeyFieldsValues(s interface{}) []interface{} {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	values := make([]interface{}, 0, sm.keyCount)

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if fieldMapping.isKey {
			values = append(values, value.Interface())
		}
		return false, nil
	}
	sm.structMapping.traverseTree("", "", &v, f)

	return values
}

// GetPointersForColumns returns pointers for the given instance and columns
// names.
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
	sm.structMapping.traverseTree("", "", &v, f)

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

	if _, err := sm.structMapping.traverseTree("", "", &v, f); err != nil {
		return nil, err
	}

	return autoKeyPointer, nil
}

// GetAutoFieldsPointers returns pointers of all auto fields
func (sm *StructMapping) GetAutoFieldsPointers(s interface{}) ([]interface{}, error) {
	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	pointers := make([]interface{}, 0, sm.autoCount)

	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if fieldMapping.isAuto {
			pointers = append(pointers, value.Addr().Interface())
		}
		return false, nil
	}

	if _, err := sm.structMapping.traverseTree("", "", &v, f); err != nil {
		return nil, err
	}

	return pointers, nil
}

// GetOpLockSQLFieldName returns the sql name of the optimistic locking field, or
// a blank string if there is none oplock field.
func (sm *StructMapping) GetOpLockSQLFieldName() string {
	return sm.opLockSQLName
}

// GetAndUpdateOpLockFieldValue returns the current value of the optimistic
// locking field, and update its value (unless it's an auto value updated by
// the database itself).
func (sm *StructMapping) GetAndUpdateOpLockFieldValue(s interface{}) (interface{}, error) {
	if sm.opLockSQLName == "" {
		return nil, fmt.Errorf("Struct %s can't update oplock field, there is no such field", sm.Name)
	}

	// TODO : check type
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)

	var currentFieldValue interface{}
	f := func(fullName string, fieldMapping *fieldMapping, value *reflect.Value) (stop bool, err error) {
		if fullName == sm.opLockSQLName {
			currentFieldValue = value.Interface()
			if !fieldMapping.isAuto {
				updateNonAutoOpLockField(value)
			}
			return true, nil
		}
		return false, nil
	}

	if _, err := sm.structMapping.traverseTree("", "", &v, f); err != nil {
		return nil, err
	}

	return currentFieldValue, nil
}

// updateNonAutoOpLockField updates the value of the optimistic locking field.
// It manages only types accepted by isValidNonAutoOpLockFieldType, and of
// course only non-auto oplock fields.
func updateNonAutoOpLockField(value *reflect.Value) {
	value.SetInt(value.Int() + 1)
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
// The callback is a treeExplorer and take 4 arguments :
//  * relation : the name of the current relation (of empty string if none)
// 	* fullName : the fill name of the SQL columns (using prefixes).
//  * fieldMapping : the fieldMapping of the field.
//  * value : the reflect.Value of the field (or nil if traverseTree got nil as startValue).
// The callback returns a boolean and an error. If the boolean is true, the walk is stopped.

func (smd *structMappingDetails) traverseTree(relation string, prefix string, startValue *reflect.Value, f treeExplorer) (bool, error) {
	var stopped bool
	var err error

	// Fields in in current StructMapping
	for _, fm := range smd.fieldsMapping {
		fullName := prefix + fm.sqlName
		if relation != "" {
			fullName = relation + "." + fullName
		}

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
	for _, sub := range smd.subStructMapping {
		var newRelation string
		if sub.relation == "" {
			newRelation = relation // no change
		} else {
			newRelation = sub.relation
		}

		if startValue != nil {
			structValue := startValue.FieldByName(sub.name)
			stopped, err = sub.structMapping.traverseTree(newRelation, prefix+sub.prefix, &structValue, f)
		} else {
			stopped, err = sub.structMapping.traverseTree(newRelation, prefix+sub.prefix, nil, f)
		}

		if stopped || err != nil {
			return stopped, err
		}
	}

	return false, nil
}
