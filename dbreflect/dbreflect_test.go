package dbreflect

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type simpleStruct struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text"`
	Other string
}

func TestStructMapping(t *testing.T) {
	Convey("NewStructMapping with a struct type", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(simpleStruct{}))

		Convey("It stores the struct name ", func() {
			So(structMap.Name, ShouldEndWith, "simpleStruct")
		})

		Convey("It store data about all tagged fields ", func() {
			So(len(structMap.GetAllColumnsNames()), ShouldEqual, 2)
			So(structMap.GetAllColumnsNames(), ShouldContain, "id")
			So(structMap.GetAllColumnsNames(), ShouldContain, "my_text")
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

func TestGetPointersForColumns(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := simpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetPointersForColumns return pointers corresponding to a given column name", func() {
			ptrs, err := structMap.GetPointersForColumns(&structInstance, "my_text")
			So(err, ShouldBeNil)
			So(len(ptrs), ShouldEqual, 1)
			So(ptrs[0], ShouldEqual, &(structInstance.Text))
		})

		Convey("GetPointersForColumns return pointers corresponding to given columns names", func() {
			ptrs, err := structMap.GetPointersForColumns(&structInstance, "id", "my_text")
			So(err, ShouldBeNil)
			So(len(ptrs), ShouldEqual, 2)
			So(ptrs[0], ShouldEqual, &(structInstance.ID))
			So(ptrs[1], ShouldEqual, &(structInstance.Text))
		})
	})
}

func TestFindFieldMapping(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := simpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("findFieldMapping return a fieldMapping for a given column name", func() {
			fm, err := structMap.findFieldMapping("my_text")
			So(err, ShouldBeNil)
			So(fm, ShouldNotBeNil)
			So(fm.sqlName, ShouldEqual, "my_text")
		})
	})
}
