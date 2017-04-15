// mdreplace is a tool for automating the maintenance of markdown files
package main // import "myitcv.io/g/cmd/mdreplace"

import (
	"flag"
	"fmt"
	"os"
)

// take input from stdin or files (args)
// -w flag will write back to files (error if stdin)
// if no -w flag, write to stdout
//
// lines must start with either:
//
// <!-- CODE: command args... -->
// <!-- REPLACE: command args... -->
//
// Commands must be valid or block will be ignored (log warning to stderr)
// Blocks must be terminated with exactly
//
// <!-- END -->
//
// No nesting supported; fatal if nested; fatal if block not terminated

// cmdStr
// cmd
// *template
// drain block (ignore comments in code sections)
// provide methods cmdStr, cmdStdOut and tmpl for the command string, command stdout and template

// provide blocks __TEMPLATE, __JSON, __LINES
// each block takes provides a transform method that transforms the stdout (io.Reader) into an interface
// and an error
// Decode(io.Reader) interface{}, error
// this decoded value is passed onto the template

// tests:
// 1. each block with a template
// 2. each block without
// 3. each block with a code block

// Limits:
// Block starts and ends must start and end the line (TODO can we remove this?)

// const (
// 	eof = -1

// 	commStart = "<!--"
// 	commEnd   = "-->"

// 	tagTmpl  = "__TEMPLATE"
// 	tagJson  = "__JSON"
// 	tagLines = "__LINES"

// 	end = "END"

// 	endBlock = commStart + " " + end + " " + commEnd

// 	debug = true
// )

// var (
// 	fHelp  = flag.Bool("h", false, "show usage information")
// 	fWrite = flag.Bool("w", false, "whether to write back to input files (cannot be used when reading from stdin)")
// )

// var (
// 	codeBlock = newSimpleBlock(codePrefix, codeSuffix, endBlock, func(s string) string { return "```\n" + s + "```" })
// 	replBlock = newSimpleBlock(replPrefix, replSuffix, endBlock, func(s string) string { return s })

// 	blocks = []block{codeBlock, replBlock}
// )

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:

  mdreplace file1 file2 ...
  mdreplace

When called with no file arguments, mdreplace works with stdin

