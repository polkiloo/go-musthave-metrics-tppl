package resettest

import (
	"fmt"

	"example.com/resettest/extpkg"
)

type Resettable interface{ Reset() }

type CustomReset struct{ V int }

func (c *CustomReset) Reset() { c.V = 0 }

type CustomResetValue struct{ V int }

func (c CustomResetValue) Reset() {}

type Other struct{ V string }

// generate:reset
type Sample struct {
	I                 int
	Str               string
	Flag              bool
	Slice             []int
	Map               map[string]string
	PtrToSlice        *[]int
	PtrToMap          *map[string]string
	PtrToStruct       *Other
	PtrWithReset      *CustomReset
	StructWithReset   CustomResetValue
	StructNoReset     Other
	InterfaceReset    Resettable
	InterfaceNonReset fmt.Stringer
	External          extpkg.External
}

// generate:reset
type Another struct{ Count int }

// generate:reset
type (
	DeclAnnotated struct{ Value int }
)
