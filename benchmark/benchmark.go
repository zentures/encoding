/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/dataence/encoding"
	"github.com/dataence/encoding/bp32"
	"github.com/dataence/encoding/composition"
	"github.com/dataence/encoding/cursor"
	dbp32 "github.com/dataence/encoding/delta/bp32"
	dfastpfor "github.com/dataence/encoding/delta/fastpfor"
	dvb "github.com/dataence/encoding/delta/variablebyte"
	"github.com/dataence/encoding/fastpfor"
	"github.com/dataence/encoding/variablebyte"
	zbp32 "github.com/dataence/encoding/zigzag/bp32"
	zfastpfor "github.com/dataence/encoding/zigzag/fastpfor"
)

type paramList []string

var (
	filesParam, dirsParam, codecsParam paramList
	pprofParam                         bool
	files                              []string
)

func (this *paramList) String() string {
	return fmt.Sprint(*this)
}

func (this *paramList) Set(value string) error {
	for _, f := range strings.Split(value, ",") {
		*this = append(*this, f)
	}

	return nil
}

func init() {
	flag.BoolVar(&pprofParam, "pprof", false, "Print result for individual files.")
	flag.Var(&filesParam, "file", "The file containing one integer per line to encode. There can be multiple of this, or comma separated list.")
	flag.Var(&dirsParam, "dir", "The directory containing a list of files with one integer per line. There can be multiple of this, or comma separated list.")
	flag.Var(&codecsParam, "codec", "The codec to use: bp32, fastpfor, variablebyte, deltabp32, deltafastpfor, deltavariablebyte, zigzagbp32. There can be multiple of this, or comma separated list.")
}

func scanIntegers(s *bufio.Scanner) ([]int32, error) {
	result := make([]int32, 0, 1000000)
	for s.Scan() {
		i, err := strconv.ParseUint(s.Text(), 10, 32)
		if err != nil {
			return nil, err
		} else {
			result = append(result, int32(i))
		}
	}

	// Run the garbage collector to get rid of all the strings that's been allocated
	// during the file read
	runtime.GC()

	return result, nil

}

func readIntegerFile(path string) ([]int32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	return scanIntegers(scanner)
}

func readGzippedIntegerFile(path string) ([]int32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gunzip, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(gunzip)

	return scanIntegers(scanner)
}

func getDirOfFiles(path string) ([]string, error) {
	filenames := make([]string, 0, 10)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		filenames = append(filenames, path+"/"+f.Name())
	}

	return filenames, nil
}

func loadIntegerFromFiles(files []string) ([][]int32, int, error) {
	max := 0
	data := make([][]int32, 0, len(files))

	for _, f := range files {
		var (
			res []int32
			err error
		)

		log.Printf("Processing %s\n", f)

		if strings.HasPrefix(f, "gz-") {
			res, err = readGzippedIntegerFile(strings.TrimPrefix(f, "gz-"))
		} else if strings.HasSuffix(f, ".gz") {
			res, err = readGzippedIntegerFile(f)
		} else {
			res, err = readIntegerFile(f)
		}

		if err != nil {
			return nil, 0, err
		}

		data = append(data, res)

		if len(res) > max {
			max = len(res)
		}
	}

	return data, max, nil
}

func getListOfFiles() []string {
	files := make([]string, 0, 10)

	for _, d := range dirsParam {
		res, err := getDirOfFiles(d)
		if err != nil {
			log.Fatal(err)
		}

		files = append(files, res...)
	}

	files = append(files, filesParam...)

	return files
}

func getListOfCodecs() (map[string]encoding.Integer, error) {
	codecs := make(map[string]encoding.Integer, 10)

	for _, codec := range codecsParam {
		switch codec {
		case "bp32":
			codecs["bp32"] = composition.New(bp32.New(), variablebyte.New())
		case "fastpfor":
			codecs["fastpfor"] = composition.New(fastpfor.New(), variablebyte.New())
		case "variablebyte":
			codecs["variablebyte"] = variablebyte.New()
		case "deltabp32":
			codecs["delta bp32"] = composition.New(dbp32.New(), dvb.New())
		case "deltafastpfor":
			codecs["delta fastpfor"] = composition.New(dfastpfor.New(), dvb.New())
		case "deltavariablebyte":
			codecs["delta variablebyte"] = dvb.New()
		case "zigzagbp32":
			codecs["zigzag bp32"] = composition.New(zbp32.New(), dvb.New())
		case "zigzagfastpfor":
			codecs["zigzag fastpfor"] = composition.New(zfastpfor.New(), dvb.New())
		}
	}

	if len(codecs) < 1 {
		return nil, fmt.Errorf("benchmark/getListOfCodecs: No codecs defined")
	}

	return codecs, nil
}

func compress(codec encoding.Integer, in, out []int32, length int, prof bool) (duration int64, ret []int32, err error) {
	inpos := cursor.New()
	outpos := cursor.New()

	now := time.Now()
	if prof {
		f, e := os.Create("cpu.compress.pprof")
		if e != nil {
			log.Fatal(e)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
	}

	if err = codec.Compress(in, inpos, len(in), out, outpos); err != nil {
		return 0, nil, err
	}
	since := time.Since(now).Nanoseconds()

	if prof {
		pprof.StopCPUProfile()
	}

	return since, out[:outpos.Get()], nil
}

func uncompress(codec encoding.Integer, in, out []int32, length int, prof bool) (duration int64, ret []int32, err error) {
	inpos := cursor.New()
	outpos := cursor.New()

	if prof {
		f, e := os.Create("cpu.uncompress.pprof")
		if e != nil {
			log.Fatal(e)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
	}

	now := time.Now()
	if err = codec.Uncompress(in, inpos, len(in), out, outpos); err != nil {
		return 0, nil, err
	}
	since := time.Since(now).Nanoseconds()

	if prof {
		pprof.StopCPUProfile()
	}

	return since, out[:outpos.Get()], nil
}

func testCodecs(codecs map[string]encoding.Integer, data [][]int32, max int, output bool) error {
	compdata := make([]int32, max+max/2)
	decompdata := make([]int32, max)

	for name, codec := range codecs {
		for i, in := range data {
			k := len(in)

			dur, out, err := compress(codec, in, compdata, k, pprofParam)
			if err != nil {
				return err
			}

			dur2, out2, err2 := uncompress(codec, out, decompdata, k, pprofParam)
			if err2 != nil {
				return err2
			}

			if output {
				fmt.Printf("% 20s % 20s: %5.2f %5.2f %5.2f\n", files[i], name, float64(len(out)*32)/float64(k), (float64(k) / (float64(dur) / 1000000000.0) / 1000000.0), (float64(k) / (float64(dur2) / 1000000000.0) / 1000000.0))
			}

			for i := 0; i < k; i++ {
				if in[i] != decompdata[i] {
					return fmt.Errorf("benchmark/testCodecs: Problem recovering. index = %d, in = %d, recovered = %d, original length = %d, recovered length = %d\n", i, in[i], out2[i], k, len(out2))
				}
			}

			runtime.GC()
		}
	}

	return nil
}

func main() {
	flag.Parse()
	files = getListOfFiles()

	codecs, err := getListOfCodecs()
	if err != nil {
		log.Fatal(err)
	}

	data, max, err := loadIntegerFromFiles(files)
	if err != nil {
		log.Fatal(err)
	}

	if err := testCodecs(codecs, data, max, false); err != nil {
		log.Fatal(err)
	}

	if err := testCodecs(codecs, data, max, false); err != nil {
		log.Fatal(err)
	}

	if err := testCodecs(codecs, data, max, true); err != nil {
		log.Fatal(err)
	}
}
