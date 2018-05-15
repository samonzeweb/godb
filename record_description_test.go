package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type typeToDescribe struct {
	ID int `db:"id"`
}

type otherTypeToDescribe struct {
	ID int `db:"id"`
}

func (*otherTypeToDescribe) TableName() string {
	return "others"
}

func TestBuildRecordDescription(t *testing.T) {
	Convey("Given a single instance pointer", t, func() {
		instance := &typeToDescribe{}

		Convey("extratType will extract the type information", func() {
			recordDesc, err := buildRecordDescription(instance)
			So(err, ShouldBeNil)
			So(recordDesc, ShouldNotBeNil)
			So(recordDesc.instanceType.Name(), ShouldEqual, "typeToDescribe")
			So(recordDesc.isSlice, ShouldBeFalse)
			So(recordDesc.isSliceOfPointers, ShouldBeFalse)
			So(recordDesc.structMapping.Name, ShouldEndWith, "typeToDescribe")
		})
	})

	Convey("Given a slice pointer", t, func() {
		slice := make([]typeToDescribe, 0, 0)

		Convey("extratType will extract the type information", func() {
			recordDesc, err := buildRecordDescription(&slice)
			So(err, ShouldBeNil)
			So(recordDesc, ShouldNotBeNil)
			So(recordDesc.instanceType.Name(), ShouldEqual, "typeToDescribe")
			So(recordDesc.isSlice, ShouldBeTrue)
			So(recordDesc.isSliceOfPointers, ShouldBeFalse)
			So(recordDesc.structMapping.Name, ShouldEndWith, "typeToDescribe")
		})
	})

	Convey("Given a slice pointer of pointers", t, func() {
		slice := make([]*typeToDescribe, 0, 0)

		Convey("extratType will extract the type information", func() {
			recordDesc, err := buildRecordDescription(&slice)
			So(err, ShouldBeNil)
			So(recordDesc, ShouldNotBeNil)
			So(recordDesc.instanceType.Name(), ShouldEqual, "typeToDescribe")
			So(recordDesc.isSlice, ShouldBeTrue)
			So(recordDesc.isSliceOfPointers, ShouldBeTrue)
			So(recordDesc.structMapping.Name, ShouldEndWith, "typeToDescribe")
		})
	})
}

func TestFillRecord(t *testing.T) {
	Convey("Given a single instance descriptor ", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)

		Convey("fillRecord call the given func with the instance pointer", func() {
			recordDesc.fillRecord(func(record interface{}) error {
				So(record, ShouldEqual, instancePtr)
				record.(*typeToDescribe).ID = 123
				return nil
			})
			So(instancePtr.ID, ShouldEqual, 123)
		})
	})

	Convey("Given a slice descriptor ", t, func() {
		slice := make([]typeToDescribe, 0, 0)
		recordDesc, _ := buildRecordDescription(&slice)

		Convey("fillRecord call the given func with a new instance pointer", func() {
			recordDesc.fillRecord(func(record interface{}) error {
				So(record, ShouldHaveSameTypeAs, &typeToDescribe{})
				record.(*typeToDescribe).ID = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, typeToDescribe{})
			So(slice[0].ID, ShouldEqual, 123)
		})
	})

	Convey("Given a slice of pointers descriptor ", t, func() {
		slice := make([]*typeToDescribe, 0, 0)
		recordDesc, _ := buildRecordDescription(&slice)

		Convey("fillRecord call the given func with a new instance pointer", func() {
			recordDesc.fillRecord(func(record interface{}) error {
				So(record, ShouldHaveSameTypeAs, &typeToDescribe{})
				record.(*typeToDescribe).ID = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, &typeToDescribe{})
			So((*slice[0]).ID, ShouldEqual, 123)
		})
	})
}

func TestGetOneInstancePointer(t *testing.T) {
	Convey("Given a single instance descriptor ", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getOneInstancePointer returns a pointer to the instance", func() {
			p := recordDesc.getOneInstancePointer()
			So(p, ShouldEqual, instancePtr)
		})
	})

	Convey("Given a slice descriptor ", t, func() {
		slice := make([]typeToDescribe, 0, 0)
		recordDesc, _ := buildRecordDescription(&slice)
		Convey("getOneInstancePointer returns a pointer to the instance", func() {
			p := recordDesc.getOneInstancePointer()
			So(p, ShouldHaveSameTypeAs, &typeToDescribe{})
		})
	})
}

func TestLen(t *testing.T) {
	Convey("Given a single instance descriptor", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("Len returns 1", func() {
			So(recordDesc.len(), ShouldEqual, 1)
		})
	})

	Convey("Given a slice descriptor", t, func() {
		slice := make([]typeToDescribe, 0, 0)
		slice = append(slice, typeToDescribe{})
		slice = append(slice, typeToDescribe{})
		recordDesc, _ := buildRecordDescription(&slice)
		Convey("Len returns the len of the slice", func() {
			So(recordDesc.len(), ShouldEqual, 2)
		})
	})
}

func TestIndex(t *testing.T) {
	Convey("Given a single instance descriptor", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("Index returns the pointer to the instance", func() {
			So(recordDesc.index(0), ShouldEqual, instancePtr)
		})
	})

	Convey("Given a slice descriptor", t, func() {
		slice := make([]typeToDescribe, 0, 0)
		first := typeToDescribe{ID: 123}
		second := typeToDescribe{ID: 456}
		slice = append(slice, first, second)
		recordDesc, _ := buildRecordDescription(&slice)
		Convey("Index returns the pointer to the instance at the given index", func() {
			So(recordDesc.index(0).(*typeToDescribe).ID, ShouldEqual, 123)
			So(recordDesc.index(1).(*typeToDescribe).ID, ShouldEqual, 456)
			recordDesc.index(0).(*typeToDescribe).ID = 1234
			So(recordDesc.index(0).(*typeToDescribe).ID, ShouldEqual, 1234)
		})
	})

	Convey("Given a slice of pointers descriptor", t, func() {
		slice := make([]*typeToDescribe, 0, 0)
		first := typeToDescribe{ID: 123}
		second := typeToDescribe{ID: 456}
		slice = append(slice, &first, &second)
		recordDesc, _ := buildRecordDescription(&slice)
		Convey("Index returns the pointer to the instance at the given index", func() {
			So(recordDesc.index(0), ShouldEqual, &first)
			So(recordDesc.index(1), ShouldEqual, &second)
		})
	})
}

func TestTableName(t *testing.T) {
	Convey("Given a record descriptor", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name", func() {
			tableName,_ := recordDesc.getTableName()
			So(tableName, ShouldEqual, "typeToDescribe")
		})
	})

	Convey("Given a record descriptor of type implmenting tableNamer interface", t, func() {
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName()", func() {
			tableName,_ := recordDesc.getTableName()
			So(tableName, ShouldEqual, "others")
		})
	})
}