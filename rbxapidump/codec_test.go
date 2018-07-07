package rbxapidump_test

import (
	"bytes"
	"fmt"
	"github.com/anaminus/go-spew/spew"
	"github.com/robloxapi/rbxapi/rbxapidump"
	"testing"
)

const testData = `Class TestClass
Class TestClassTag [testTag]
Class TestClassSuper : TestSuperclass
Class TestClassSuperTag : TestSuperclass [testTag]
Class TestClassTag2 [testTag0] [testTag1]
Class TestClassSuperTag2 : TestSuperclass [testTag0] [testTag1]
Class TestClassMember
	Property TestType TestClassMember.TestProperty
Class TestClassTagMember [testTag]
	Property TestType TestClassTagMember.TestProperty
Class TestClassSuperMember : TestSuperclass
	Property TestType TestClassSuperMember.TestProperty
Class TestClassSuperTagMember : TestSuperclass [testTag]
	Property TestType TestClassSuperTagMember.TestProperty
Class TestClassTag2Member [testTag0] [testTag1]
	Property TestType TestClassTag2Member.TestProperty
Class TestClassSuperTag2Member : TestSuperclass [testTag0] [testTag1]
	Property TestType TestClassSuperTag2Member.TestProperty
Class TestMembers
	Property TestValueType TestMembers.TestProperty
	Function TestReturnType TestMembers:TestFunction()
	YieldFunction TestReturnType TestMembers:TestYieldFunction()
	Event TestMembers.TestEvent()
	Callback TestReturnType TestMembers.TestCallback()
	Property TestValueType TestMembers.TestPropertyTag [testTag]
	Function TestReturnType TestMembers:TestFunctionTag() [testTag]
	YieldFunction TestReturnType TestMembers:TestYieldFunctionTag() [testTag]
	Event TestMembers.TestEventTag() [testTag]
	Callback TestReturnType TestMembers.TestCallbackTag() [testTag]
	Property TestValueType TestMembers.TestProperty
	Function TestReturnType TestMembers:TestFunctionParam(TestParamType testParamName)
	YieldFunction TestReturnType TestMembers:TestYieldFunctionParam(TestParamType testParamName)
	Event TestMembers.TestEventParam(TestParamType testParamName)
	Callback TestReturnType TestMembers.TestCallbackParam1(TestType0 testName0, TestType1 testName1)
	Function TestReturnType TestMembers:TestFunctionParam1(TestType0 testName0, TestType1 testName1)
	YieldFunction TestReturnType TestMembers:TestYieldFunctionParam1(TestType0 testName0, TestType1 testName1)
	Event TestMembers.TestEventParam1(TestType0 testName0, TestType1 testName1)
	Callback TestReturnType TestMembers.TestCallbackParam1(TestType0 testName0, TestType1 testName1)
Class TestSubTag [testTag: [testSubtag]]
Enum TestEnum
Enum TestEnumTag [testTag]
Enum TestEnumTag2 [testTag0] [testTag1]
Enum TestEnumItem
	EnumItem TestEnumItem.TestItem0 : 0
Enum TestEnumItemTag [testTag]
	EnumItem TestEnumItemTag.TestItem0 : 0
Enum TestEnumItemTag2 [testTag0] [testTag1]
	EnumItem TestEnumItemTag2.TestItem0 : 0
Enum TestEnumItems
	EnumItem TestEnumItems.TestItem0 : 0
	EnumItem TestEnumItems.TestItem1 : 1
	EnumItem TestEnumItems.TestItem2 : 2
	EnumItem TestEnumItems.TestItem8 : 8
	EnumItem TestEnumItems.TestItem7 : 7
	EnumItem TestEnumItems.TestItem6 : 6
	EnumItem TestEnumItems.TestItemN : 3735928559
	EnumItem TestEnumItems.TestItem0Tag : 0 [testTag]
	EnumItem TestEnumItems.TestItemNTag : 3735928559 [testTag]
	EnumItem TestEnumItems.TestItem0Tag2 : 0 [testTag0] [testTag1]
	EnumItem TestEnumItems.TestItemNTag2 : 3735928559 [testTag0] [testTag1]
`

func TestCodec(t *testing.T) {
	root, err := rbxapidump.Decode(bytes.NewReader([]byte(testData)))
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	if n, err := rbxapidump.Encode(&buf, root); err != nil {
		t.Error(n, err)
	}

	spew.Dump(root)

	if buf.String() != testData {
		fmt.Println(buf.String())
		t.Error("encoding does not match source")
	}
}
