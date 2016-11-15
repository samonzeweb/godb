package godb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type typeToExtract struct {
	Id int `db:"id"`
}

func TestExtratType(t *testing.T) {
	Convey("Given a single instance pointer", t, func() {
		instance := &typeToExtract{}

		Convey("extratType will extract the type information", func() {
			typeDesc, err := extractType(instance)
			So(err, ShouldBeNil)
			So(typeDesc, ShouldNotBeNil)
			So(typeDesc.InstanceType.Name(), ShouldEqual, "typeToExtract")
			So(typeDesc.IsSlice, ShouldBeFalse)
			So(typeDesc.IsSliceOfPointers, ShouldBeFalse)
			So(typeDesc.StructMapping.Name, ShouldEndWith, "typeToExtract")
		})
	})

	Convey("Given a slice pointer", t, func() {
		slice := make([]typeToExtract, 0, 0)

		Convey("extratType will extract the type information", func() {
			typeDesc, err := extractType(&slice)
			So(err, ShouldBeNil)
			So(typeDesc, ShouldNotBeNil)
			So(typeDesc.InstanceType.Name(), ShouldEqual, "typeToExtract")
			So(typeDesc.IsSlice, ShouldBeTrue)
			So(typeDesc.IsSliceOfPointers, ShouldBeFalse)
			So(typeDesc.StructMapping.Name, ShouldEndWith, "typeToExtract")
		})
	})

	Convey("Given a slice pointer of pointers", t, func() {
		slice := make([]*typeToExtract, 0, 0)

		Convey("extratType will extract the type information", func() {
			typeDesc, err := extractType(&slice)
			So(err, ShouldBeNil)
			So(typeDesc, ShouldNotBeNil)
			So(typeDesc.InstanceType.Name(), ShouldEqual, "typeToExtract")
			So(typeDesc.IsSlice, ShouldBeTrue)
			So(typeDesc.IsSliceOfPointers, ShouldBeTrue)
			So(typeDesc.StructMapping.Name, ShouldEndWith, "typeToExtract")
		})
	})
}

func TestFillTarget(t *testing.T) {
	Convey("Given a single instance descriptor ", t, func() {
		instancePtr := &typeToExtract{}
		typeDesc, _ := extractType(instancePtr)

		Convey("fillTarget call the given func with the instance pointer", func() {
			typeDesc.fillTarget(func(target interface{}) error {
				So(target, ShouldEqual, instancePtr)
				target.(*typeToExtract).Id = 123
				return nil
			})
			So(instancePtr.Id, ShouldEqual, 123)
		})
	})

	Convey("Given a slice descriptor ", t, func() {
		slice := make([]typeToExtract, 0, 0)
		typeDesc, _ := extractType(&slice)

		Convey("fillTarget call the given func with a new instance pointer", func() {
			typeDesc.fillTarget(func(target interface{}) error {
				So(target, ShouldHaveSameTypeAs, &typeToExtract{})
				target.(*typeToExtract).Id = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, typeToExtract{})
			So(slice[0].Id, ShouldEqual, 123)
		})
	})

	Convey("Given a slice of pointers descriptor ", t, func() {
		slice := make([]*typeToExtract, 0, 0)
		typeDesc, _ := extractType(&slice)

		Convey("fillTarget call the given func with a new instance pointer", func() {
			typeDesc.fillTarget(func(target interface{}) error {
				So(target, ShouldHaveSameTypeAs, &typeToExtract{})
				target.(*typeToExtract).Id = 123
				return nil
			})
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldHaveSameTypeAs, &typeToExtract{})
			So((*slice[0]).Id, ShouldEqual, 123)
		})
	})
}
