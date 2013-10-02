/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package variablebyte

import (
	"testing"
	"github.com/reducedb/encoding"
)

func TestIntegratedVariableByte(t *testing.T) {
	sizes := []int{100, 100*10, 100*100, 100*1000, 100*10000}
	encoding.TestCodec(NewIntegratedVariableByte(), data, sizes, t)
}

func BenchmarkIntegratedVariableByteCompress(b *testing.B) {
	encoding.BenchmarkCompress(NewIntegratedVariableByte(), data, b)
}

func BenchmarkIntegratedVariableByteUncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewIntegratedVariableByte(), data, b)
}
