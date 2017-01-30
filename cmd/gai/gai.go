package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kisielk/gotool"
)

const (
	// goBuildPkgHeader is the line prefix used in the output of go install etc
	// the starts a block of line errors
	goBuildPkgHeader = "# "

	// maxGoType
	maxGoType = 50
)

var (
	fDebug = flag.Bool("v", false, "be very verbose about what gai is doing")
)

type multiFlag []string

func (i *multiFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *multiFlag) String() string {
	return fmt.Sprint(*i)
}

var (
	fDeps multiFlag
)

func init() {
	flag.Var(&fDeps, "D", "file listing dependencies; may appear multiple times")
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gai: ")
	flag.Parse()

	defer func() {
		// if err, ok := recover().(error); ok {
		// 	log.Fatalln(err)
		// }
	}()

	wd, err := os.Getwd()
	if err != nil {
		fatalf("could not get working directory: %v", err)
	}

	fArgs := flag.Args()
	specs := make([]string, len(fArgs))

	copy(specs, fArgs)

	for _, fn := range fDeps {
		specs = append(specs, readLines(fn)...)
	}

	pkgs := gotool.ImportPaths(specs)

	if len(pkgs) == 0 {
		fatalf("could not resolve any import paths from %v", specs)
	}

	pkgs, roots, all := buildDeps(pkgs)

	allFlat := make([]string, 0, len(all))
	for k := range all {
		allFlat = append(allFlat, k)
	}

	getPkgFails := func(vs []string) []string {
		var res []string
		for _, line := range vs {
			if strings.HasPrefix(line, goBuildPkgHeader) {
				line = strings.TrimPrefix(line, goBuildPkgHeader)
				res = append(res, line)
			}
		}
		return res
	}

	var failWork []*depPkg

	failedInstalls := getPkgFails(goDo(allFlat, "go", "install"))
	for _, v := range failedInstalls {
		p := all[v]

		p.buildStatus = statusFailed
		failWork = append(failWork, p)
	}

	var f *depPkg

	for len(failWork) != 0 {
		f, failWork = failWork[0], failWork[1:]
		for vv := range f.rdeps {
			if vv.buildStatus == statusPassed {
				vv.buildStatus = statusDepFailed
				failWork = append(failWork, vv)
			}
		}
	}

	res := make(chan string)
	gotyped := make(map[*depPkg]bool)

	gotFail := false

	done := make(chan bool)

	go func() {
		for {
			v, ok := <-res
			if !ok {
				break
			}

			if v != "" {
				gotFail = true
				fmt.Fprint(os.Stderr, v)
			}
		}

		close(done)
	}()

	var wg sync.WaitGroup

	togotype := make(map[string]bool)
	for _, v := range pkgs {
		togotype[v] = true
	}
	for _, v := range failedInstalls {
		togotype[v] = true
	}

	for ip := range togotype {

		v := all[ip]

		if v.buildStatus == statusDepFailed {
			continue
		}

		rd := v.pkg.Dir

		if filepath.IsAbs(rd) {
			r, err := filepath.Rel(wd, rd)
			if err != nil {
				fatalf("could not calculate filepath.Rel(%q, %q): %v", wd, rd, err)
			}

			rd = r
		}

		wg.Add(1)

		go func(ip, dir string) {
			out := ""
			r := goDo([]string{dir}, "gotype", "-a")

			// sort the lines
			splits := make(linesByNumber, len(r))
			for i := range r {
				splits[i] = strings.SplitN(r[i], ":", 4)
			}

			sort.Sort(splits)

			for i := range r {
				r[i] = strings.Join(splits[i], ":")
			}

			if len(r) > 0 {
				out = fmt.Sprintf("# %v\n%v\n", ip, strings.Join(r, "\n"))
			}

			res <- out
			wg.Done()

		}(v.ip, rd)

		if v.buildStatus == statusPassed {
			for d := range v.rdeps {

				if d.buildStatus == statusDepFailed {
					continue
				}

				if !gotyped[d] {
					gotyped[d] = true
					roots = append(roots, d)
				}
			}
		}
	}

	wg.Wait()

	close(res)

	<-done

	if gotFail {
		os.Exit(1)
	}
}

type status uint

const (
	statusPassed status = iota
	statusFailed
	statusDepFailed
)

type depPkg struct {
	ip string

	pkg         *build.Package
	buildStatus status

	deps  map[*depPkg]struct{}
	rdeps map[*depPkg]struct{}
}

func newDepPkg(ip string) *depPkg {
	return &depPkg{
		ip:    ip,
		deps:  make(map[*depPkg]struct{}),
		rdeps: make(map[*depPkg]struct{}),
	}
}

