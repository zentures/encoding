/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"sort"
	"errors"
	"math/rand"
	"github.com/willf/bitset"
)

const (
	c1 int64 = 0xcc9e2d51
	c2 int64 = 0x1b873593
)

func GenerateUniform(N, max int) []uint32 {
	if N*2 > max {
		return negate(GenerateUniform(max - N, max), max)
	}

	if 2048*N > max {
		r, _ := generateUniformBitmap(N, max)
		return r

	}

	r, _ := generateUniformHash(N, max)
	return r
}

func GenerateClustered(N, max int) []uint32 {
	ans := make([]uint32, N)
	fillClustered(ans, 0, N, 0, max)
	return ans
}

func fillUniform(ans[]uint32, offset, length, min, max int) {
	v := GenerateUniform(length, max - min)
	for k := 0; k < len(v); k++ {
		ans[k + offset] = uint32(min) + v[k]
	}
}

func fillClustered(ans[]uint32, offset, length, min, max int) {
	btwn := max - min
	if btwn == length || length <= 10 {
		fillUniform(ans, offset, length, min, max)
		return
	}

	r := rand.New(rand.NewSource(c1))
	cut := length/2
	if btwn - length - 1 > 0 {
		cut += int(r.Int31n(int32(btwn - length - 1)))
	}

	p := r.Float64()
	if p < 0.25 {
		fillUniform(ans, offset, length/2, min, min + cut)
		fillClustered(ans, offset + length/2, length - length/2, min + cut, max)
	} else if p < 0.5 {
		fillClustered(ans, offset, length/2, min, min + cut)
		fillUniform(ans, offset + length/2, length - length/2, min + cut, max)
	} else {
		fillClustered(ans, offset, length/2, min, min + cut)
		fillClustered(ans, offset + length/2, length - length/2, min + cut, max)
	}
}

func negate(x []uint32, max int) []uint32 {
	ans := make([]uint32, max - len(x))

	var i, c uint32

	for j := 0; j < len(x); j++ {
		v := x[j]
		for ; i < v; i++ {
			ans[c] = i;
			c += 1
		}
		i += 1
	}

	for int(c) < len(ans) {
		ans[c] = i
		c += 1
		i += 1
	}

	return ans
}

func generateUniformBitmap(N, max int) ([]uint32, error) {
	if N > max {
		return nil, errors.New("encoding/generateUniformBitmap: N > max, not possible")
	}

	r := rand.New(rand.NewSource(c1))
	ans := make([]uint32, N)
	bs := bitset.New(uint(max))
	cardinality := uint(0)

	for int(cardinality) < N {
		v := r.Int31n(int32(max))
		if !bs.Test(uint(v)) {
			bs.Set(uint(v))
			cardinality += 1
		}
	}

	for i, c := uint32(0), 0; c < N; i++ {
		if bs.Test(uint(i)) {
			ans[c] = i
			c += 1
		}
	}

	return ans, nil
}

func generateUniformHash(N, max int) ([]uint32, error) {
	if N > max {
		return nil, errors.New("encoding/generateUniformBitmap: N > max, not possible")
	}

	r := rand.New(rand.NewSource(c2))
	ans := make([]uint32, N)
	s := make(map[int]bool)

	for len(s) < N {
		s[int(r.Int31n(int32(max)))] = true
	}

	c := 0
	tmpans := make([]int, N)
	for k, _ := range s {
		tmpans[c] = k
	}

	sort.Ints(tmpans)

	for i := 0; i < len(tmpans); i++ {
		ans[i] = uint32(tmpans[i])
	}

	return ans, nil
}
