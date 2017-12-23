/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bp32

import (
	"log"
	"testing"

	"github.com/dataence/encoding/benchtools"
	"github.com/dataence/encoding/generators"
)

var (
	data []int32
	size int = 12800000
)

func init() {
	log.Printf("bp32/init: generating %d int32s\n", size)
	data = generators.GenerateClustered(size, size*2)
	log.Printf("bp32/init: generated %d integers for test", size)
}

func TestCodec(t *testing.T) {
	sizes := []int{128, 128 * 10, 128 * 100, 128 * 1000, 128 * 10000}
	benchtools.TestCodec(New(), data, sizes)
}
