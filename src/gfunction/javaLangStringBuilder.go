/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2023 by  the Jacobin authors. Consult jacobin.org.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0) All rights reserved.
 */

package gfunction

import (
	"jacobin/types"
)

// Implementation of some of the functions in Java/lang/Class.

func Load_Lang_StringBuilder() {

	MethodSignatures["java/lang/StringBuilder.isLatin1()Z"] =
		GMeth{
			ParamSlots: 0,
			GFunction:  isLatin1,
		}

}

// "java/lang/StringBuilder.isLatin1()Z"
func isLatin1([]interface{}) interface{} {
	// TODO: Someday, jacobin will need to discern between StringLatin1 and StringUTF16.
	return types.JavaBoolTrue
}
