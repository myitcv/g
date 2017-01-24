// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	_log "log"
	"os"
	"path/filepath"
)

var (
	fVerbose = flag.Bool("v", false, "log interesting messages")
	fFix     = flag.Bool("f", false, "fix the files passed as arguments")
	fIndent  = flag.String("indent", "\t", "the indent string")
	fPrefix  = flag.String("prefix", "", "the prefix string")
)

type Config struct {
	Prefix string
	Indent string
}

const (
	ConfigFileName = ".jsonlintconfig.json"
)

const (
	logPrefix = ""
	logFlags  = 0
)

var config Config
var log = _log.New(os.Stdout, logPrefix, logFlags)
var elog = _log.New(os.Stderr, logPrefix, logFlags)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	flag.Parse()

	files := flag.Args()

	if len(files) == 0 {
		files = []string{os.Stdin.Name()}
	}

	f := &formatter{}

	for _, file := range files {
		f.file = file
		f.format()
	}

	if f.failed {
		os.Exit(1)
	}
}

type formatter struct {
	failed bool
	file   string
}

var procFile = fmt.Errorf("error handling file")

func (f *formatter) failf(format string, args ...interface{}) {
	f.failed = true
	var fn string

	if f.file == os.Stdin.Name() {
		fn = "<stdin>"
	} else {
		fn = f.file
	}

	elog.Printf("%v: %v\n", fn, fmt.Sprintf(format, args...))

	panic(procFile)
}

func (f *formatter) format() {
	var file *os.File
	var err error

	defer func() {
		if err := recover(); err != nil && err != procFile {
			panic(err)
		}

		if file != nil {
			file.Close()
		}
	}()

	if f.file == os.Stdin.Name() {
		file = os.Stdin
	} else {
		file, err = os.Open(f.file)
		if err != nil {
			f.failf("unable to open: %v", err)
		}
	}

	var r io.Reader
	var orig []byte

	if !*fFix {
		cs, err := ioutil.ReadAll(file)
		if err != nil {
			f.failf("unable to read: %v", err)
		}

		orig = cs
		r = bytes.NewBuffer(orig)
	} else {
		r = file
	}

	dec := json.NewDecoder(r)

	var j interface{}
	err = dec.Decode(&j)

	if *fFix {
		clErr := file.Close()
		if clErr != nil {
			f.failf("unable to close: %v", err)
		}
	}

	if err != nil {
		f.failf("does not contain valid JSON: %v", err)
	}

	c := deriveConfig(file)

	if *fVerbose {
		log.Printf("For file %v using config %#v\n", file.Name(), c)
	}

	b, err := json.MarshalIndent(j, c.Prefix, c.Indent)
	if err != nil {
		f.failf("could not be formatted: %v", err)
	}

	b = append(b, []byte("\n")...)

	if *fFix {
		if file == os.Stdin {
			_, err = os.Stdout.Write(b)
		} else {
			err = ioutil.WriteFile(file.Name(), b, 0644)
		}
		if err != nil {
			f.failf("could not write formatted JSON back to file: %v", err)
		}
	} else {
		if !bytes.Equal(orig, b) {
			f.failf("is not well-formatted")
		}
	}
}

func deriveConfig(file *os.File) (res Config) {
	var err error

	res.Indent = *fIndent
	res.Prefix = *fPrefix

	var dir string

	if file == os.Stdin {
		d, err := os.Getwd()
		if err != nil {
			elog.Fatalf("Could not get working directory: %v", err)
		}

		dir = d
	} else {
		abs, err := filepath.Abs(file.Name())
		if err != nil {
			elog.Fatalf("Could not get absolute path to %v: %v", file.Name(), err)
		}

		dir = filepath.Dir(abs)
	}

	var fi *os.File

	for {
		fp := filepath.Join(dir, ConfigFileName)

		if *fVerbose {
			log.Printf("Checking for config file %v\n", fp)
		}

		fi, err = os.Open(fp)

		if err == nil {
			break
		}

		p := filepath.Dir(dir)

		if p == dir {
			break
		}

		dir = p
	}

	if fi == nil {
		return
	}

	if *fVerbose {
		log.Printf("Found config file at %v\n", fi.Name())
	}

	dec := json.NewDecoder(fi)
	err = dec.Decode(&res)

	clErr := fi.Close()

	if clErr != nil {
		elog.Fatalf("Could not close file %v: %v", fi.Name(), clErr)
	}

	if err != nil {
		elog.Fatalf("Unable to decode config from %v: %v", fi.Name(), err)
	}

	return res
}