func buildDeps(specs []string) ([]string, []*depPkg, map[string]*depPkg) {
	var pNames []string
	var nonCore []*depPkg
	seen := make(map[string]*depPkg)
	toDo := make([]*depPkg, 0, len(specs))

	wd, err := os.Getwd()
	if err != nil {
		fatalf("could not get working directory: %v", err)
	}

	for _, v := range specs {

		pkg, err := build.Import(v, wd, 0)
		if err != nil {
			fatalf("could not import %v: %v", v, err)
		}
		p := newDepPkg(pkg.ImportPath)
		p.pkg = pkg

		toDo = append(toDo, p)
		seen[pkg.ImportPath] = p
		pNames = append(pNames, pkg.ImportPath)
	}

	var p *depPkg

	for len(toDo) != 0 {
		p, toDo = toDo[0], toDo[1:]

		if p.pkg == nil {
			pkg, err := build.Import(p.ip, wd, 0)
			if err != nil {
				fatalf("could not import %v: %v", p.ip, err)
			}
			p.pkg = pkg
		}

		if p.pkg.Goroot {
			continue
		}

		nonCore = append(nonCore, p)

		var toCheck []string
		toCheck = append(toCheck, p.pkg.Imports...)
		toCheck = append(toCheck, p.pkg.TestImports...)

		for _, i := range toCheck {
			t, ok := seen[i]
			if !ok {
				t = newDepPkg(i)
				toDo = append(toDo, t)
				seen[i] = t
			}

			p.deps[t] = struct{}{}
			t.rdeps[p] = struct{}{}
		}
	}

	// find the roots
	var roots []*depPkg
	for _, v := range nonCore {

		hasNonCore := false

		for d := range v.deps {
			if !d.pkg.Goroot {
				hasNonCore = true
			}
		}

		if !hasNonCore {
			roots = append(roots, v)
		}
	}

	return pNames, roots, seen
}

func goDo(specs []string, c string, args ...string) []string {
	// we don't have a good way of detecting whether a failure of go install
	// is "good" or "bad"; exit codes might tell us something but this does
	// not appear to be documented; maybe rsc's work on the cmd/go suite will
	// change this situation

	var cArgs []string
	cArgs = append(cArgs, args...)
	cArgs = append(cArgs, specs...)

	cmd := exec.Command(c, cArgs...)
	cmdStr := fmt.Sprintf("%v %v", c, strings.Join(cArgs, " "))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fatalf("could not create stderr pipe for %v: %v", cmdStr, err)
	}

	var res []string

	cmd.Start()

	sc := bufio.NewScanner(stderr)
	for sc.Scan() {
		line := sc.Text()
		res = append(res, line)
	}

	if err := sc.Err(); err != nil {
		fatalf("scan error reading %v: %v", cmdStr, err)
	}

	// we don't care for the failed results...
	cmd.Wait()

	infof("running %v %v", cmdStr, res)

	return res
}

type linesByNumber [][]string

func (l linesByNumber) Len() int      { return len(l) }
func (l linesByNumber) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l linesByNumber) Less(i, j int) bool {
	li, lj := l[i], l[j]

	// each line will have at least one part
	if i := strings.Compare(l[i][0], l[j][0]); i < 0 {
		return true
	}

	// in case they both had just a single part
	linei, err := strconv.ParseUint(li[1], 10, 64)
	if err != nil {
		return i < j
	}
	linej, err := strconv.ParseUint(lj[1], 10, 64)
	if err != nil {
		return i < j
	}
	if linei < linej {
		return true
	}

	coli, err := strconv.ParseUint(li[1], 10, 64)
	if err != nil {
		return i < j
	}
	colj, err := strconv.ParseUint(lj[1], 10, 64)
	if err != nil {
		return i < j
	}
	if coli < colj {
		return true
	}

	return i < j
}

func readLines(file string) []string {
	var fi *os.File
	var err error

	if file == "-" {
		fi = os.Stdin
	} else {
		fi, err = os.Open(file)
		if err != nil {
			fatalf("could not open %v: %v", file, err)
		}
	}

	sc := bufio.NewScanner(fi)
	var res []string

	for sc.Scan() {
		res = append(res, sc.Text())
	}
	if err = sc.Err(); err != nil {
		fatalf("could not scan %v: %v", file, err)
	}

	return res
}

func resolvePkgSpec(spec []string) []string {
	var res []string

	args := append([]string{"list", "-e"}, spec...)

	gl := exec.Command("go", args...)

	// we only care for stdout
	out, err := gl.Output()
	if err != nil {
		fatalf("could not run go list: %v", err)
	}

	buf := bytes.NewBuffer(out)

	sc := bufio.NewScanner(buf)

	for sc.Scan() {
		res = append(res, sc.Text())
	}

	return res
}

func infof(format string, args ...interface{}) {
	if *fDebug {
		log.Printf(format, args...)
	}
}

func fatalf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}
