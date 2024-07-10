/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2023-4 by  the Jacobin authors. Consult jacobin.org.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0) All rights reserved.
 */

package gfunction

import (
	"bytes"
	"fmt"
	"jacobin/classloader"
	"jacobin/excNames"
	"jacobin/object"
	"jacobin/types"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// We don't run String's static initializer block because the initialization
// is already handled in String creation

func Load_Lang_String() {

	// === OBJECT INSTANTIATION ===

	// String instantiation without parameters i.e. String string = new String();
	// need to replace eventually by enabling the Java initializer to run
	MethodSignatures["java/lang/String.<clinit>()V"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  stringClinit,
		}

	// String(byte[] bytes) - instantiate a String from a byte array
	MethodSignatures["java/lang/String.<init>()V"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  newEmptyString,
		}

	// String(byte[] bytes) - instantiate a String from a byte array
	MethodSignatures["java/lang/String.<init>([B)V"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  newStringFromBytes,
		}

	// String(byte[] ascii, int hibyte) *** DEPRECATED
	MethodSignatures["java/lang/String.<init>([BI)V"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  trapDeprecated,
		}

	// String(byte[] bytes, int offset, int length)	- instantiate a String from a byte array SUBSET
	MethodSignatures["java/lang/String.<init>([BII)V"] =
		GMeth{
			ParamSlots: 3,
			GFunction:  newStringFromBytesSubset,
		}

	// String(byte[] ascii, int hibyte, int offset, int count) *** DEPRECATED
	MethodSignatures["java/lang/String.<init>([BIII)V"] =
		GMeth{
			ParamSlots: 4,
			GFunction:  trapDeprecated,
		}

	// String(byte[] bytes, int offset, int length, String charsetName) *********** CHARSET
	MethodSignatures["java/lang/String.<init>([BIILjava/lang/String;)V"] =
		GMeth{
			ParamSlots: 4,
			GFunction:  trapFunction,
		}

	// String(byte[] bytes, int offset, int length, Charset charset) ************** CHARSET
	MethodSignatures["java/lang/String.<init>([BIILjava/nio/charset/Charset;)V"] =
		GMeth{
			ParamSlots: 4,
			GFunction:  trapFunction,
		}

	// String(byte[] bytes, String charsetName) *********************************** CHARSET
	MethodSignatures["java/lang/String.<init>([BLjava/lang/String;)V"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  trapFunction,
		}

	// String(byte[] bytes, Charset charset) ************************************** CHARSET
	MethodSignatures["java/lang/String.<init>([BLjava/nio/charset/Charset;)V"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  trapFunction,
		}

	// String(byte[] bytes, Charset charset) ************************************** CHARSET
	MethodSignatures["java/lang/String.<init>([C)V"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  newStringFromChars,
		}

	// String(char[] value) ****************************** works fine with JDK libraries
	// String(char[] value, int offset, int count) ******* works fine with JDK libraries

	// String(int[] codePoints, int offset, int count) ************************ CODEPOINTS
	MethodSignatures["java/lang/String.<init>([III)V"] =
		GMeth{
			ParamSlots: 3,
			GFunction:  trapFunction,
		}

	// String(String original) - works fine in Java

	// String(StringBuffer buffer) ********************************************* StringBuffer
	MethodSignatures["java/lang/String.<init>(Ljava/lang/StringBuffer;)V"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  trapFunction,
		}

	// String(StringBuilder builder) ******************************************* StringBuilder
	MethodSignatures["java/lang/String.<init>(Ljava/lang/StringBuilder;)V"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  trapFunction,
		}

	// ==== METHOD FUNCTIONS (in alpha order by their Java names) ====

	MethodSignatures["java/lang/String.charAt(I)C"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringCharAt,
		}

	// Compare 2 strings lexicographically, case-sensitive (upper/lower).
	// The return value is a negative integer, zero, or a positive integer
	// as the String argument is greater than, equal to, or less than this String,
	// case-sensitive.
	MethodSignatures["java/lang/String.compareTo(Ljava/lang/String;)I"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  compareToCaseSensitive,
		}

	// Compare 2 strings lexicographically, ignoring case (upper/lower).
	// The return value is a negative integer, zero, or a positive integer
	// as the String argument is greater than, equal to, or less than this String,
	// ignoring case considerations.
	MethodSignatures["java/lang/String.compareToIgnoreCase(Ljava/lang/String;)I"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  compareToIgnoreCase,
		}

	MethodSignatures["java/lang/String.concat(Ljava/lang/String;)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringConcat,
		}
	MethodSignatures["java/lang/String.contains(Ljava/lang/CharSequence;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringContains,
		}

	MethodSignatures["java/lang/String.contentEquals(Ljava/lang/CharSequence;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  javaLangStringContentEqualsL,
		}

	MethodSignatures["java/lang/String.contentEquals(Ljava/lang/StringBuffer;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  javaLangStringContentEqualsL,
		}

	MethodSignatures["java/lang/String.equals(Ljava/lang/Object;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringEquals,
		}

	MethodSignatures["java/lang/String.equalsIgnoreCase(Ljava/lang/String;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringEqualsIgnoreCase,
		}

	// Return a formatted string using the reference object string as the format string
	// and the supplied arguments as input object arguments.
	// E.g. String string = String.format("%s %i", "ABC", 42);
	MethodSignatures["java/lang/String.format(Ljava/lang/String;[Ljava/lang/Object;)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  sprintf,
		}

	// This method is equivalent to String.format(this, args).
	MethodSignatures["java/lang/String.formatted([Ljava/lang/Object;)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  sprintf,
		}

	// Return a formatted string using the specified locale, format string, and arguments.
	MethodSignatures["java/lang/String.format(Ljava/util/Locale;Ljava/lang/String;[Ljava/lang/Object;)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 3,
			GFunction:  trapFunction,
		}

	// get the bytes from a string
	MethodSignatures["java/lang/String.getBytes()[B"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  getBytesFromString,
		}

	// void getBytes(int srcBegin, int srcEnd, byte[] dst, int dstBegin)  ********************* DEPRECATED
	MethodSignatures["java/lang/String.getBytes(II[BI)V"] =
		GMeth{
			ParamSlots: 4,
			GFunction:  trapDeprecated,
		}

	// getBytes([BIIBI)V
	// original Java source: https://gist.github.com/platypusguy/03c1a9e3acb1cb2cfc2d821aa2dd4490
	MethodSignatures["java/lang/String.getBytes([BIIBI)V"] =
		GMeth{
			ParamSlots: 5,
			GFunction:  stringGetBytesBIIBI,
		}

	// get the bytes from a string, given the Charset string name ************************ CHARSET
	MethodSignatures["java/lang/String.getBytes(Ljava/lang/String;)[B"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  trapFunction,
		}

	// get the bytes from a string, given the specified Charset object ******************* CHARSET
	MethodSignatures["java/lang/String.getBytes(Ljava/nio/charset/Charset;)[B"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  trapFunction,
		}

	// Return the length of a String.
	MethodSignatures["java/lang/String.isLatin1()Z"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  stringIsLatin1,
		}

	MethodSignatures["java/lang/String.lastIndexOf(Ljava/lang/String;)I"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  lastIndexOfString,
		}
	// Return the length of a String.
	MethodSignatures["java/lang/String.length()I"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  stringLength,
		}

	MethodSignatures["java/lang/String.matches(Ljava/lang/String;)Z"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringMatches,
		}

	MethodSignatures["java/lang/String.regionMatches(ILjava/lang/String;II)Z"] =
		GMeth{
			ParamSlots: 4,
			GFunction:  stringRegionMatchesILII,
		}

	// Returns a string whose value is the concatenation of this string repeated the specified number of times.
	MethodSignatures["java/lang/String.repeat(I)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringRepeat,
		}

	// Replace a single character by another in the given string.
	MethodSignatures["java/lang/String.replace(CC)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  stringReplaceCC,
		}

	// Split a string into an array of strings.
	MethodSignatures["java/lang/String.split(Ljava/lang/String;)[Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  stringSplit,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.substring(I)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  substringToTheEnd,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.substring(II)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  substringStartEnd,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.toCharArray()[C"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  toCharArray,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.toLowerCase()Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  toLowerCase,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.toUpperCase()Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  toUpperCase,
		}

	// Return a string in all lower case, using the reference object string as input.
	MethodSignatures["java/lang/String.trim()Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  trimString,
		}

	// Return a string representing a boolean value.
	MethodSignatures["java/lang/String.valueOf(Z)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfBoolean,
		}

	// Return a string representing a char value.
	MethodSignatures["java/lang/String.valueOf(C)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfChar,
		}

	// Return a string representing a char array.
	MethodSignatures["java/lang/String.valueOf([C)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfCharArray,
		}

	// Return a string representing a char subarray.
	MethodSignatures["java/lang/String.valueOf([CII)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 3,
			GFunction:  valueOfCharSubarray,
		}

	// Return a string representing a double value.
	MethodSignatures["java/lang/String.valueOf(D)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  valueOfDouble,
		}

	// Return a string representing a float value.
	MethodSignatures["java/lang/String.valueOf(F)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfFloat,
		}

	// Return a string representing an int value.
	MethodSignatures["java/lang/String.valueOf(I)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfInt,
		}

	// Return a string representing an int value.
	MethodSignatures["java/lang/String.valueOf(J)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 2,
			GFunction:  valueOfLong,
		}

	// Return a string representing the value of an Object.
	MethodSignatures["java/lang/String.valueOf(Ljava/lang/Object;)Ljava/lang/String;"] =
		GMeth{
			ParamSlots: 1,
			GFunction:  valueOfObject,
		}

}

// ==== INSTANTIATION AND INITIALIZATION FUNCTIONS ====

// Instantiate a new empty string - "java/lang/String.<init>()V"
func newEmptyString(params []interface{}) interface{} {
	// params[0] = target object for string (updated)
	bytes := make([]byte, 0)
	object.UpdateStringObjectFromBytes(params[0].(*object.Object), bytes)
	return nil
}

// Instantiate a new string object from a Go byte array.
// "java/lang/String.<init>([B)V"
func newStringFromBytes(params []interface{}) interface{} {
	// params[0] = reference string (to be updated with byte array)
	// params[1] = byte array object
	bytes := params[1].(*object.Object).FieldTable["value"].Fvalue.([]byte)
	object.UpdateStringObjectFromBytes(params[0].(*object.Object), bytes)
	return nil
}

// Construct a string object from a subset of a Go byte array.
// "java/lang/String.<init>([BII)V"
func newStringFromBytesSubset(params []interface{}) interface{} {
	// params[0] = reference string (to be updated with byte array)
	// params[1] = byte array object
	// params[2] = start offset
	// params[3] = end offset
	bytes := params[1].(*object.Object).FieldTable["value"].Fvalue.([]byte)

	// Get substring start and end offset
	ssStart := params[2].(int64)
	ssEnd := params[3].(int64)

	// Validate boundaries.
	totalLength := int64(len(bytes))
	if totalLength < 1 || ssStart < 0 || ssEnd < 1 || ssStart > (totalLength-1) || (ssStart+ssEnd) > totalLength {
		errMsg1 := "Either nil input byte array, invalid substring offset, or invalid substring length"
		errMsg2 := fmt.Sprintf("\n\twhole='%s' wholelen=%d, offset=%d, sslen=%d\n\n", string(bytes), totalLength, ssStart, ssEnd)
		return getGErrBlk(excNames.StringIndexOutOfBoundsException, errMsg1+errMsg2)
	}

	// Compute subarray and update params[0].
	bytes = bytes[ssStart : ssStart+ssEnd]
	object.UpdateStringObjectFromBytes(params[0].(*object.Object), bytes)
	return nil

}

// Instantiate a new string object from a Go int64 array (Java char array).
// "java/lang/String.<init>([C)V"
func newStringFromChars(params []interface{}) interface{} {
	// params[0] = reference string (to be updated with byte array)
	// params[1] = byte array object
	ints := params[1].(*object.Object).FieldTable["value"].Fvalue.([]int64)

	var bytes []byte
	for _, ii := range ints {
		bytes = append(bytes, byte(ii&0xFF))
	}
	object.UpdateStringObjectFromBytes(params[0].(*object.Object), bytes)
	return nil
}

// "java/lang/String.<clinit>()V" -- String class initialisation
func stringClinit([]interface{}) interface{} {
	klass := classloader.MethAreaFetch(types.StringClassName)
	if klass == nil {
		errMsg := fmt.Sprintf("Could not find class %s in the MethodArea", types.StringClassName)
		return getGErrBlk(excNames.ClassNotLoadedException, errMsg)
	}
	klass.Data.ClInit = types.ClInitRun // just mark that String.<clinit>() has been run
	return nil
}

// ==== METHODS FOR STRING ACTIVITIES ====

// Get character at the given index.
// "java/lang/String.charAt(I)C"
func stringCharAt(params []interface{}) interface{} {
	// Unpack the reference string and convert it to a rune array.
	ptrObj := params[0].(*object.Object)
	str := object.GoStringFromStringObject(ptrObj)
	runeArray := []rune(str)

	// Get index.
	index := params[1].(int64)

	// Return indexed character.
	runeValue := runeArray[index]
	return int64(runeValue)
}

// "java/lang/String.compareTo(Ljava/lang/String;)I"
func compareToCaseSensitive(params []interface{}) interface{} {
	obj := params[0].(*object.Object)
	str1 := object.GoStringFromStringObject(obj)
	obj = params[1].(*object.Object)
	str2 := object.GoStringFromStringObject(obj)
	if str2 == str1 {
		return int64(0)
	}
	if str1 < str2 {
		return int64(-1)
	}
	return int64(1)
}

// "java/lang/String.compareToIgnoreCase(Ljava/lang/String;)I"
func compareToIgnoreCase(params []interface{}) interface{} {
	obj := params[0].(*object.Object)
	str1 := strings.ToLower(object.GoStringFromStringObject(obj))
	obj = params[1].(*object.Object)
	str2 := strings.ToLower(object.GoStringFromStringObject(obj))
	if str2 == str1 {
		return int64(0)
	}
	if str1 < str2 {
		return int64(-1)
	}
	return int64(1)
}

// "java/lang/String.concat(Ljava/lang/String;)Ljava/lang/String;"
func stringConcat(params []interface{}) interface{} {
	var str1, str2 string

	fld := params[0].(*object.Object).FieldTable["value"]
	str1 = string(fld.Fvalue.([]byte))
	fld = params[1].(*object.Object).FieldTable["value"]
	str2 = string(fld.Fvalue.([]byte))
	str := str1 + str2
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.contains(Ljava/lang/CharSequence;)Z"
// charSequence is an interface, generally implemented via String or array of chars
// Here, we assume one of those two options.
func stringContains(params []interface{}) interface{} {
	// get the search string (the string we're searching for, i.e., "foo" in "seafood")
	searchFor := params[1].(*object.Object)
	var searchString string
	switch searchFor.FieldTable["value"].Fvalue.(type) {
	case []uint8:
		searchString = string(searchFor.FieldTable["value"].Fvalue.([]byte))
	case string:
		searchString = searchFor.FieldTable["value"].Fvalue.(string)
	}
	searchIn := params[0].(*object.Object)

	// now get the target string (the string being searched)
	var targetString string
	switch searchIn.FieldTable["value"].Fvalue.(type) {
	case []uint8:
		targetString = string(searchIn.FieldTable["value"].Fvalue.([]byte))
	case string:
		targetString = searchIn.FieldTable["value"].Fvalue.(string)
	}

	if strings.Contains(targetString, searchString) {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

func javaLangStringContentEqualsL(params []interface{}) interface{} {
	obj := params[0].(*object.Object)
	str1 := string(obj.FieldTable["value"].Fvalue.([]byte))
	obj = params[1].(*object.Object)
	str2 := string(obj.FieldTable["value"].Fvalue.([]byte))

	// Are they equal in value?
	if str1 == str2 {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

// Are 2 strings equal?
// "java/lang/String.equals(Ljava/lang/Object;)Z"
func stringEquals(params []interface{}) interface{} {
	// params[0]: reference string object
	// params[1]: compare-to string Object
	obj := params[0].(*object.Object)
	str1 := object.GoStringFromStringObject(obj)
	obj = params[1].(*object.Object)
	str2 := object.GoStringFromStringObject(obj)

	// Are they equal in value?
	if str1 == str2 {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

// Are 2 strings equal, ignoring case?
// "java/lang/String.equalsIgnoreCase(Ljava/lang/String;)Z"
func stringEqualsIgnoreCase(params []interface{}) interface{} {
	// params[0]: reference string object
	// params[1]: compare-to string Object
	obj := params[0].(*object.Object)
	str1 := object.GoStringFromStringObject(obj)
	obj = params[1].(*object.Object)
	str2 := object.GoStringFromStringObject(obj)

	// Are they equal in value?
	upstr1 := strings.ToUpper(str1)
	upstr2 := strings.ToUpper(str2)
	if upstr1 == upstr2 {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

// "java/lang/String.format(Ljava/lang/String;[Ljava/lang/Object;)Ljava/lang/String;"
// "java/lang/String.formatted([Ljava/lang/Object;)Ljava/lang/String;"
func sprintf(params []interface{}) interface{} {
	// params[0]: format string
	// params[1]: argument slice (array of object pointers)
	return StringFormatter(params)
}

// String formatting given a format string and a slice of arguments.
// Called by sprintf, javaIoConsole.go, and javaIoPrintStream.go.
func StringFormatter(params []interface{}) interface{} {
	// params[0]: format string
	// params[1]: argument slice (array of object pointers)

	// Check the parameter length. It should be 2.
	lenParams := len(params)
	if lenParams < 1 || lenParams > 2 {
		errMsg := fmt.Sprintf("StringFormatter: Invalid parameter count: %d", lenParams)
		return getGErrBlk(excNames.IllegalArgumentException, errMsg)
	}
	if lenParams == 1 { // No parameters beyond the format string
		formatStringObj := params[0].(*object.Object)
		return formatStringObj
	}

	// Check the format string.
	var formatString string
	switch params[0].(type) {
	case *object.Object:
		formatStringObj := params[0].(*object.Object) // the format string is passed as a pointer to a string object
		switch formatStringObj.FieldTable["value"].Ftype {
		case types.ByteArray:
			formatString = object.GoStringFromStringObject(formatStringObj)
		default:
			errMsg := fmt.Sprintf("StringFormatter: In the format string object, expected Ftype=%s but observed: %s",
				types.ByteArray, formatStringObj.FieldTable["value"].Ftype)
			return getGErrBlk(excNames.IllegalArgumentException, errMsg)
		}
	default:
		errMsg := fmt.Sprintf("StringFormatter: Expected a string object for the format string but observed: %T", params[0])
		return getGErrBlk(excNames.IllegalArgumentException, errMsg)
	}

	// Make sure that the argument slice is a reference array.
	valuesOut := []any{}
	fld := params[1].(*object.Object).FieldTable["value"]
	if !strings.HasPrefix(fld.Ftype, types.RefArray) {
		errMsg := fmt.Sprintf("StringFormatter: Expected Ftype=%s for params[1]: fld.Ftype=%s, fld.Fvalue=%v",
			types.RefArray, fld.Ftype, fld.Fvalue)
		return getGErrBlk(excNames.IllegalArgumentException, errMsg)
	}

	// valuesIn = the reference array
	valuesIn := fld.Fvalue.([]*object.Object)

	// Main loop for reference array.
	for ii := 0; ii < len(valuesIn); ii++ {

		// Get the current object's value field.
		fld := valuesIn[ii].FieldTable["value"]

		// If type is string object, process it.
		if fld.Ftype == types.ByteArray {
			str := string(fld.Fvalue.([]byte))
			valuesOut = append(valuesOut, str)
		} else {
			// Not a string object.
			switch fld.Ftype {
			case types.ByteArray:
				str := string(fld.Fvalue.([]byte))
				valuesOut = append(valuesOut, str)
			case types.Byte:
				valuesOut = append(valuesOut, uint8(fld.Fvalue.(int64)))
			case types.Bool:
				var zz bool
				if fld.Fvalue.(int64) == 0 {
					zz = false
				} else {
					zz = true
				}
				valuesOut = append(valuesOut, zz)
			case types.Char:
				cc := fmt.Sprint(fld.Fvalue.(int64))
				valuesOut = append(valuesOut, cc)
			case types.Double:
				valuesOut = append(valuesOut, fld.Fvalue.(float64))
			case types.Float:
				valuesOut = append(valuesOut, fld.Fvalue.(float64))
			case types.Int:
				valuesOut = append(valuesOut, fld.Fvalue.(int64))
			case types.Long:
				valuesOut = append(valuesOut, fld.Fvalue.(int64))
			case types.Short:
				valuesOut = append(valuesOut, fld.Fvalue.(int64))
			default:
				errMsg := fmt.Sprintf("StringFormatter: Invalid parameter %d is of type %s", ii+1, fld.Ftype)
				return getGErrBlk(excNames.IllegalArgumentException, errMsg)
			}
		}
	}

	// Use golang fmt.Sprintf to do the heavy lifting.
	str := fmt.Sprintf(formatString, valuesOut...)

	// Return a pointer to an object.Object that wraps the string byte array.
	return object.StringObjectFromGoString(str)
}

// java/lang/String.getBytes()[B
func getBytesFromString(params []interface{}) interface{} {
	// params[0] = reference string with byte array to be returned
	bytes := object.ByteArrayFromStringObject(params[0].(*object.Object))
	return populator("[B", types.ByteArray, bytes)
}

// java/lang/String.getBytes([BIIBI)V
// JDK17 Java source: https://gist.github.com/platypusguy/03c1a9e3acb1cb2cfc2d821aa2dd4490
func stringGetBytesBIIBI(params []any) any {
	fmt.Fprintln(os.Stderr, "java/lang/String.getBytes([BIIBI)V *****************")
	return nil
}

// java/lang/String.lastIndex(string)
// finds the last instance of the search string in the base string. Returns an
// index to the first character if the string is found, -1 if the string is not found
func lastIndexOfString(params []any) any {
	baseStringObject := params[0].(*object.Object)
	baseString := object.GoStringFromStringObject(baseStringObject)

	searchStringObject := params[1].(*object.Object)
	searchString := object.GoStringFromStringObject(searchStringObject)

	lastIndex := strings.LastIndex(baseString, searchString)
	return int64(lastIndex)
}

// "java/lang/String.isLatin1()Z"
func stringIsLatin1(params []interface{}) interface{} {
	// TODO: Someday, the answer might be false.
	return types.JavaBoolTrue // true
}

// "java/lang/String.length()I"
func stringLength(params []interface{}) interface{} {
	// params[0] = string object whose string length is to be measured
	obj := params[0].(*object.Object)
	bytes := object.ByteArrayFromStringObject(obj)
	return int64(len(bytes))
}

// java/lang/String.matches(Ljava/lang/String;)Z
// is the string in params[0] a match for the regex in params[1]?
func stringMatches(params []any) any {
	if len(params) != 2 {
		errMsg := fmt.Sprintf("stringMatches: Expected a string and a regular expression")
		return getGErrBlk(excNames.IllegalArgumentException, errMsg)
	}
	baseStringObject := params[0].(*object.Object)
	baseString := object.GoStringFromStringObject(baseStringObject)

	regexStringObject := params[1].(*object.Object)
	regexString := object.GoStringFromStringObject(regexStringObject)

	regex, err := regexp.Compile(regexString)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid regular expression: %s", regexString)
		return getGErrBlk(excNames.PatternSyntaxException, errMsg)
	}
	if regex.MatchString(baseString) {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

// do two regions in a string match?
// https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/lang/String.html#regionMatches(boolean,int,java.lang.String,int,int)
// param[0] = the base string, param[1] = offset of region in base string, param[2] pointer to second string,
// param[3] offset in second string, param[4] length of region to comapre.
func stringRegionMatchesILII(params []any) any {
	baseStringObject := params[0].(*object.Object)
	baseByteArray := object.ByteArrayFromStringObject(baseStringObject)
	baseOffset := params[1].(int64)

	compareStringObject := params[2].(*object.Object)
	compareByteArray := object.ByteArrayFromStringObject(compareStringObject)
	compareOffset := params[3].(int64)

	if baseOffset < 0 || compareOffset < 0 { // in the JDK, this is the indicated response, rather than an exception(!)
		return types.JavaBoolFalse
	}

	regionLength := params[4].(int64)
	if baseOffset+regionLength > int64(len(baseByteArray)) || // again, erroneous values simply return false
		compareOffset+regionLength > int64(len(compareByteArray)) {
		return types.JavaBoolFalse
	}

	section1 := baseByteArray[baseOffset : baseOffset+regionLength]
	section2 := compareByteArray[compareOffset : compareOffset+regionLength]
	if bytes.Equal(section1, section2) {
		return types.JavaBoolTrue
	}
	return types.JavaBoolFalse
}

// "java/lang/String.repeat(I)Ljava/lang/String;"
func stringRepeat(params []interface{}) interface{} {
	// params[0] = base string
	// params[1] = int64 repetition factor
	oldStr := object.GoStringFromStringObject(params[0].(*object.Object))
	var newStr string
	count := params[1].(int64)
	for ii := int64(0); ii < count; ii++ {
		newStr = newStr + oldStr
	}

	// Return new string in an object.
	obj := object.StringObjectFromGoString(newStr)
	return obj

}

// "java/lang/String.replace(CC)Ljava/lang/String;"
func stringReplaceCC(params []interface{}) interface{} {
	// params[0] = base string
	// params[1] = character to be replaced
	// params[2] = replacement character
	str := object.GoStringFromStringObject(params[0].(*object.Object))
	oldChar := byte((params[1].(int64)) & 0xFF)
	newChar := byte((params[2].(int64)) & 0xFF)
	newStr := strings.ReplaceAll(str, string(oldChar), string(newChar))

	// Return final string in an object.
	obj := object.StringObjectFromGoString(newStr)
	return obj

}

// "java/lang/String.split(Ljava/lang/String;)[Ljava/lang/String;"
func stringSplit(params []interface{}) interface{} {
	// params[0] = base string
	// params[1] = regular expression in a string
	// TODO: As of 2024-07-10, a string, not a regular expression, is assumed to be in params[1].
	oldStr := object.GoStringFromStringObject(params[0].(*object.Object))
	splitter := object.GoStringFromStringObject(params[1].(*object.Object))
	newStrArray := strings.Split(oldStr, splitter)
	var outObjArray []*object.Object
	for ix := 0; ix < len(newStrArray); ix++ {
		outObjArray = append(outObjArray, object.StringObjectFromGoString(newStrArray[ix]))
	}
	return populator("[Ljava/lang/String;", types.RefArray, outObjArray)

}

// "java/lang/String.substring(I)Ljava/lang/String;"
func substringToTheEnd(params []interface{}) interface{} {
	// params[0] = base string
	// params[1] = start offset
	str := object.GoStringFromStringObject(params[0].(*object.Object))

	// Get substring start offset and compute end offset
	ssStart := params[1].(int64)
	ssEnd := int64(len(str))

	// Validate boundaries.
	totalLength := int64(len(str))
	if totalLength < 1 || ssStart < 0 || ssEnd < 1 || ssStart > (totalLength-1) || ssEnd > totalLength {
		errMsg1 := "Either: nil input byte array, invalid substring offset, or invalid substring length"
		errMsg2 := fmt.Sprintf("\n\twhole='%s' wholelen=%d, offset=%d, sslen=%d\n\n", str, totalLength, ssStart, ssEnd)
		return getGErrBlk(excNames.StringIndexOutOfBoundsException, errMsg1+errMsg2)
	}

	// Compute substring.
	str = str[ssStart:ssEnd]

	// Return new string in an object.
	obj := object.StringObjectFromGoString(str)
	return obj

}

// "java/lang/String.substring(II)Ljava/lang/String;"
func substringStartEnd(params []interface{}) interface{} {
	// params[0] = base string
	// params[1] = start offset
	// params[2] = end offset
	str := object.GoStringFromStringObject(params[0].(*object.Object))

	// Get substring start and end offset
	ssStart := params[1].(int64)
	ssEnd := params[2].(int64)

	// Validate boundaries.
	totalLength := int64(len(str))
	if totalLength < 1 || ssStart < 0 || ssEnd < 1 || ssStart > (totalLength-1) || ssEnd > totalLength {
		errMsg1 := "Either: nil input byte array, invalid substring offset, or invalid substring length"
		errMsg2 := fmt.Sprintf("\n\twhole='%s' wholelen=%d, offset=%d, sslen=%d\n\n", str, totalLength, ssStart, ssEnd)
		return getGErrBlk(excNames.StringIndexOutOfBoundsException, errMsg1+errMsg2)
	}

	// Compute substring.
	str = str[ssStart:ssEnd]

	// Return new string in an object.
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.toCharArray()[C"
func toCharArray(params []interface{}) interface{} {
	// params[0]: input string
	obj := params[0].(*object.Object)
	bytes := obj.FieldTable["value"].Fvalue.([]byte)
	var iArray []int64
	for _, bb := range bytes {
		iArray = append(iArray, int64(bb))
	}
	return populator("[C", types.IntArray, iArray)
}

// "java/lang/String.toLowerCase()Ljava/lang/String;"
func toLowerCase(params []interface{}) interface{} {
	// params[0]: input string
	str := strings.ToLower(object.GoStringFromStringObject(params[0].(*object.Object)))
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.toUpperCase()Ljava/lang/String;"
func toUpperCase(params []interface{}) interface{} {
	// params[0]: input string
	str := strings.ToUpper(object.GoStringFromStringObject(params[0].(*object.Object)))
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.trim()Ljava/lang/String;"
func trimString(params []interface{}) interface{} {
	// params[0]: input string
	str := strings.Trim(object.GoStringFromStringObject(params[0].(*object.Object)), " ")
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(Z)Ljava/lang/String;"
func valueOfBoolean(params []interface{}) interface{} {
	// params[0]: input boolean
	value := params[0].(int64)
	var str string
	if value != 0 {
		str = "true"
	} else {
		str = "false"
	}
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(C)Ljava/lang/String;"
func valueOfChar(params []interface{}) interface{} {
	// params[0]: input char
	value := params[0].(int64)
	str := fmt.Sprintf("%c", value)
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf([C)Ljava/lang/String;"
func valueOfCharArray(params []interface{}) interface{} {
	// params[0]: input char array
	propObj := params[0].(*object.Object)
	intArray := propObj.FieldTable["value"].Fvalue.([]int64)
	var str string
	for _, ch := range intArray {
		str += fmt.Sprintf("%c", ch)
	}
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf([CII)Ljava/lang/String;"
func valueOfCharSubarray(params []interface{}) interface{} {
	// params[0]: input char array
	// params[1]: input offset
	// params[2]: input count
	propObj := params[0].(*object.Object)
	intArray := propObj.FieldTable["value"].Fvalue.([]int64)
	var wholeString string
	for _, ch := range intArray {
		wholeString += fmt.Sprintf("%c", ch)
	}
	// Get substring offset and count
	ssOffset := params[1].(int64)
	ssCount := params[2].(int64)

	// Validate boundaries.
	wholeLength := int64(len(wholeString))
	if wholeLength < 1 || ssOffset < 0 || ssCount < 1 || ssOffset > (wholeLength-1) || (ssOffset+ssCount) > wholeLength {
		errMsg := "Either: nil input byte array, invalid substring offset, or invalid substring length"
		return getGErrBlk(excNames.StringIndexOutOfBoundsException, errMsg)
	}

	// Compute substring.
	str := wholeString[ssOffset : ssOffset+ssCount]

	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(D)Ljava/lang/String;"
func valueOfDouble(params []interface{}) interface{} {
	// params[0]: input double
	value := params[0].(float64)
	str := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.Contains(str, ".") {
		str += ".0"
	}
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(F)Ljava/lang/String;"
func valueOfFloat(params []interface{}) interface{} {
	// params[0]: input float
	value := params[0].(float64)
	str := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.Contains(str, ".") {
		str += ".0"
	}
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(I)Ljava/lang/String;"
func valueOfInt(params []interface{}) interface{} {
	// params[0]: input int
	value := params[0].(int64)
	str := fmt.Sprintf("%d", value)
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(J)Ljava/lang/String;"
func valueOfLong(params []interface{}) interface{} {
	// params[0]: input long
	value := params[0].(int64)
	str := fmt.Sprintf("%d", value)
	obj := object.StringObjectFromGoString(str)
	return obj
}

// "java/lang/String.valueOf(Ljava/lang/Object;)Ljava/lang/String;"
func valueOfObject(params []interface{}) interface{} {
	// params[0]: input Object
	ptrObj := params[0].(*object.Object)
	str := ptrObj.FormatField("")
	obj := object.StringObjectFromGoString(str)
	return obj
}
