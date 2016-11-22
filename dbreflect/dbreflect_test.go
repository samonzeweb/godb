package dbreflect

import (
	"reflect"
	"testing"

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
	// without prefix
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
		So(len(structMap.subStructMapping), ShouldEqual, 2)

		Convey("It store data about sub struct without prefix ", func() {
			So(structMap.subStructMapping[0].prefix, ShouldEqual, "")
			So(structMap.subStructMapping[0].structMapping.Name, ShouldEndWith, "SimpleStruct")
		})

		Convey("It store data about sub struct with prefix ", func() {
			So(structMap.subStructMapping[1].prefix, ShouldEqual, "nested_")
			So(structMap.subStructMapping[1].structMapping.Name, ShouldEndWith, "SubStruct")
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

		Convey("GetPointersForColumns return pointers corresponding to given columns names (not nested)", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 2)
			So(columns[0], ShouldEqual, "id")
			So(columns[1], ShouldEqual, "my_text")
		})
	})

	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAllColumnsNames return pointers corresponding to a given column name", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 5)
			So(columns[0], ShouldEqual, "iamnotnested")
			So(columns[1], ShouldEqual, "id")
			So(columns[2], ShouldEqual, "my_text")
			So(columns[3], ShouldEqual, "nested_foo")
			So(columns[4], ShouldEqual, "nested_bar")
		})
	})
}

func TestGetNonAutoColumnsNames(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetNonAutoColumnsNames return pointers corresponding to a given column name", func() {
			columns := structMap.GetNonAutoColumnsNames()
			So(len(columns), ShouldEqual, 4)
			So(columns[0], ShouldEqual, "iamnotnested")
			So(columns[1], ShouldEqual, "my_text")
			So(columns[2], ShouldEqual, "nested_foo")
			So(columns[3], ShouldEqual, "nested_bar")
		})
	})
}

func TestGetAllFieldsPointers(t *testing.T) {
	Convey("Given a StructMapping and a struct instance (nested)", t, func() {
		structInstance := ComplexStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAllFieldsPointers return pointers corresponding to a given column name", func() {
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

		Convey("GetNonAutoFieldsValues return pointers corresponding to a given column name", func() {
			values := structMap.GetNonAutoFieldsValues(&structInstance)
			So(len(values), ShouldEqual, 4)
			So(values[0], ShouldEqual, structInstance.IAmNotNested)
			So(values[1], ShouldEqual, structInstance.Text)
			So(values[2], ShouldEqual, structInstance.Foobar.Foo)
			So(values[3], ShouldEqual, structInstance.Foobar.Bar)
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

		Convey("GetAutoKeyPointer returns a pointer to the columns wich is key and auto", func() {
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
