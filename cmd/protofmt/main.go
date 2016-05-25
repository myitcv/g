package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	fHelpShort   = flag.Bool("h", false, "Show usage text (same as --help).")
	fHelpLong    = flag.Bool("help", false, "Show usage text (same as -h).")
	fImportPaths = importPaths([]string{"."})
)

func init() {
	flag.Var(&fImportPaths, "I", "Path to search for imports (flag can be used multiple times)")
}

// TODO - this command should not require import paths etc... should be purely syntactical...
// for another day
func main() {
	flag.Usage = usage
	flag.Parse()
	if *fHelpShort || *fHelpLong || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, p := range fImportPaths {
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			fatalf("Import dir does not exist (as a directory)")
		}
	}

	if len(flag.Args()) == 0 {
		fatalf("Need to specify at least one file to parse")
	}

	f := &formatter{
		files:       flag.Args(),
		importPaths: fImportPaths,

		output: os.Stdout,
	}

	f.fmt()
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:  %s [options] <foo.proto> ...\n", os.Args[0])
	flag.PrintDefaults()
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
