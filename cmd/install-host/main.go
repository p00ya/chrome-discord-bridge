// package main implements a command-line utility for registering a Chrome
// native messaging host with Chrome.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n"+
		"%s [-system] [-o ORIGIN]... [-d DESC] NAME BINARY\n\n", os.Args[0])
	flag.PrintDefaults()
}

const (
	exitSuccess      = 0
	exitInvalidUsage = 1
	exitFailure      = 2
)

type originList []string

func (ss *originList) String() string {
	return strings.Join(*ss, ", ")
}

// Set appends the value to the set.
func (ss *originList) Set(value string) error {
	if _, err := url.Parse(value); err != nil {
		return errors.New("invalid URL")
	}
	*ss = append(*ss, value)
	return nil
}

func validateName(name string) (ok bool) {
	ok, _ = regexp.MatchString(`^([a-z0-9_]+)(\.[a-z0-9_]+)*$`, name)
	return
}

func main() {
	sys := flag.Bool("system", false, "Install system-wide (instead of for current user)")
	desc := flag.String("d", "", "Host description")
	var origins originList
	flag.Var(&origins, "o", "Allowed-origin URL.  Repeat flag for multiple URLs")

	flag.Usage = printUsage
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Error: expected 2 arguments, got %d\n", flag.NArg())
		printUsage()
		os.Exit(exitInvalidUsage)
	}

	name := flag.Arg(0)
	if !validateName(name) {
		fmt.Fprintf(os.Stderr, "Error: invalid host name \"%s\"\n", name)
		os.Exit(exitInvalidUsage)
	}
	binary := flag.Arg(1)
	switch fi, err := os.Stat(binary); {
	case err != nil:
		fmt.Fprintf(os.Stderr, "Warning: accessing binary: %v\n", err)
	case fi.Mode()&0100 == 0:
		fmt.Fprintf(os.Stderr, "Warning: binary %s is not executable\n", binary)
	}
	absPath, err := filepath.Abs(binary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving absolute path to %s\n", binary)
		os.Exit(exitFailure)
	}

	m := Manifest{
		Name:           name,
		Description:    *desc,
		Path:           absPath,
		AllowedOrigins: origins,
	}

	if *sys {
		err = InstallSystem(m)
	} else {
		err = InstallCurrentUser(m)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitFailure)
	}

	fmt.Printf("Wrote manifest for %s\n", name)
}
