package dbreflect

import (
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type SimpleStruct struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text"`
	Other string
}

type ComplexStruct struct {
	// without prefix, empty string is mandatory as it is a key,value
	// (see https://golang.org/pkg/reflect/#StructTag)
	SimpleStruct `db:""`
	// with prefix
	Foobar SubStruct `db:"nested_"`
	// ignored
	Ignored SubStruct
	// not nested field
	IAmNotNested string `db:"iamnotnested"`
}

type SubStruct struct {
	Foo string `db:"foo"`
	Bar string `db:"bar"`
}

type BadStructMultipleAutoKey struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text,key,auto"`
	Other string
}

type StructMultipleAuto struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text,auto"`
	Other string
}

type StructWithScannableStruct struct {
	ID   int       `db:"id,key,auto"`
	Time time.Time `db:"a_day"`
}

type StructWithOptimiticLocking struct {
	ID      int    `db:"id,key,auto"`
	Text    string `db:"my_text,auto"`
	Version int    `db:"version,oplock"`
}

type StructWithInvalidOptimiticLocking struct {
	ID         int    `db:"id,key,auto"`
	Text       string `db:"my_text,auto"`
	BadVersion string `db:"version,oplock"`
}

type ComplexStructsWithRelations struct {
	// no prefix but a relation
	SimpleStruct `db:",rel=firsttable"`
	// prefix and relation
	Foobar SubStruct `db:"nested_,rel=secondtable"`		
}

func TestStructMapping(t *testing.T) {
	Convey("NewStructMapping with a struct type", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(SimpleStruct{}))

		Convey("It stores the struct name ", func() {
			So(structMap.Name, ShouldEndWith, "SimpleStruct")
		})

		Convey("It store data about all tagged fields ", func() {
			So(len(structMap.GetAllColumnsNames()), ShouldEqual, 2)
			So(structMap.GetAllColumnsNames(), ShouldContain, "id")
			So(structMap.GetAllColumnsNames(), ShouldContain, "my_text")
		})
	})

	Convey("NewStructMapping with a complex struct type with sub-structs", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(ComplexStruct{}))
		structMapDetails := structMap.structMapping
		So(len(structMapDetails.subStructMapping), ShouldEqual, 2)

		Convey("It store data about sub struct without prefix ", func() {
			So(structMapDetails.subStructMapping[0].prefix, ShouldEqual, "")
			So(structMapDetails.subStructMapping[0].structMapping.name, ShouldEndWith, "SimpleStruct")
		})

		Convey("It store data about sub struct with prefix ", func() {
			So(structMapDetails.subStructMapping[1].prefix, ShouldEqual, "nested_")
			So(structMapDetails.subStructMapping[1].structMapping.name, ShouldEndWith, "SubStruct")
		})
	})

	Convey("NewStructMapping with nested structs and relations", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(ComplexStructsWithRelations{}))
		structMapDetails := structMap.structMapping
		So(len(structMapDetails.subStructMapping), ShouldEqual, 2)

		Convey("It store data about sub struct without prefix ", func() {
			So(structMapDetails.subStructMapping[0].prefix, ShouldEqual, "")
			So(structMapDetails.subStructMapping[0].relation, ShouldEndWith, "firsttable")
		})

		Convey("It store data about sub struct with prefix ", func() {
			So(structMapDetails.subStructMapping[1].prefix, ShouldEqual, "nested_")
			So(structMapDetails.subStructMapping[1].relation, ShouldEndWith, "secondtable")
		})
	})	
}

func TestScannableStructs(t *testing.T) {
	Convey("Calling NewStructMapping with a struct ", t, func() {
		structWithScannableStruct := StructWithScannableStruct{}
		structMap, err := NewStructMapping(reflect.TypeOf(structWithScannableStruct))
		So(err, ShouldBeNil)

		Convey("NewStructMapping consider time.Time as a field", func() {
			So(structMap.GetNonAutoColumnsNames()[0], ShouldEqual, "a_day")
		})
	})
}

