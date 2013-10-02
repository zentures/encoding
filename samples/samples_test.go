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
	"github.com/reducedb/encoding/bp32"
	"github.com/reducedb/encoding/variablebyte"
)

func TestEncoding(t *testing.T) {
	testEncodingWithFile("ts.txt.gz", t)
	//testEncodingWithFile("srcip.txt.gz", t)
	//testEncodingWithFile("dstip.txt.gz", t)
	testEncodingWithFile("latency.txt.gz", t)
}

func testEncodingWithFile(path string, t *testing.T) {
	data, err := readFileOfIntegers(path)
	if err == nil {
		log.Printf("encoding/TestEncoding: Read %d integers (%d bytes) from %s.\n", len(data), len(data)*4, path)
	} else {
		log.Printf("encoding/TestEncoding: Error opening ts.txt.gz: %v\n", err)
	}

	log.Printf("encoding/TestEncoding: Testing comprssion BP32+VariableByte\n")
	encoding.TestCodec(composition.NewIntegratedComposition(bp32.NewIntegratedBP32(), variablebyte.NewIntegratedVariableByte()), data, []int{len(data)}, t)

	log.Printf("encoding/TestEncoding: Testing comprssion VariableByte\n")
	encoding.TestCodec(variablebyte.NewIntegratedVariableByte(), data, []int{len(data)}, t)

	b := make([]byte, len(data)*4)
	for i := 0; i < len(data); i++ {
		binary.LittleEndian.PutUint32(b[i*4:], data[i])
	}

	encoding.TestGzip(b, t)
	encoding.TestLZW(b, t)
}


func readFileOfIntegers(path string) ([]uint32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gunzip, err := gzip.NewReader(file)
	if err != nil {
		log.Printf("encoding/readTimestamps: Error opening gzip reader: %v\n", err)
		return nil, err
	}

	result := make([]uint32, 0, 50000000)
	scanner := bufio.NewScanner(gunzip)
	for scanner.Scan() {
		i, e := strconv.ParseUint(scanner.Text(), 10, 32)
		if e != nil {
			log.Printf("encoding/readTimestamps: Error reading from %s. %v\n", path, e)
		} else {
			result = append(result, uint32(i))
		}
	}

	// Run the garbage collector to get rid of all the strings that's been allocated
	// during the file read
	runtime.GC()

	return result, scanner.Err()
}
