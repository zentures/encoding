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
	"time"
	"bufio"
	"strconv"
	"runtime/pprof"
	"compress/gzip"
	"github.com/reducedb/encoding"
	"github.com/reducedb/encoding/composition"
	"github.com/reducedb/encoding/bp32"
	"github.com/reducedb/encoding/variablebyte"
)

func TestTimestampEncoding(t *testing.T) {
	data, err := readTimestamps("ts.txt.gz")
	if err == nil {
		log.Printf("encoding/TestTimeStampEncoding: Read %d integers (%d bytes) from ts.txt.gz.", len(data), len(data)*4)
	} else {
		log.Printf("encoding/TestTimeStampEncoding: Error reading ts.txt.gz: %v\n", err)
	}

	f, e := os.Create("cpu.compress.prof")
	if e != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.Printf("encoding/TestTimeStampEncoding: Testing comprssion BP32+VariableByte\n")
	codec := composition.NewIntegratedComposition(bp32.NewIntegratedBP32(), variablebyte.NewIntegratedVariableByte())
	now := time.Now()
	pprof.StartCPUProfile(f)
	compressed := encoding.Compress(codec, data, len(data))
	pprof.StopCPUProfile()

	log.Printf("encoding/TestTimeStampEncoding: Compressed %d integers from %d bytes to %d bytes in %d ns\n", len(data), len(data)*4, len(compressed)*4, time.Since(now).Nanoseconds())

	f2, e2 := os.Create("cpu.uncompress.prof")
	if e2 != nil {
		log.Fatal(e2)
	}
	defer f2.Close()

	log.Printf("encoding/TestTimeStampEncoding: Testing decompression\n")
	now = time.Now()
	pprof.StartCPUProfile(f2)
	recovered := encoding.Uncompress(codec, compressed, len(data))
	pprof.StopCPUProfile()
	log.Printf("encoding/TestTimeStampEncoding: Uncompressed from %d bytes to %d bytes in %d ns\n", len(compressed)*4, len(recovered)*4, time.Since(now).Nanoseconds())

	for i := 0; i < len(data); i++ {
		if data[i] != recovered[i] {
			t.Fatalf("encoding/TestTimeStampEncoding: Problem recovering. Original length = %d, recovered length = %d\n", len(data), len(recovered))
		}
	}
}

func readTimestamps(path string) ([]uint32, error) {
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

	return result, scanner.Err()
}