func TestNewStructMappingErrors(t *testing.T) {
	Convey("Calling NewStructMapping without a struct ", t, func() {
		dummy := true
		_, err := NewStructMapping(reflect.TypeOf(dummy))

		Convey("NewStructMapping return an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetAllColumnsNames(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := SimpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetPointersForColumns return all columns names (not nested)", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 2)
			So(columns[0], ShouldEqual, "id")
			So(columns[1], ShouldEqual, "my_text")
		})
	})

	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAllColumnsNames return all columns names (nested)", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 5)
			So(columns[0], ShouldEqual, "iamnotnested")
			So(columns[1], ShouldEqual, "id")
			So(columns[2], ShouldEqual, "my_text")
			So(columns[3], ShouldEqual, "nested_foo")
			So(columns[4], ShouldEqual, "nested_bar")
		})
	})

	Convey("Given a StructMapping and nested structs with relations", t, func() {
		structInstance := ComplexStructsWithRelations{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAllColumnsNames return all columns names with relations", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 4)
			So(columns[0], ShouldEqual, "firsttable.id")
			So(columns[1], ShouldEqual, "firsttable.my_text")
			So(columns[2], ShouldEqual, "secondtable.nested_foo")
			So(columns[3], ShouldEqual, "secondtable.nested_bar")
		})
	})	
}

func TestGetNonAutoColumnsNames(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetNonAutoColumnsNames returns all non auto columns names", func() {
			columns := structMap.GetNonAutoColumnsNames()
			So(len(columns), ShouldEqual, 4)
			So(columns[0], ShouldEqual, "iamnotnested")
			So(columns[1], ShouldEqual, "my_text")
			So(columns[2], ShouldEqual, "nested_foo")
			So(columns[3], ShouldEqual, "nested_bar")
		})
	})
}

func TestGetAutoColumnsNames(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := StructMultipleAuto{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoColumnsNames returns all auto columns names", func() {
			columns := structMap.GetAutoColumnsNames()
			So(len(columns), ShouldEqual, 2)
			So(columns[0], ShouldEqual, "id")
			So(columns[1], ShouldEqual, "my_text")
		})
	})
}

func TestGetKeyColumnsNames(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := StructMultipleAuto{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetKeyColumnsNames returns all auto columns names", func() {
			columns := structMap.GetKeyColumnsNames()
			So(len(columns), ShouldEqual, 1)
			So(columns[0], ShouldEqual, "id")
		})
	})
}

func TestGetAllFieldsPointers(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAllFieldsPointers returns all fields pointers", func() {
			ptrs := structMap.GetAllFieldsPointers(&structInstance)
			So(len(ptrs), ShouldEqual, 5)
			So(ptrs[0], ShouldEqual, &(structInstance.IAmNotNested))
			So(ptrs[1], ShouldEqual, &(structInstance.ID))
			So(ptrs[2], ShouldEqual, &(structInstance.Text))
			So(ptrs[3], ShouldEqual, &(structInstance.Foobar.Foo))
			So(ptrs[4], ShouldEqual, &(structInstance.Foobar.Bar))
		})
	})
}

func TestGetNonAutoFieldsValues(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{
			SimpleStruct: SimpleStruct{
				ID:   1,
				Text: "a text",
			},
			Foobar: SubStruct{
				Foo: "FOO",
				Bar: "BAR",
			},
			IAmNotNested: "not, i'm not",
		}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetNonAutoFieldsValues return non auto fields values", func() {
			values := structMap.GetNonAutoFieldsValues(&structInstance)
			So(len(values), ShouldEqual, 4)
			So(values[0], ShouldEqual, structInstance.IAmNotNested)
			So(values[1], ShouldEqual, structInstance.Text)
			So(values[2], ShouldEqual, structInstance.Foobar.Foo)
			So(values[3], ShouldEqual, structInstance.Foobar.Bar)
		})
	})
}

