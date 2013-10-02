/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"testing"
	"github.com/reducedb/encoding"
)

func TestIntegratedBP32(t *testing.T) {
	sizes := []int{128, 128*10, 128*100, 128*1000, 128*10000}
	encoding.TestCodec(NewIntegratedBP32(), data, sizes, t)
}

func BenchmarkIntegratedBP32Compress(b *testing.B) {
	encoding.BenchmarkCompress(NewIntegratedBP32(), data, b)
}

func BenchmarkIntegratedBP32Uncompress(b *testing.B) {
	encoding.BenchmarkUncompress(NewIntegratedBP32(), data, b)
}
