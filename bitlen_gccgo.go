// +build gccgo

package encoding

import (
    "math/big"
)

func clz(uint64) uint64 __asm__("__clzdi2")

func bitlen(x big.Word) (n int) {
    if x == 0 { return 0 }
    return 64-int(clz(uint64(x)))
}
