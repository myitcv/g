// concsh allows you to concurrently run commands from your shell.
package main // import "myitcv.io/g/cmd/concsh"

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// all args after the first -- are then considered as a --- (notice the extra -) separated
// list of commands to be run concurrently
//
// output is interleaved on a line-by-line basis
//
// there is no shell evaluation of arguments
//
// TODO add ability to limit concurrency (probably not worth it for things like git which are most
// likely to be network-bounded)
// TODO could support shell evalulation of lines (command line version already covered?)?
// TODO improve panics; some situations we might be able to better detect/handle?
// TODO add some mode whereby commands are executed only if all commands are valid (means
// that stdin read commands not executed until stdin is closed)
//
// exit code is 0 if all commands succeed without error; one of the non-zero exit codes otherwise

type result struct {
	exitCode int
	lines    []outLine
}

type outLine struct {
	err  bool
	line string
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix(os.Args[0] + ": ")

	flag.Usage = usage

	flag.Parse()

	var argSets [][]string

	s := 0
	nr := 0
	results := make(chan result)

	if len(os.Args) == 1 {
		// read from stdin
		sc := bufio.NewScanner(os.Stdin)
		line := 1

		for sc.Scan() {
			args, err := split(sc.Text())
			if err != nil {
				infof("could not parse command on line %v: %v", line, err)
			}

			nr += runCmd(args, results)
			line++
		}
		if err := sc.Err(); err != nil {
			fatalf("unable to read from stdin: %v", err)
		}
	} else {
		for i, v := range os.Args {
			if v == "--" {
				s = i
				break
			}
		}

		var args []string

		for _, v := range os.Args[s+1:] {
			if v == "---" {
				argSets = append(argSets, args)
				args = nil
			} else {
				args = append(args, v)
			}
		}

		// in case we did not have a final ---
		argSets = append(argSets, args)

	}
	// nr += runCmd(args, results)
	// if nr == 0 {
	// 	fmt.Println("No commands to execute")
	// 	return
	// }

	exitCode := 0

	for res := range results {

		for _, v := range res.lines {
			if v.err {
				fmt.Fprint(os.Stderr, v.line)
			} else {
				fmt.Fprint(os.Stdout, v.line)
			}
		}

		if res.exitCode != 0 {
			exitCode = res.exitCode
		}

		nr--
		if nr == 0 {
			break
		}
	}

	os.Exit(exitCode)
}

func usage() {
	fmt.Fprintln(os.Stderr, `concsh allows you to concurrently run commands from your shell

Usage:
	concsh -- comand1 arg1_1 arg1_2 ... --- command2 arg2_1 arg 2_2 ... --- ...
	concsh

In the case no arguments are provided, concsh will read the commands to execute from stdin, one per line
	`)

	flag.PrintDefaults()
}

func runCmd(args []string, results chan result) int {
	res := 0

	if len(args) > 0 {
		res = 1
		go runCmdImpl(args, results)
	}

	return res
}

// based on the nice clean, algorithm in go generate
// https://github.com/golang/go/blob/c1730ae424449f38ea4523207a56c23b2536a5de/src/cmd/go/generate.go#L292

func split(line string) ([]string, error) {
	var words []string

Words:
	for {
		line = strings.TrimLeft(line, " \t")
		if len(line) == 0 {
			break
		}
		if line[0] == '"' {
			for i := 1; i < len(line); i++ {
				c := line[i] // Only looking for ASCII so this is OK.
				switch c {
				case '\\':
					if i+1 == len(line) {
						return nil, fmt.Errorf("bad backslash")
					}
					i++ // Absorb next byte (If it's a multibyte we'll get an error in Unquote).

				case '"':
					word, err := strconv.Unquote(line[0 : i+1])
					if err != nil {
						return nil, fmt.Errorf("bad quoted string")
					}
					words = append(words, word)
					line = line[i+1:]

					// Check the next character is space or end of line.
					if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
						return nil, fmt.Errorf("expect space after quoted argument")
					}
					continue Words
				}
			}
			return nil, fmt.Errorf("mismatched quoted string")
		}
		i := strings.IndexAny(line, " \t")
		if i < 0 {
			i = len(line)
		}
		words = append(words, line[0:i])
		line = line[i:]
	}

	return words, nil
}

func runCmdImpl(args []string, results chan result) {

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	readDone := make(chan struct{})

	outres := make(chan string)
	errres := make(chan string)

	go read(stdout, outres, readDone)
	go read(stderr, errres, readDone)

	// because we have one process, two pipes to read from
	pc := 2

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	var lines []outLine

	for {
		select {
		case <-readDone:
			pc--
		case s := <-outres:
			lines = append(lines, outLine{
				line: s,
			})
		case s := <-errres:
			lines = append(lines, outLine{
				err:  true,
				line: s,
			})
		}

		if pc == 0 {
			break
		}
	}

	res := result{
		lines: lines,
	}

	err = cmd.Wait()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			panic(err)
		}

		exitCode := 0

		switch ws := ee.Sys().(type) {
		case syscall.WaitStatus:
			exitCode = ws.ExitStatus()
		default:
			panic(fmt.Errorf("Need to add case for %T", ws))
		}

		res.exitCode = exitCode
	}

	results <- res
}

func read(in io.ReadCloser, res chan string, done chan struct{}) {
	b := bufio.NewReader(in)

	for {
		line, err := b.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line != "" {
					res <- line + "\n"
				}
			}

			// notice we are ignoring io errors... because any other
			// errors will be handled by the Wait on the cmd with a non-zero
			// exit code

			break
		}

		res <- line
	}

	done <- struct{}{}
}

func fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}
