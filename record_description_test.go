package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type typeToDescribe struct {
	Id int `db:"id"`
}

type otherTypeToDescribe struct {
	Id int `db:"id"`
}

func (*otherTypeToDescribe) TableName() string {
	return "others"
}

func TestbuildRecordDescription(t *testing.T) {
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
				record.(*typeToDescribe).Id = 123
				return nil
			})
			So(instancePtr.Id, ShouldEqual, 123)
		})
	})

	Convey("Given a slice descriptor ", t, func() {
		slice := make([]typeToDescribe, 0, 0)
		recordDesc, _ := buildRecordDescription(&slice)

		Convey("fillRecord call the given func with a new instance pointer", func() {
			recordDesc.fillRecord(func(record interface{}) error {
				So(record, ShouldHaveSameTypeAs, &typeToDescribe{})
				record.(*typeToDescribe).Id = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, typeToDescribe{})
			So(slice[0].Id, ShouldEqual, 123)
		})
	})

	Convey("Given a slice of pointers descriptor ", t, func() {
		slice := make([]*typeToDescribe, 0, 0)
		recordDesc, _ := buildRecordDescription(&slice)

		Convey("fillRecord call the given func with a new instance pointer", func() {
			recordDesc.fillRecord(func(record interface{}) error {
				So(record, ShouldHaveSameTypeAs, &typeToDescribe{})
				record.(*typeToDescribe).Id = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, &typeToDescribe{})
			So((*slice[0]).Id, ShouldEqual, 123)
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

func TestTableName(t *testing.T) {
	Convey("Given a record descriptor", t, func() {
		instancePtr := &typeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns by default the struct name a table name", func() {
			tableName := recordDesc.getTableName()
			So(tableName, ShouldEqual, "typeToDescribe")
		})
	})

	Convey("Given a record descriptor of type implmenting tableNamer interface", t, func() {
		instancePtr := &otherTypeToDescribe{}
		recordDesc, _ := buildRecordDescription(instancePtr)
		Convey("getTableName returns the string given by TableName()", func() {
			tableName := recordDesc.getTableName()
			So(tableName, ShouldEqual, "others")
		})
	})
}
