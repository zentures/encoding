/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package samples

import (
	"testing"
	"log"
	"os"
	"bufio"
	"encoding/binary"
	"compress/gzip"
	"strconv"
	"runtime"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/composition"
	"github.com/reducedb/encoding/fastpfor"
	"github.com/reducedb/encoding/bp32"
	"github.com/reducedb/encoding/variablebyte"
)

func TestEncodingGenerateClustered(t *testing.T) {
	var size int = 10000000

	log.Printf("bp32/init: generating %d int32s\n", size)
	data := encoding.GenerateClustered(size, size*2)
	testEncodingWithData(data, t)
}

func TestEncodingGenerateUniform(t *testing.T) {
	var size int = 10000000

	log.Printf("bp32/init: generating %d int32s\n", size)
	data := encoding.GenerateUniform(size, size*2)
	testEncodingWithData(data, t)
}

func TestEncodingFiles(t *testing.T) {
	for _, f := range []string{"ts.txt.gz", "scrip.txt.gz", "dstip.txt.gz", "latency.txt.gz"} {
		data, err := readFileOfIntegers(f)
		if err == nil {
			log.Printf("encoding/testEncodingWithFile: Read %d integers (%d bytes) from %s.\n", len(data), len(data)*4, f)
			testEncodingWithData(data, t)
		} else {
			log.Printf("encoding/testEncodingWithFile: Error opening ts.txt.gz: %v\n", err)
		}
	}
}

func TestEncodingPprof(t *testing.T) {
	data, err := readFileOfIntegers("ts.txt.gz")
	if err == nil {
		log.Printf("encoding/testEncodingWithFile: Read %d integers (%d bytes) from ts.txt.gz.\n", len(data), len(data)*4)
	} else {
		log.Printf("encoding/testEncodingWithFile: Error opening ts.txt.gz: %v\n", err)
	}

	encoding.TestCodecPprof(composition.NewComposition(bp32.NewZigZagBP32(), variablebyte.NewDeltaVariableByte()), data, []int{len(data)}, t)
}

func testEncodingWithData(data []int32, t *testing.T) {
	log.Printf("encoding/testEncodingWithData: Testing comprssion FastPFOR+VariableByte\n")
	encoding.TestCodec(composition.NewComposition(fastpfor.New(), variablebyte.NewDeltaVariableByte()), data, []int{len(data)}, t)

	log.Printf("encoding/testEncodingWithData: Testing comprssion ZigZag BP32+VariableByte\n")
	encoding.TestCodec(composition.NewComposition(bp32.NewZigZagBP32(), variablebyte.NewDeltaVariableByte()), data, []int{len(data)}, t)

	log.Printf("encoding/testEncodingWithData: Testing comprssion Delta BP32+VariableByte\n")
	encoding.TestCodec(composition.NewComposition(bp32.NewDeltaBP32(), variablebyte.NewDeltaVariableByte()), data, []int{len(data)}, t)

	log.Printf("encoding/testEncodingWithData: Testing comprssion Delta VariableByte\n")
	encoding.TestCodec(variablebyte.NewDeltaVariableByte(), data, []int{len(data)}, t)

	log.Printf("encoding/testEncodingWithData: Testing comprssion BP32+VariableByte\n")
	encoding.TestCodec(composition.NewComposition(bp32.NewBP32(), variablebyte.NewVariableByte()), data, []int{len(data)}, t)

	log.Printf("encoding/testEncodingWithData: Testing comprssion VariableByte\n")
	encoding.TestCodec(variablebyte.NewVariableByte(), data, []int{len(data)}, t)

	b := make([]byte, len(data)*4)
	for i := 0; i < len(data); i++ {
		binary.LittleEndian.PutUint32(b[i*4:], uint32(data[i]))
	}

	encoding.RunTestGzip(b, t)
	encoding.RunTestLZW(b, t)
}

func readFileOfIntegers(path string) ([]int32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gunzip, err := gzip.NewReader(file)
	if err != nil {
		log.Printf("encoding/readFileOfIntegers: Error opening gzip reader: %v\n", err)
		return nil, err
	}

	result := make([]int32, 0, 50000000)
	scanner := bufio.NewScanner(gunzip)
	for scanner.Scan() {
		i, e := strconv.ParseUint(scanner.Text(), 10, 32)
		if e != nil {
			log.Printf("encoding/readFileOfIntegers: Error reading from %s. %v\n", path, e)
		} else {
			result = append(result, int32(i))
		}
	}

	// Run the garbage collector to get rid of all the strings that's been allocated
	// during the file read
	runtime.GC()

	return result, scanner.Err()
}
