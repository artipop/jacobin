/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2021 by Andrew Binstock. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)
 */

package classloader

import (
	"io"
	"jacobin/globals"
	"jacobin/stringPool"
	"jacobin/trace"
	"os"
	"strings"
	"testing"
)

// These are the tests in this file (in order of apppearance):
//
// ---- general CP ----
// size of CP							TestInvalidCPsize
//
// ---- constant pool (CP) entries (in order of the numeric value of CP entry type) ----
// missing initial dummy entry			TestMissingInitialDummyEntry
// invalid index into UTF8 entries		TestInvalidIndexInUTF8Entry
// invalid char in UTF8 entry			TestInvalidStringInUTF8Entry
// IntConsts (valid and invalid)		TestIntConsts
// Floats (valid and invalid)			TestFloatConsts
// LongConst and missing dummy entry	TestMissingDummyEntryAfterLongConst
// Double and missing dummy entry		TestDoubleConst
// StringConst (valid and invalid)		TestStringConsts
// invalid index to FieldRef			TestInvalidFieldRef
// FieldRef with invalid name & type	TestFieldRefWithInvalidNameAndTypeIndex
// MethodRef pointing to name with
//     an invalid character in it		TestMethodRefWithInvalidMethodName
// various errors in Interfaces			TestValidInterfaceRefEntry
// valid MethodHandle					TestValidMethodHandleEntry
// invalid MethodHandle (refKind=4) 	TestMethodHandle4PointsToFieldRef
// valid MethodHandle pting to Interface TestValidMethodHandlePointingToInterface
// valid MethodHandle, w/ inv class name TestMethodHandleIndex8ButInvalidName
// invalid MethodHandle (refKind=9)		TestInvalidMethodHandleRefKind9
// valid MethodType 					TestValidMethodType
// valid and invalid Dynamic entries	TestDynamics
// valid InvokeDynamic					TestValidInvokeDynamic
// invalid InvokeDynamic (i.e. missing)	TestInvalidInvokeDynamic
// valid & invalid module names	    	TestModuleNames
// valid & invalid CP module names		TestCPModuleNames
// valid package name					TestCPPackageNames
// valid padkage name (w/out using CP)  TestPackageName
//
// ---- fields (these are different from FieldRefs above) ----
// invalid field name					TestInvalidFieldNames
// invalid field description syntax		TestInvalidFieldDescription
// valid and invalid method description TestMethodDescription
//
// ---- misc routines ----
// syntax of unqualified names			TestUnqualifiedName
// formatCheckStructure routine			TestStructuralValidation
// validity of loadable items			TestLoadableItem

// Note: generates an error if the klass.cpCount of entries does not match the actual number
func TestInvalidCPsize(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"Exceptions"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"testMethod"})

	klass.cpCount = 4 // the error we're testing. There are only two entries, not 4

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Did not get error for mismatch between CP count field and actual number of CP entries")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "Error in size of constant pool") {
		t.Error("Did not get expected error msg for invalid CP count. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestMissingInitialDummyEntry(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"Exceptions"})

	klass.cpCount = 1 // the error we're testing. There are only two entries, not 4

	err := formatCheckClass(&klass)
	if err == nil {
		t.Error("Did not get error for missing initial dummy entry")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "Missing dummy entry in first slot of constant pool") {
		t.Error("Did not get expected error msg for missing initial CP dummy entry. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestInvalidIndexInUTF8Entry(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 4}) // the error: there are only 2 UTF8 entries (see below)

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"Exceptions"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"testMethod"})

	klass.cpCount = 2

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for incorrect ut8Refs index, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "points to invalid UTF8 entry") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestInvalidStringInUTF8Entry(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})

	invalidUtf8bytes := []byte{'B', 'a', 'd', 0xFA} // the last char is disallowed in UTF8 entries
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{string(invalidUtf8bytes)})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"testMethod"})

	klass.cpCount = 2

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for invalid UTF8 string, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "contains an invalid character") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestIntConsts(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{IntConst, 1}) // error, should point to IntConst[0]

	klass.intConsts = append(klass.intConsts, 42)

	klass.cpCount = 2

	// first test an index to non-existent IntConst entry

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for incorrect IntConst, but got none.")
	}

	// now add rec and test valid index to IntConst entry
	klass.intConsts = append(klass.intConsts, 43)

	err = formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error for valid IntConst")
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	// this is the error message left over from the first test of the invalid entry
	if !strings.Contains(msg, "invalid entry in CP intConsts") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestFloatConsts(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{FloatConst, 1}) // error, should point to FloatConst[0]

	klass.floats = append(klass.floats, 42.0)

	klass.cpCount = 2

	// first test an index to non-existent IntConst entry

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for incorrect FloatConst, but got none.")
	}

	// now add rec and test valid index to IntConst entry
	klass.floats = append(klass.floats, 43.0)

	err = formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error for valid FloatConst")
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	// this is the error message left over from the first test of the invalid entry
	if !strings.Contains(msg, "invalid entry in CP floats") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// tests LongConst and the entry afterwards (which should be a dummy entry)
