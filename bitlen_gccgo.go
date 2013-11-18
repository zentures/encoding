// +build gccgo

package encoding

// this is apparetly the old way -> func clz(uint64) uint64 __asm__("__clzdi2")

//extern __clzdi2
func clz(uint64) uint64

func bitlen(x uint64) (n int) {
	if x == 0 {
		return 0
	}
	return 64 - int(clz(x))
}
