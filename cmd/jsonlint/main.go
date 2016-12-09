// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var fFix = flag.Bool("f", false, "fix the files passed as arguments")
var fSpaces = new(spaces)

func init() {
	flag.Var(fSpaces, "s", "indent using the provided number of spaces")
}

type spaces struct {
	number *int
}

func (s *spaces) String() string {
	return fmt.Sprintf("%v", s.number)
}

func (s *spaces) Get() interface{} {
	return s.number
}

func (s *spaces) Set(toSet string) error {
	i, err := strconv.Atoi(toSet)
	if err != nil {
		return err
	}
	s.number = &i
	return nil
}

func main() {
	flag.Parse()

	var files []*os.File

	for _, fName := range flag.Args() {
		f, err := os.Open(fName)
		if err != nil {
			panic(err)
		}

		files = append(files, f)
	}
	if len(files) == 0 {
		files = append(files, os.Stdin)
	}

	fail := false
	for _, file := range files {
		dec := json.NewDecoder(file)

		var j interface{}
		err := dec.Decode(&j)
		if err != nil {
			fmt.Printf("%v does not contain valid JSON: %v\n", file.Name(), err)
			fail = true
			continue
		}

		indent := "\t"
		if fSpaces.number != nil {
			indent = strings.Repeat(" ", *fSpaces.number)
		}

		b, err := json.MarshalIndent(j, "", indent)
		if err != nil {
			fmt.Printf("%v could not be formatted\n", file.Name())
			fail = true
			continue
		}

		b = append(b, []byte("\n")...)

		if *fFix {
			if file == os.Stdin {
				_, err = os.Stdout.Write(b)
			} else {
				err = ioutil.WriteFile(file.Name(), b, 0644)
			}
			if err != nil {
				fmt.Printf("Could not write formatted JSON to %v\n", file.Name())
				fail = true
				continue
			}
		} else {
			_, err := file.Seek(0, 0)
			if err != nil {
				panic(err)
			}
			contents, err := ioutil.ReadAll(file)
			if err != nil {
				panic(err)
			}
			if string(contents) != string(b) {
				fmt.Printf("%v is not formatted\n", file.Name())
				fail = true
				continue
			}
		}
	}

	if fail {
		os.Exit(1)
	}
}