func TestMissingDummyEntryAfterLongConst(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{LongConst, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0}) // this should be a dummy entry

	klass.longConsts = append(klass.longConsts, int64(123))

	klass.cpCount = 3

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for missing dummy entry after long, but got none.")
	}

	// now correct the CP by inserting a dummy entry and make sure it tests right
	klass.cpIndex[2] = cpEntry{Dummy, 0}
	err = formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error with dummy entry after LongConst.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	// tests the remaining error string from the failed test.
	if !strings.Contains(msg, "Missing dummy entry") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestDoubleConst(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{DoubleConst, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0}) // this should be a dummy entry

	klass.doubles = append(klass.doubles, 3.14159)

	klass.cpCount = 3

	// this test validates the double and the requirement that a dummy entry follow the double
	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for missing dummy entry after double, but got none.")
	}

	// now correct the CP by inserting a dummy entry and make sure it tests right
	klass.cpIndex[2] = cpEntry{Dummy, 0}
	err = formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error with dummy entry after DoubleConst.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	// tests the remaining error string from the failed test.
	if !strings.Contains(msg, "Missing dummy entry") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// StringConsts are just indices into the UTF8 entries. So, we just make
// sure they actually point to an actual entry in utf8Refs
func TestStringConsts(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{StringConst, 1}) // error, should point to UTF8[0]

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{content: "Hello, Dolly!"})

	klass.cpCount = 2

	// first test a StringConst that points to a non-existent UTF8 entry

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for incorrect StringConst, but got none.")
	}

	// now add rec and test valid index to UTF8 entry
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{content: "Oh, hello, Dolly!"})

	err = formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error for valid StringConst")
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	// this is the error message left over from the first test of the invalid entry
	if !strings.Contains(msg, "invalid entry in CP utf8Refs") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestInvalidFieldRef(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{FieldRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0}) // unimportant entry

	klass.fieldRefs = append(klass.fieldRefs, fieldRefEntry{
		classIndex:       1, // this points to a non-existent class ref
		nameAndTypeIndex: 0,
	})

	klass.cpCount = 3

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for invalid class index in FieldRef entry, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "points to an invalid entry in ClassRefs") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestFieldRefWithInvalidNameAndTypeIndex(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{FieldRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})

	klass.fieldRefs = append(klass.fieldRefs, fieldRefEntry{
		classIndex:       2, // this correctly points to the ClassRef entry at klass.cpIndex[2]
		nameAndTypeIndex: 1, // this points to a non-existent class ref, causing the tested error
	})
	klass.classRefs = append(klass.classRefs, 0)

	klass.cpCount = 3

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for invalid nameAndType index in FieldRef entry, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "points to an invalid entry in nameAndType") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// a MethodRef points to a class index and a nameAndType index. The name in
// nameAndType must point to a valid class name. If that class name begins with
// a < then it must be <init>. This test makes sure of this latter part.
func TestMethodRefWithInvalidMethodName(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})

	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex:       2, // this correctly points to the ClassRef entry at klass.cpIndex[2]
		nameAndTypeIndex: 3, // this points to a nameAndType entry that points to an invalid class name
	})

	klass.classRefs = append(klass.classRefs, 3)

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       4, // points to cpIndex[4], which is UTF8 rec w/ invalid name
		descriptorIndex: 0,
	})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"<invalidName>"})

	klass.cpCount = 5

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for invalid method name in MethodRef's nameAndType entry, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "an entry with an invalid method name") {
		t.Error("Did not get expected error msg. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// this test validates both InterfaceRefs and NameAndType refs.
func TestValidInterfaceRefEntry(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{Interface, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})

	ire := interfaceRefEntry{classIndex: 2, nameAndTypeIndex: 3}
	klass.interfaceRefs = append(klass.interfaceRefs, ire)

	name1 := "interface"
	name1index := stringPool.GetStringIndex(&name1)

	// klass.classRefs = append(klass.classRefs, 4)
	klass.classRefs = append(klass.classRefs, name1index)

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       4, // points to cpIndex[4], which is UTF8
		descriptorIndex: 5,
	})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"interface"}) // not used -- string pool is used instead
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"B"})

	klass.cpCount = 6

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got but did not expect error in test of valid InterfaceRef.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if len(msg) != 0 {
		t.Error("Got unexpected output to stderr: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// Make sure that all the intricacies of MethodHandles pass the format check
// when a valid MethodHandle entry is run through it.
func TestValidMethodHandleEntry(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})

	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  5, // this requires that the next field be CP entry for MethodRef
		referenceIndex: 2, // index into CP of MethodRef entry
	})

	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex: 7, // points to classRef entry for class name,
		// which poitns to UTF8 record, here: "classname"
		nameAndTypeIndex: 3,
	})

	name1 := "classname"
	name1index := stringPool.GetStringIndex(&name1)
	klass.classRefs = append(klass.classRefs, name1index)

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"classname"}) // not used -- string pool is used instead
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"nAndType-methname"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"D"})

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       5, // points to UTF8[1], i.e., nAndTYpe-methname
		descriptorIndex: 6, // points to UTF8[2], i.e., "D"
	})

	klass.cpCount = 8

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got but did not expect error in test of valid MethodHandle with.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if len(msg) != 0 {
		t.Error("Got unexpected output to stderr: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// MethodHandles with reference kind 1-4 need to point to a FieldRef
// this test checks that an error is generated when that's not the case
func TestMethodHandle4PointsToFieldRef(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})

	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  4, // this requires that the next field be CP entry for MethodRef
		referenceIndex: 2, // index into CP of MethodRef entry, but it should be FieldRef
	})

	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex: 7, // points to classRef entry for class name,
		// which poitns to UTF8 record, here: "classname"
		nameAndTypeIndex: 3,
	})

	klass.classRefs = append(klass.classRefs, 4)

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"classname"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"nAndType-methname"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"D"})

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       5, // points to UTF8[1], i.e., nAndTYpe-methname
		descriptorIndex: 6, // points to UTF8[2], i.e., "D"
	})

	klass.cpCount = 8

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error in test of invalid MethodHandle but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "which does not point to a FieldRef") {
		t.Error("Got unexpected output to stderr: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// method handles with refKind == 6, can point to an interface if
// Java version for the class >= 52. We test for this both with
// Java versions above and below 52 (that is, Java 8)
func TestValidMethodHandlePointingToInterface(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.javaVersion = 54

	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{Interface, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})

	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  6, // this requires that the next field be CP entry for MethodRef
		referenceIndex: 2, // index into CP of MethodRef entry: points to Interface
	})

	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex: 7, // points to classRef entry for class name,
		// which poitns to UTF8 record, here: "classname"
		nameAndTypeIndex: 3,
	})

	klass.interfaceRefs = append(klass.interfaceRefs, interfaceRefEntry{
		classIndex:       7,
		nameAndTypeIndex: 3,
	})

	name1 := "classname"
	name1index := stringPool.GetStringIndex(&name1)
	klass.classRefs = append(klass.classRefs, name1index)

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"classname"}) // not used --- stringPool used instead
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"nAndType-methname"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"D"})

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       5, // points to UTF8[1], i.e., nAndTYpe-methname
		descriptorIndex: 6, // points to UTF8[2], i.e., "D"
	})

	klass.cpCount = 8

	// testing with klassavaVersion = 54, which should be OK

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got but did not expect error in test of valid MethodHandle with" +
			" refIndex = 6, and Java version = 54, but got one.")
	}

	// now run the same test with klass.javaVersion < 52, which should generate an error
	klass.javaVersion = 50
	err = formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Was expecting error in thest of MethodHandle with refIndex = 6" +
			" pointint to an interface and Java version of 50, but did not get one")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "or in Java version 52 or later") {
		t.Error("Got unexpected output error message: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

// MethodHandles refKind = 8 must have a method name of "<init>"
// Note: this test commented out due to issue JACOBIN-183. Should be reinstated
//       in some form, when that issue is resolved.
//
// func TestMethodHandleIndex8ButInvalidName(t *testing.T) {
// 	globals.InitGlobals("test")
// 	trace.Init()
//
// 	// redirect stderr & stdout to capture results from stderr
// 	normalStderr := os.Stderr
// 	r, w, _ := os.Pipe()
// 	os.Stderr = w
//
// 	normalStdout := os.Stdout
// 	_, wout, _ := os.Pipe()
// 	os.Stdout = wout
//
// 	// variables we'll need.
// 	klass := ParsedClass{}
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})
// 	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})
//
// 	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
// 		referenceKind: 8, // this requires that the method name be <init>,
// 		// but it's "methName"
// 		referenceIndex: 2, // index into CP of MethodRef entry
// 	})
//
// 	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
// 		classIndex: 7, // points to classRef entry for class name,
// 		// which poitns to UTF8 record, here: "methName"
// 		nameAndTypeIndex: 3,
// 	})
//
// 	klass.classRefs = append(klass.classRefs, 4)
//
// 	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"methName"})
// 	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"nAndType-methname"})
// 	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"D"})
//
// 	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
// 		nameIndex:       5, // points to UTF8[1], i.e., nAndTYpe-methname
// 		descriptorIndex: 6, // points to UTF8[2], i.e., "D"
// 	})
//
// 	klass.cpCount = 8
//
// 	err := formatCheckConstantPool(&klass)
// 	if err == nil {
// 		t.Error("Expected error for invalid method name, but didn't get any")
// 	}
//
// 	// restore stderr and stdout to what they were before
// 	_ = w.Close()
// 	out, _ := io.ReadAll(r)
// 	os.Stderr = normalStderr
// 	msg := string(out[:])
//
// 	if !strings.Contains(msg, "should be <init>") {
// 		t.Error("Got unexpected error message: " + msg)
// 	}
//
// 	_ = wout.Close()
// 	os.Stdout = normalStdout
// }