Flags:
`)
	flag.PrintDefaults()
}

// type stateFn func() stateFn

func main() {
	flag.Usage = usage
	flag.Parse()

	// if *fHelp {
	// 	usage()
	// 	os.Exit(0)
	// }

	// r := newRunner(os.Stdin)
	// done := make(chan error)

	// go r.process(os.Stdout, done)

	// err := <-done

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }
}

// type runner struct {
// 	file *os.File
// 	hash hash.Hash

// 	items chan item

// 	b *simpleBlock

// 	input string
// 	start int
// 	pos   int
// 	line  int
// 	width *int
// }

// func newRunner(in *os.File) *runner {
// 	h := sha1.New()

// 	t := io.TeeReader(in, h)

// 	c, err := ioutil.ReadAll(t)
// 	if err != nil {
// 		fatalf("could not read from input %v: %v", in.Name(), err)
// 	}

// 	return &runner{
// 		file: in,
// 		hash: hash,

// 		items: make(chan item),

// 		input: string(c),
// 	}
// }

// func (r *runner) emit() {
// 	r.items <- item{
// 		out: r.input[r.start:r.pos],
// 	}
// 	r.start = r.pos
// }

// func (r *runner) drainLine() stateFn {
// 	for {
// 		if r.input[r.pos] == '\n' {
// 			r.pos++
// 			r.emit()

// 			return r.readText
// 		}

// 		if r.next() == eof {
// 			break
// 		}
// 	}

// 	if r.pos > r.start {
// 		r.emit()
// 	}
// }

// func (r *runner) readText() stateFn {
// 	for _, b := range blocks {
// 		if b.isHeader(line) {
// 			r.b = b
// 			return r.readBlockStart
// 		}
// 	}

// 	return r.drainLine
// }

// func (r *runner) next() (ru int) {
// 	if r.pos >= len(r.input) {
// 		r.width = 0
// 		return eof
// 	}

// 	ru, r.width = utf8.DecodeRuneInString(r.input[r.pos:])
// 	r.pos += r.width

// 	return ru
// }

// func (r *runner) readBlockStart() stateFn {
// 	r.pos += len(r.b.hPrefix)
// 	r.emit()

// 	for {
// 		if r.input[r.pos] == '\n' {
// 			break
// 		}

// 		if strings.HasPrefix(r.input, commEnd) {

// 		}
// 	}

// 	end := strings.Index(cmdString, commEnd)
// 	var rem string

// 	if end != -1 {
// 		rem, cmdStr = cmdStr[end:], cmdStr[0:end]
// 		r.unread(rem)
// 	}

// 	r.b.cmdStr = cmdStr

// 	cmd, err := split(cmdStr)
// 	if err != nil {
// 		return r.errorf("could not parse command: %v", err)
// 	}

// 	if end != -1 {
// 		return r.drainBlock
// 	}

// 	for {
// 		line, err := r.next()

// 		if r.b.isEnd(line) {
// 			return r.readBlockEnd
// 		}
// 	}

// 	return r.drainBlock
// }

// func (r *runner) readBlockEnd() stateFn {
// 	res, err := r.b.explode()
// 	if err != nil {
// 		return r.errorf("could not replace block: %v", err)
// 	}

// 	r.emit(res)

// 	return r.readText
// }

// func (r *runner) drainBlock() stateFn {
// 	for {
// 		// in this state we must read first
// 		line, err := r.next()

// 		if err != nil {
// 			if err == errEof {
// 				return r.errorf("expecting end of block, saw end of file")
// 			}

// 			return r.errorf("error getting next line: %v", err)
// 		}

// 		if r.b.isEnd(line) {
// 			return r.readBlockEnd
// 		}

// 		for _, b := range blocks {
// 			if b.isHeader(line) {
// 				return r.errorf("unexpected nested header")
// 			}
// 		}

// 		if err := r.b.sink(line); err != nil {
// 			return r.errorf("could not sink block line: %v", err)
// 		}
// 	}
// }

// func (r *runner) errorf(format string, args ...interface{}) stateFn {
// 	r.items <- item{
// 		err: fmt.Errorf("%v:%v %v", r.input.Name(), r.lno, fmt.Sprintf(format, args...)),
// 	}

// 	return nil
// }

// type item struct {
// 	out string
// 	err error
// }

// func (r *runner) process(out io.Writer, done chan error) {
// 	go func() {
// 		for state := r.readText; state != nil; {
// 			state = state()
// 		}
// 		close(r.items)
// 	}()

// 	for i := range r.items {
// 		// any error means there will be nothing else coming through
// 		// the items channel
// 		if i.err != nil {
// 			if i.err == errEof {
// 				continue
// 			}

// 			done <- i.err
// 			return
// 		}

// 		_, err := fmt.Fprintln(out, i.out)
// 		if err != nil {
// 			// TODO don't think we can recover in this instance...
// 			done <- fmt.Errorf("unable to write to output: %v", err)
// 			return
// 		}
// 	}

// 	done <- nil
// }

// func (r *runner) readCmd() stateFn {
// 	var words []string

// Words:
// 	for {
// 		r.acceptRun(" \t")

// 		if r.input[r.pos] == '"' {
// 			for i := 1; i < len(line); i++ {
// 				c := line[i] // Only looking for ASCII so this is OK.
// 				switch c {
// 				case '\\':
// 					if i+1 == len(line) {
// 						return nil, fmt.Errorf("bad backslash")
// 					}
// 					i++ // Absorb next byte (If it's a multibyte we'll get an error in Unquote).

// 				case '"':
// 					word, err := strconv.Unquote(line[0 : i+1])
// 					if err != nil {
// 						return nil, fmt.Errorf("bad quoted string")
// 					}
// 					words = append(words, word)
// 					line = line[i+1:]

// 					// Check the next character is space or end of line.
// 					if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
// 						return nil, fmt.Errorf("expect space after quoted argument")
// 					}
// 					continue Words
// 				}
// 			}
// 			return nil, fmt.Errorf("mismatched quoted string")
// 		}

// 		i := strings.IndexAny(line, " \t")
// 		if i < 0 {
// 			i = len(line)
// 		}
// 		words = append(words, line[0:i])
// 		line = line[i:]
// 	}

// 	return words, nil
// }

// func (r *runner) acceptRun(valid string) {
// 	for strings.IndexRune(valid, r.next()) >= 0 {
// 	}
// 	r.backup()
// }

// func (r *runner) peek() int {
// 	rune := r.next()
// 	r.backup()
// 	return rune
// }

// func (r *runner) backup() {
// 	if r.width == nil {
// 		// TODO this is ugly
// 		panic("tried to backup twice")
// 	}

// 	r.pos -= *r.width
// 	r.width = nil
// }

// func (r *runner) ignore() {
// 	r.start = r.pos
// }

// func debugf(format string, args ...interface{}) {
// 	if debug {
// 		log.Printf(format, args...)
// 	}
// }
