/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package generators

import (
	"fmt"
	"testing"
)

func TestGenerateClustered(t *testing.T) {
	a := GenerateClustered(20, 1000)
	fmt.Println(a)
}

func TestGenerateUniform(t *testing.T) {
	a := GenerateUniform(20, 1000)
	fmt.Println(a)
}
