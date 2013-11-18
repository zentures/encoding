// +build !gccgo,!amd64,!386,!arm

// (gccgo) OR ((NOT amd64) AND (NOT 386) AND (NOT ARM))
package encoding

import (
	"math/big"
)

func bitlen(x big.Word) (n int) {
	return 32 - int(nlz1a(uint32(x)))
}
