/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2024 by  the Jacobin Authors. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)  Consult jacobin.org.
 */

package classloader

// This file contains utility routines for runtime operations involving the
// class's constant pool (CP). It was formerly in the jvm package, but moved
// here to avoid circular dependencies.

import (
	"unsafe"
)

type CpType struct {
	EntryType int
	RetType   int
	IntVal    int64
	FloatVal  float64
	AddrVal   uintptr
	StringVal *string
}

var IS_ERROR = 0
var IS_STRUCT_ADDR = 1
var IS_FLOAT64 = 2
var IS_INT64 = 3
var IS_STRING_ADDR = 4

// Utility routines for runtime operations

// FetchCPentry looks up an entry in a CP and return its type and
// its value. The returned value is a struct that serves as
// a substitute for a discriminated union.
// The fields are:
//  1. EntryType: The CP entry type. Equals 0 if an error occurred.
//     The five EntryType values are listed above: IS_ERROR, etc.
//  2. RetType: int that identifies the type of the returned value.
//     The options are:
//     0 = error
//     1 = address of item other than string
//     2 = float64
//     3 = int64
//     4 = address of string
//  3. three fields that hold an int64, float64, or 64-bit address, respectively.
//     The calling function checks the RetType field to determine which
//     of these three fields holds the returned value.
func FetchCPentry(cpp *CPool, index int) CpType {
	if cpp == nil {
		return CpType{EntryType: 0, RetType: IS_ERROR}
	}
	cp := *cpp
	// if index is out of range, return error
	if index < 1 || index >= len(cp.CpIndex) {
		return CpType{EntryType: 0, RetType: IS_ERROR}
	}

	entry := cp.CpIndex[index]

	switch entry.Type {
	// integers
	case IntConst:
		retInt := int64(cp.IntConsts[entry.Slot])
		return CpType{EntryType: int(entry.Type), RetType: IS_INT64, IntVal: retInt}

	case LongConst:
		retInt := cp.LongConsts[entry.Slot]
		return CpType{EntryType: int(entry.Type), RetType: IS_INT64, IntVal: retInt}

	case MethodType: // method type is an integer
		retInt := int64(cp.MethodTypes[entry.Slot])
		return CpType{EntryType: int(entry.Type), RetType: IS_INT64, IntVal: retInt}

	// floating point
	case FloatConst:
		retFloat := float64(cp.Floats[entry.Slot])
		return CpType{EntryType: int(entry.Type), RetType: IS_FLOAT64, FloatVal: retFloat}

	case DoubleConst:
		retFloat := cp.Doubles[entry.Slot]
		return CpType{EntryType: int(entry.Type), RetType: IS_FLOAT64, FloatVal: retFloat}

	// addresses of strings
	case ClassRef: // points to a CP entry, which is a UTF-8 holding the class name
		e := cp.ClassRefs[entry.Slot]
		className := FetchUTF8stringFromCPEntryNumber(&cp, e)
		return CpType{EntryType: int(entry.Type),
			RetType: IS_STRING_ADDR, StringVal: &className}

	case StringConst: // points to a CP entry, which is a UTF-8 string constant
		e := cp.CpIndex[entry.Slot]
		// should point to a UTF-8
		if e.Type != UTF8 {
			return CpType{EntryType: 0, RetType: IS_ERROR}
		}

		str := cp.Utf8Refs[e.Slot]
		return CpType{EntryType: int(entry.Type),
			RetType: IS_STRING_ADDR, StringVal: &str}
	case UTF8: // same code as for ClassRef
		v := &(cp.Utf8Refs[entry.Slot])
		return CpType{EntryType: int(entry.Type), RetType: IS_STRING_ADDR, StringVal: v}

	// addresses of structures or other elements
	case Dynamic:
		v := unsafe.Pointer(&(cp.Dynamics[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	case Interface:
		v := unsafe.Pointer(&(cp.InterfaceRefs[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	case InvokeDynamic:
		v := unsafe.Pointer(&(cp.InvokeDynamics[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	case MethodHandle:
		v := unsafe.Pointer(&(cp.MethodHandles[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	case MethodRef:
		v := unsafe.Pointer(&(cp.MethodRefs[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	case NameAndType:
		v := unsafe.Pointer(&(cp.NameAndTypes[entry.Slot]))
		return CpType{EntryType: int(entry.Type), RetType: IS_STRUCT_ADDR, AddrVal: uintptr(v)}

	// error: name of module or package would
	// not normally be retrieved here
	case Module,
		Package:
		return CpType{EntryType: 0, RetType: IS_ERROR}
	}

	return CpType{EntryType: 0, RetType: IS_ERROR}
}

func GetMethInfoFromCPmethref(CP *CPool, cpIndex int) (string, string, string) {
	if cpIndex < 1 || cpIndex >= len(CP.CpIndex) {
		return "", "", ""
	}

	if CP.CpIndex[cpIndex].Type != MethodRef {
		return "", "", ""
	}
	methodRef := CP.CpIndex[cpIndex].Slot
	classIndex := CP.MethodRefs[methodRef].ClassIndex
	// nameAndTypeIndex := CP.MethodRefs[methodRef].NameAndType

	classRefIdx := CP.CpIndex[classIndex].Slot
	classIdx := CP.ClassRefs[classRefIdx]
	classNameIdx := CP.CpIndex[classIdx]
	className := CP.Utf8Refs[classNameIdx.Slot]

	// now get the method signature
	nameAndTypeCPindex := CP.MethodRefs[methodRef].NameAndType
	nameAndTypeIndex := CP.CpIndex[nameAndTypeCPindex].Slot
	nameAndType := CP.NameAndTypes[nameAndTypeIndex]
	methNameCPindex := nameAndType.NameIndex
	methNameUTF8index := CP.CpIndex[methNameCPindex].Slot
	methName := CP.Utf8Refs[methNameUTF8index]

	// and get the method signature/description
	methSigCPindex := nameAndType.DescIndex
	methSigUTF8index := CP.CpIndex[methSigCPindex].Slot
	methSig := CP.Utf8Refs[methSigUTF8index]

	return className, methName, methSig
}

// accepts the index of a CP entry, which should point to a classref
// and resolves it to return a string containing the class name.
// Returns an empty string if an error occurred
func GetClassNameFromCPclassref(CP *CPool, cpIndex uint16) string {
	entry := FetchCPentry(CP, int(cpIndex))
	if entry.RetType != IS_STRING_ADDR {
		return ""
	} else {
		return *entry.StringVal
	}
}