func TestGetKeyFieldsValues(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := SimpleStruct{
			ID:    123,
			Text:  "a text",
			Other: "what ?",
		}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetKeyFieldsValues return non auto fields values", func() {
			values := structMap.GetKeyFieldsValues(&structInstance)
			So(len(values), ShouldEqual, 1)
			So(values[0], ShouldEqual, structInstance.ID)
		})
	})
}

func TestGetPointersForColumns(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := SimpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetPointersForColumns return pointers corresponding to given columns names (not nested)", func() {
			ptrs, err := structMap.GetPointersForColumns(&structInstance, "id", "my_text")
			So(err, ShouldBeNil)
			So(len(ptrs), ShouldEqual, 2)
			So(ptrs[0], ShouldEqual, &(structInstance.ID))
			So(ptrs[1], ShouldEqual, &(structInstance.Text))
		})
	})

	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetPointersForColumns return pointers corresponding to a given column name", func() {
			ptrs, err := structMap.GetPointersForColumns(&structInstance, "my_text", "nested_bar", "iamnotnested")
			So(err, ShouldBeNil)
			So(len(ptrs), ShouldEqual, 3)
			So(ptrs[0], ShouldEqual, &(structInstance.Text))
			So(ptrs[1], ShouldEqual, &(structInstance.Foobar.Bar))
			So(ptrs[2], ShouldEqual, &(structInstance.IAmNotNested))
		})
	})
}

func TestGetAutoKeyPointer(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoKeyPointer returns a pointer to the columns which is key and auto", func() {
			pointer, err := structMap.GetAutoKeyPointer(&structInstance)
			So(err, ShouldBeNil)
			So(pointer, ShouldEqual, &(structInstance.ID))
		})
	})

	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := BadStructMultipleAutoKey{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoKeyPointer returns an error if ther eare multiple auto+key fields", func() {
			pointer, err := structMap.GetAutoKeyPointer(&structInstance)
			So(err, ShouldNotBeNil)
			So(pointer, ShouldBeNil)
		})
	})
}

func TestGetAutoFieldsPointers(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := StructMultipleAuto{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoFieldsPointers returns all auto fields pointers", func() {
			pointers, err := structMap.GetAutoFieldsPointers(&structInstance)
			So(err, ShouldBeNil)
			So(len(pointers), ShouldEqual, 2)
			So(pointers[0], ShouldEqual, &(structInstance.ID))
			So(pointers[1], ShouldEqual, &(structInstance.Text))
		})
	})
}

func TestOpLockFieldName(t *testing.T) {
	Convey("Given a StructMapping with an optimistic locking field", t, func() {
		structWithOpLockField := StructWithOptimiticLocking{}
		structMap, err := NewStructMapping(reflect.TypeOf(structWithOpLockField))
		So(err, ShouldBeNil)

		Convey("NewStructMapping detects optimictic locking field", func() {
			sqlFieldName := structMap.GetOpLockSQLFieldName()
			So(sqlFieldName, ShouldEqual, "version")
		})
	})
}

func TestGetAndUpdateOpLockField(t *testing.T) {
	Convey("Given a StructMapping with an optimistic locking field", t, func() {
		structWithOpLockField := StructWithOptimiticLocking{}
		structMap, err := NewStructMapping(reflect.TypeOf(structWithOpLockField))
		So(err, ShouldBeNil)

		Convey("NewStructMapping detects optimictic locking field", func() {
			structWithOpLockField.Version = 123
			currentValue, err := structMap.GetAndUpdateOpLockFieldValue(&structWithOpLockField)
			So(err, ShouldBeNil)
			So(currentValue, ShouldEqual, 123)
			So(structWithOpLockField.Version, ShouldEqual, 124)
		})
	})
}

func TestInvalidOpLockFieldType(t *testing.T) {
	Convey("Given a StructMapping with an invalid optimistic locking field", t, func() {
		structWithInvalidOpLockField := StructWithInvalidOptimiticLocking{}
		_, err := NewStructMapping(reflect.TypeOf(structWithInvalidOpLockField))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "type")

	})
}