func TestInvalidMethodHandleRefKind9(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})

	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  9, // this requires that the reference index point to an interface
		referenceIndex: 2, // should point to an interface but does not
	})

	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex: 7, // points to classRef entry for class name,
		// which poitns to UTF8 record, here: "methName"
		nameAndTypeIndex: 3,
	})

	klass.classRefs = append(klass.classRefs, 4)

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"methName"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"nAndType-methname"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"D"})

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       5, // points to UTF8[1], i.e., nAndTYpe-methname
		descriptorIndex: 6, // points to UTF8[2], i.e., "D"
	})

	klass.cpCount = 8

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Expected error for ReferenceIndex not pointing to Interface, but got none. ")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "reference kind  of 9 which does not point to an interface") {
		t.Error("Got unexpected error message: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestValidMethodType(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.javaVersion = 54

	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})

	klass.methodTypes = append(klass.methodTypes, 2) // points to first UTF8 rec

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"(IDLjava/lang/Thread;)Ljava/lang/Object;"})

	klass.cpCount = 3

	// testing with valid UTF8 record re method type (which must begin with open paren)

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Got unexpected error validating format check of MethodType.")
	}

	// now run the same test an invalid method type (no opening paren)
	klass.utf8Refs[0] = utf8Entry{"IDLjava/lang/Thread;)Ljava/lang/Object;"}
	err = formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Was expecting error in test of MethodType pointing to a type" +
			" string that did not begin with '('")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "not point to a type that starts with an open parenthesis") {
		t.Error("Got unexpected output error message: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestDynamics(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// first create the Bootstrap Method we'll need to point to
	klass := ParsedClass{}
	klass.javaVersion = 55

	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{LongConst, 0}) // LongConst (which is loadable)
	klass.cpIndex = append(klass.cpIndex, cpEntry{Dummy, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{Dynamic, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})

	klass.cpCount = 11

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"BootstrapMethods"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"java/test"}) // not used -- string pool used instead
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"Z"})

	klass.longConsts = append(klass.longConsts, int64(2200))
	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  5, //
		referenceIndex: 6, // points to MethodRef entry
	})
	klass.bootstraps = append(klass.bootstraps, bootstrapMethod{
		methodRef: 3,
		args:      []int{1},
	})
	klass.bootstrapCount = 1
	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex:       7,
		nameAndTypeIndex: 9,
	})

	name1 := "java/test"
	name1index := stringPool.GetStringIndex(&name1)
	klass.classRefs = append(klass.classRefs, name1index)
	klass.classRefs = append(klass.classRefs, name1index)

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       8,
		descriptorIndex: 10,
	})

	klass.dynamics = append(klass.dynamics, dynamic{
		bootstrapIndex: 0,
		nameAndType:    9,
	})

	// ---- attribute
	bsmAttrib := attr{
		attrName:    2,
		attrSize:    8,
		attrContent: []byte{00, 01, 00, 03, 00, 01, 00, 01}, // these bytes represent:
		// 00, 01, // bootstrap count
		// 00, 03, // CP[3] -> MethodHandle
		// 00, 01, // arg count for this first bootstrap method
		// 00, 01, // CP[1] -> LongConst (which is loadable, so all is good.
	}
	klass.attributes = append(klass.attributes, bsmAttrib)
	klass.attribCount = 1

	// variables we'll need.

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Unexpected error in testing dynamic CP entry")
	}

	// now have nameAndType descriptor point to a method, rather than a field.
	// That should cause an error.

	klass.utf8Refs[2] = utf8Entry{content: "V"} // V stands for a void method
	err2 := formatCheckConstantPool(&klass)
	if err2 == nil {
		t.Error("Expected error for invalid dynamic CP entry, but got none.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])
	_ = wout.Close()
	os.Stdout = normalStdout

	if err != nil {
		println(msg)
	}
}

func TestValidInvokeDynamic(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// first create the Bootstrap Method we'll need to point to
	klass := ParsedClass{}
	klass.javaVersion = 55

	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{LongConst, 0}) // LongConst (which is loadable)
	klass.cpIndex = append(klass.cpIndex, cpEntry{Dummy, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodHandle, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{InvokeDynamic, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{ClassRef, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})
	klass.cpIndex = append(klass.cpIndex, cpEntry{NameAndType, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 2})

	klass.cpCount = 11

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"BootstrapMethods"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"java/test"}) // not used --- string pool instead
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"Z"})

	klass.longConsts = append(klass.longConsts, int64(2200))
	klass.methodHandles = append(klass.methodHandles, methodHandleEntry{
		referenceKind:  5, //
		referenceIndex: 6, // points to MethodRef entry
	})
	klass.bootstraps = append(klass.bootstraps, bootstrapMethod{
		methodRef: 3,
		args:      []int{1},
	})
	klass.bootstrapCount = 1
	klass.methodRefs = append(klass.methodRefs, methodRefEntry{
		classIndex:       7,
		nameAndTypeIndex: 9,
	})

	name1 := "java/test"
	name1index := stringPool.GetStringIndex(&name1)
	klass.classRefs = append(klass.classRefs, name1index)
	klass.classRefs = append(klass.classRefs, name1index)

	klass.nameAndTypes = append(klass.nameAndTypes, nameAndTypeEntry{
		nameIndex:       8,
		descriptorIndex: 10,
	})

	klass.invokeDynamics = append(klass.invokeDynamics, invokeDynamic{
		bootstrapIndex: 0,
		nameAndType:    9,
	})

	// ---- attribute
	bsmAttrib := attr{
		attrName:    2,
		attrSize:    8,
		attrContent: []byte{00, 01, 00, 03, 00, 01, 00, 01}, // these bytes represent:
		// 00, 01, // bootstrap count
		// 00, 03, // CP[3] -> MethodHandle
		// 00, 01, // arg count for this first bootstrap method
		// 00, 01, // CP[1] -> LongConst (which is loadable, so all is good.
	}
	klass.attributes = append(klass.attributes, bsmAttrib)
	klass.attribCount = 1

	// variables we'll need.

	err := formatCheckConstantPool(&klass)
	if err != nil {
		t.Error("Unexpected error in testing InvokeDynamic CP entry")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])
	_ = wout.Close()
	os.Stdout = normalStdout

	if err != nil {
		println(msg)
	}
}

