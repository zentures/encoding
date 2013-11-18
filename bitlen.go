// +build !gccgo,!amd64,!386,!arm

// (gccgo) OR ((NOT amd64) AND (NOT 386) AND (NOT ARM))
package encoding

func bitlen(x uint64) (n int) {
	return 32 - int(nlz1a(uint32(x)))
}
