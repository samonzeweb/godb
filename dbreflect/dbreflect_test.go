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

type structWithMultipeAutoKey struct {
	ID    int    `db:"id,key,auto"`
	Text  string `db:"my_text"`
	Other string `db:"other,key,auto"`
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

func TestGetAllColumnsNames(t *testing.T) {
	Convey("Given a struct mapping", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(simpleStruct{}))

		Convey("GetAllColumnsNames returns all columns names", func() {
			columns := structMap.GetAllColumnsNames()
			So(len(columns), ShouldEqual, 2)
			So(columns[0], ShouldEqual, "id")
			So(columns[1], ShouldEqual, "my_text")
		})
	})
}

func TestGetNonAutoColumnsNames(t *testing.T) {
	Convey("Given a struct mapping", t, func() {
		structMap, _ := NewStructMapping(reflect.TypeOf(simpleStruct{}))

		Convey("GetNonAutoColumnsNames returns non auto columns names", func() {
			columns := structMap.GetNonAutoColumnsNames()
			So(len(columns), ShouldEqual, 1)
			So(columns[0], ShouldEqual, "my_text")
		})
	})

}

func TestGetAllFieldsPointers(t *testing.T) {
	Convey("Given a struct mapping and an instance", t, func() {
		instance := simpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(instance))

		Convey("GetAllFieldsPointers returns all columns names", func() {
			pointers := structMap.GetAllFieldsPointers(&instance)
			So(len(pointers), ShouldEqual, 2)
			So(pointers[0], ShouldEqual, &(instance.ID))
			So(pointers[1], ShouldEqual, &(instance.Text))
		})
	})
}

func TestGetNonAutoFieldsValues(t *testing.T) {
	Convey("Given a struct mapping and an instance", t, func() {
		instance := simpleStruct{
			ID:    123,
			Text:  "Foo bar",
			Other: "Baz",
		}
		structMap, _ := NewStructMapping(reflect.TypeOf(instance))

		Convey("GetNonAutoFieldsValues returns values for non auto columns", func() {
			values := structMap.GetNonAutoFieldsValues(&instance)
			So(len(values), ShouldEqual, 1)
			So(values[0].(string), ShouldEqual, instance.Text)
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

		Convey("GetPointersForColumns returns pointers corresponding to given columns names", func() {
			ptrs, err := structMap.GetPointersForColumns(&structInstance, "id", "my_text")
			So(err, ShouldBeNil)
			So(len(ptrs), ShouldEqual, 2)
			So(ptrs[0], ShouldEqual, &(structInstance.ID))
			So(ptrs[1], ShouldEqual, &(structInstance.Text))
		})
	})
}

func TestGetAutoKeyPointer(t *testing.T) {
	Convey("Given a StructMapping and a struct instance", t, func() {
		structInstance := simpleStruct{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoKeyPointer returns pointer for the auto+key column", func() {
			ptr, err := structMap.GetAutoKeyPointer(&structInstance)
			So(err, ShouldBeNil)
			So(ptr, ShouldEqual, &(structInstance.ID))
		})
	})

	Convey("Given a StructMapping and a struct instance of multiple auto+key struct", t, func() {
		structInstance := structWithMultipeAutoKey{}
		structMap, _ := NewStructMapping(reflect.TypeOf(&structInstance))

		Convey("GetAutoKeyPointer returns pointer for the auto+key column", func() {
			ptr, err := structMap.GetAutoKeyPointer(&structInstance)
			So(err, ShouldNotBeNil)
			So(ptr, ShouldBeNil)
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