// checking the error message for non-existent InvokeDynamic entry.
// This is simply to proof the fix in JACOBIN-81
func TestInvalidInvokeDynamic(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{InvokeDynamic, 0})

	klass.cpCount = 2

	err := formatCheckConstantPool(&klass)
	if err == nil {
		t.Error("Did not get expected error for missing InvokeDynamic.")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = normalStderr
	msg := string(out[:])

	if !strings.Contains(msg, "points to a non-existent invokeDynamic slot") {
		t.Error("Did not get the expected error message for missing InvokeDynamic. Got: " + msg)
	}

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestModuleNames_Test0(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.moduleName = "@invalid"
	if checkModuleName(klass.moduleName) == nil {
		t.Error("Expecting error on invalid module name (@invalid), but got none.")
	}

	klass.moduleName = "\\@valid"
	if checkModuleName(klass.moduleName) != nil {
		_ = w.Close()
		out, _ := io.ReadAll(r)
		msg := string(out[:])
		t.Error("Unexpected error occurred with valid module name: \\@valid\n" +
			"Error message: " + msg)
	} else {
		_ = w.Close()
	}

	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestModuleNames_Test1(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	if checkModuleName("") == nil {
		t.Error("Expected error on test of empty module name, but got none")
	}

	if checkModuleName("@invalid") == nil {
		t.Error("Expected error in module name starting with @, but got none")
	}

	if checkModuleName("invalid:") == nil {
		t.Error("Expected error in module name with non-escaped :, but got none")
	}

	if checkModuleName("invalid"+"\\") == nil {
		t.Error("Expected error in module name with ending \\:, but got none")
	}

	if checkModuleName("invalid"+"\\n") == nil {
		t.Error("Expected error in module name with ending \\n:, but got none")
	}

	if checkModuleName("valid"+"\\@") != nil {
		t.Error("Got unexpected error in module name with ending \\@")
	}

	if checkModuleName("goodname") != nil {
		t.Error("Expected no error in module name 'goodname', but got one")
	}

	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestCPModuleNames(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	// _, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{Module, 0})

	klass.cpCount = 2
	klass.moduleName = "\\@valid"
	klass.javaVersion = 55
	klass.classIsModule = true

	if formatCheckConstantPool(&klass) != nil {
		_ = w.Close()
		out, _ := io.ReadAll(r)
		msg := string(out[:])
		t.Error("Unexpected error occurred with valid module name: \\@valid\n" +
			"Error message: " + msg)
	} else {
		_ = w.Close()
	}

	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestCPPackageNames(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	r, w, _ := os.Pipe()
	// _, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{Package, 0})

	klass.cpCount = 2
	klass.packageName = "\\@valid"
	klass.javaVersion = 55
	klass.classIsModule = true

	if formatCheckConstantPool(&klass) != nil {
		_ = w.Close()
		out, _ := io.ReadAll(r)
		msg := string(out[:])
		t.Error("Unexpected error occurred with valid package name: \\@valid\n" +
			"Error message: " + msg)
	} else {
		_ = w.Close()
	}

	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestPackageName(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	// _, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	if checkPackageName("") == nil {
		t.Error("Expected error on test of empty package name, but got none")
	}

	if checkPackageName("@invalid") == nil {
		t.Error("Expected error in package name starting with @, but got none")
	}

	if checkPackageName("invalid:") == nil {
		t.Error("Expected error in package name with non-escaped :, but got none")
	}

	if checkPackageName("invalid"+"\\") == nil {
		t.Error("Expected error in package name with ending \\:, but got none")
	}

	if checkPackageName("invalid"+"\\n") == nil {
		t.Error("Expected error in package name with ending \\n:, but got none")
	}

	if checkPackageName("valid"+"\\@") != nil {
		t.Error("Got unexpected error in package name with ending \\@")
	}

	if checkPackageName("goodname") != nil {
		t.Error("Expected no error in package name 'goodname', but got one")
	}

	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

// Tests module name without using CP records. Identical logic to TestPackageName(), except
// error messages are different.
func TestModuleName(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to capture results from stderr
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	// _, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	if checkModuleName("") == nil {
		t.Error("Expected error on test of empty module name, but got none")
	}

	if checkModuleName("@invalid") == nil {
		t.Error("Expected error in module name starting with @, but got none")
	}

	if checkModuleName("invalid:") == nil {
		t.Error("Expected error in module name with non-escaped :, but got none")
	}

	if checkModuleName("invalid"+"\\") == nil {
		t.Error("Expected error in module name with ending \\:, but got none")
	}

	if checkModuleName("goodname") != nil {
		t.Error("Expected no error in module name 'goodnae', but got one")
	}

	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

// field names in Java cannot begin with a digit and they cannot contain
// whitespace. We check for both here.
func TestInvalidFieldNames(t *testing.T) {

	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to avoid noisy output
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"bad name"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"B"})

	klass.cpCount = 3

	klass.fieldCount = 1
	klass.fields = append(klass.fields, field{
		accessFlags: 0,
		name:        0, // points to the first utf8Refs entry
		description: 1, // points to the 2nd utf8Refs entry
		attributes:  nil,
	})

	err := formatCheckFields(&klass)
	if err == nil {
		t.Error("Did not get expected error for field name with embedded space.")
	}

	// now test a field name that begins with a digit
	klass.utf8Refs[0] = utf8Entry{"99bottlesOfBeer"}
	err = formatCheckFields(&klass)
	if err == nil {
		t.Error("Did not get expected error for field name starting with digit")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

// the field description must start with one only a few characters, of which
// 's' (our test value) is not one. We also test for an empty description
func TestInvalidFieldDescription(t *testing.T) {

	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to avoid noisy output
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})

	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"validName"})
	klass.utf8Refs = append(klass.utf8Refs, utf8Entry{"s"})

	klass.cpCount = 3

	klass.fieldCount = 1
	klass.fields = append(klass.fields, field{
		accessFlags: 0,
		name:        0,
		description: 1,
		attributes:  nil,
	})

	err := formatCheckFields(&klass)
	if err == nil {
		t.Error("Did not get expected error for invalid field description for " +
			"field: validName")
	}

	// now test for empty description string
	klass.utf8Refs[1] = utf8Entry{""}
	err = formatCheckFields(&klass)
	if err == nil {
		t.Error("Did not get expected error for empty field description for " +
			"field: validName")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestMethodDescription(t *testing.T) {
	if validateMethodDesc("") == nil {
		t.Error("Did not get expected error for empty method descriptor")
	}

	if validateMethodDesc("V") != nil {
		t.Error("Got unexpected error for valid method descriptor: 'V'")
	}

	if validateMethodDesc("notValid") == nil {
		t.Error("Did not get expected error for invalid method descriptor: notValid")
	}
}

func TestCheckStructure(t *testing.T) {
	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to avoid noisy output
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	pClass := ParsedClass{}

	pClass.cpCount = 1
	pClass.cpIndex = nil
	if formatCheckStructure(&pClass) == nil {
		t.Error("Expecting error in mismatch of cpCpount and cpIndex.len, but got none")
	}
	pClass.cpCount = 0

	pClass.interfaceCount = 2
	pClass.interfaces = nil
	if formatCheckStructure(&pClass) == nil {
		t.Error("Expecting error in mismatch of interfaceCount and interfaces.len, but got none")
	}
	pClass.interfaceCount = 0

	pClass.methodCount = 3
	pClass.methods = nil
	if formatCheckStructure(&pClass) == nil {
		t.Error("Expecting error in mismatch of methodCount and methods.len, but got none")
	}
	pClass.methodCount = 0

	pClass.attribCount = 4
	pClass.attributes = nil
	if formatCheckStructure(&pClass) == nil {
		t.Error("Expecting error in mismatch of attribCount and attributess.len, but got none")
	}
	pClass.attribCount = 0

	pClass.bootstrapCount = 5
	pClass.bootstraps = nil
	if formatCheckStructure(&pClass) == nil {
		t.Error("Expecting error in mismatch of bootstrapCount and bootstraps.len, but got none")
	}
	pClass.bootstrapCount = 0

	// restore stderr and stdout to what they were before
	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

// unqualified names in Java have a set of restrictions on the syntax, which
// varies depending on whether the name is the name of a method.
func TestUnqualifiedName(t *testing.T) {
	isMethod := true
	isNotMethod := false

	if validateUnqualifiedName("", isNotMethod) != false {
		t.Error("Expected 'false' for test of empty unqualified name, but got OK")
	}

	if validateUnqualifiedName("[array]", isNotMethod) != false {
		t.Error("Expected 'false' for test of unqualified name '[array]', but got OK")
	}

	if validateUnqualifiedName("isArray", isNotMethod) == false {
		t.Error("Expected 'true' for test of unqualified name 'isArray', but got false")
	}

	if validateUnqualifiedName("<clinit>", isMethod) == false {
		t.Error("Expected 'true' for test of unqualified method name '<clinit>', but got false")
	}

	if validateUnqualifiedName("java/isOpen", isMethod) != false {
		t.Error("Expected 'false' for test of unqualified method name 'java/isOpen', but got true")
	}

	if validateUnqualifiedName("invalid<>", isMethod) != false {
		t.Error("Expected 'false' for test of unqualified method name 'invalid<>', but got true")
	}
}

func TestStructuralValidation(t *testing.T) {

	globals.InitGlobals("test")
	trace.Init()

	// redirect stderr & stdout to avoid noisy output
	normalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	normalStdout := os.Stdout
	_, wout, _ := os.Pipe()
	os.Stdout = wout

	// variables we'll need.
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 1})

	klass.cpCount = 2 // this is an error, it should be 3. Format check should catch this.

	err := formatCheckStructure(&klass)
	if err == nil {
		t.Error("Did not get expected error for mismatch between CP count and " +
			"total number of CP entries")
	}

	klass.interfaces = append(klass.interfaces, 42)
	klass.interfaces = append(klass.interfaces, 43)
	klass.interfaceCount = 4 // should be 2, so an error should ensue
	if formatCheckStructure(&klass) == nil {
		t.Error("Did not get expected error for mistmatch between interfaceCount and " +
			"total number of interfaces")
	}

	// correct interfaceCount and run similar test on methodCount
	klass.interfaceCount = 2
	klass.methods = append(klass.methods, method{})
	klass.methodCount = 5 // should be 1, so an error should ensue
	if formatCheckStructure(&klass) == nil {
		t.Error("Did not get expected error for mistmatch between methodCount and " +
			"total number of methods")
	}

	// correct methodCount and run similar test on attributeCount
	klass.methodCount = 1
	klass.attributes = append(klass.attributes, attr{})
	klass.attributes = append(klass.attributes, attr{})
	klass.attributes = append(klass.attributes, attr{})
	klass.attribCount = 6 // should be 3, so an error should ensue
	if formatCheckStructure(&klass) == nil {
		t.Error("Did not get expected error for mistmatch between attribCount and " +
			"total number of class attributes")
	}

	// restore stderr and stdout to what they were before
	_ = w.Close()
	os.Stderr = normalStderr

	_ = wout.Close()
	os.Stdout = normalStdout
}

func TestLoadableItem(t *testing.T) {
	klass := ParsedClass{}
	klass.cpIndex = append(klass.cpIndex, cpEntry{})
	klass.cpIndex = append(klass.cpIndex, cpEntry{UTF8, 0})
	klass.cpIndex = append(klass.cpIndex, cpEntry{MethodType, 0})

	if validateItemIsLodable(&klass, 8) {
		t.Error("Invalid value for loadable item did not return an error.")
	}

	if validateItemIsLodable(&klass, 1) {
		t.Error("Index for loadable item -> to UTF8 (which is not loadable) did not return error")
	}

	if !validateItemIsLodable(&klass, 2) {
		t.Error("Valid index for loadable item returned an error")
	}
}
