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

func TestDeltaVariableByte(t *testing.T) {
	sizes := []int{100, 100*10, 100*100, 100*1000, 100*10000}
	encoding.TestCodec(NewDeltaVariableByte(), data, sizes, t)
}

func BenchmarkDeltaVariableByteCompress(b *testing.B) {
	encoding.BenchmarkCompress(NewDeltaVariableByte(), data, b)
}

func BenchmarkDeltaVariableByteUncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewDeltaVariableByte(), data, b)
}
