/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2024 by  the Jacobin Authors. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)  Consult jacobin.org.
 */

package object

import (
	"jacobin/stringPool"
	"jacobin/types"
	"strings"
	"unicode"
)

func GoStringFromJavaByteArray(jbarr []types.JavaByte) string {
	var sb strings.Builder
	for _, b := range jbarr {
		sb.WriteByte(byte(b))
	}
	return sb.String()
}

func JavaByteArrayFromGoString(str string) []types.JavaByte {
	jbarr := make([]types.JavaByte, len(str))
	for i, b := range str {
		jbarr[i] = types.JavaByte(b)
	}
	return jbarr
}

func JavaByteArrayFromGoByteArray(gbarr []byte) []types.JavaByte {
	jbarr := make([]types.JavaByte, len(gbarr))
	for i, b := range gbarr {
		jbarr[i] = types.JavaByte(b)
	}
	return jbarr
}

func GoByteArrayFromJavaByteArray(jbarr []types.JavaByte) []byte {
	gbarr := make([]byte, len(jbarr))
	for i, b := range jbarr {
		gbarr[i] = byte(b)
	}
	return gbarr
}

// JavaByteFromStringObject: convenience method to extract a Java byte array from a String object (Java string)
func JavaByteArrayFromStringObject(obj *Object) []types.JavaByte {
	if obj != nil && obj.KlassName == types.StringPoolStringIndex {
		return obj.FieldTable["value"].Fvalue.([]types.JavaByte)
	} else {
		return nil
	}
}

// StringObjectFromJavaByteArray: convenience method to create a string object from a JavaByte array
func StringObjectFromJavaByteArray(bytes []types.JavaByte) *Object {
	newStr := NewStringObject()
	newStr.FieldTable["value"] = Field{Ftype: types.ByteArray, Fvalue: bytes}
	return newStr
}

// JavaByteArrayFromStringPoolIndex: convenience method to get a byte array using a string pool index
func JavaByteArrayFromStringPoolIndex(index uint32) []types.JavaByte {
	if index < stringPool.GetStringPoolSize() {
		str := *stringPool.GetStringPointer(index)
		return JavaByteArrayFromGoString(str)
	} else {
		return nil
	}
}

func JavaByteArrayEquals(jbarr1, jbarr2 []types.JavaByte) bool {
	if jbarr1 == nil || jbarr2 == nil {
		if jbarr1 == nil && jbarr2 == nil {
			return true
		}
		return false
	}

	if len(jbarr1) != len(jbarr2) {
		return false
	}
	for i, b := range jbarr1 {
		if b != jbarr2[i] {
			return false
		}
	}
	return true
}

func JavaByteArrayEqualsIgnoreCase(jbarr1, jbarr2 []types.JavaByte) bool {
	if jbarr1 == nil || jbarr2 == nil {
		if jbarr1 == nil && jbarr2 == nil {
			return true
		}
		return false
	}

	if len(jbarr1) != len(jbarr2) {
		return false
	}
	for i, b := range jbarr1 {
		if unicode.ToLower(rune(b)) != unicode.ToLower(rune(jbarr2[i])) {
			return false
		}
	}
	return true
}
