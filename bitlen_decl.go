// +build !gccgo
// +build amd64 386 arm

package encoding

import (
	"math/big"
)

// This is defined in util_{amd64,386}.s, copied from pkg/math/big/arith_{amd64/386}.s
func bitlen(x big.Word) (n int)
