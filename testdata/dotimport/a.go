package dotimport

import (
	. "flag" // want `package flag should not be dot-imported`
)

var x = String("x", "y", "z")
