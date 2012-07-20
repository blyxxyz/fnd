package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const alphaNum = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	regexpFlag        = flag.String("e", "", "Use regexp")
	filetypeFlag      = flag.String("type", "all", "Search only (f)iles, (d)irectories or (l)inks. Comma separated.")
	caseSensitiveFlag = flag.Bool("s", false, "Case sensitive search")
)

func showUsage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [pattern]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func readConfigFile(filePath string) map[string]string {
	var result map[string]string
	return result
}

//escape alpha numeric patterns
func regexpEscape(c byte) string {
	if bytes.IndexByte([]byte(alphaNum), c) == -1 {
		if c == '\000' {
			return "\\000"
		} else {
			return "\\" + string(c)
		}
	}
	return string(c)
}

// Translate *nix like pattern. Ex: *.py to .*\.py for regexp
func unixRegexp(pattern string) string {
	res := ""
	for _, c := range pattern {
		switch c {
		default:
			res = res + regexpEscape(byte(c))
		case '*':
			res = res + ".*"
		case '?':
			res = res + "."
		}
	}
	return res
}

func printFile(directory string, fileinfo os.FileInfo, stdout io.Writer) {
	filename := filepath.Join(directory, fileinfo.Name())

	if directory[0] != os.PathSeparator {
		filename = "." + string(os.PathSeparator) + filename
	}
	fmt.Fprintf(stdout, "%s\n", filename)
}

func parseDir(directory string, options map[string]string, stdout io.Writer) {
	dir, err := os.Open(directory)
	if err != nil { // can't open? just ignore it
		return
	}

	// this reads the whole content of the dir. It may not be a good idea.
	dirInfoSlice, _ := dir.Readdir(-1)
	for _, fileinfo := range dirInfoSlice {
		filename := fileinfo.Name()
		if options["caseSensitive"] == "false" {
			filename = strings.ToLower(filename)
		}
		ok := false
		if options["filetype"] != "all" {
			if options["filetype_f"] == "true" && !fileinfo.IsDir() && fileinfo.Mode() != os.ModeSymlink {
				ok = true
			}
			if options["filetype_d"] == "true" && fileinfo.IsDir() {
				ok = true
			}
			if options["filetype_l"] == "true" && fileinfo.Mode() == os.ModeSymlink {
				ok = true
			}
		} else {
			ok = true
		}
		if ok {
			matched, _ := regexp.Match(options["pattern"],
				[]byte(filename))
			if matched {
				printFile(directory, fileinfo, stdout)
			}
		}
		if fileinfo.IsDir() {
			parseDir(filepath.Join(directory, filename),
				options, stdout)
		}
	}
}

func Find(options map[string]string, stdout io.Writer) {
	if options["caseSensitive"] == "false" {
		options["pattern"] = strings.ToLower(options["pattern"])
	}
	options["filetype_f"] = "false"
	options["filetype_d"] = "false"
	options["filetype_l"] = "false"

	if options["filetype"] == "" { // just in case we don't have any option
		options["filetype"] = "all"
	}
	if options["filetype"] != "all" {
		for _, filetype := range strings.Split(options["filetype"], ",") {
			options["filetype_"+filetype] = "true"
		}
	}
	parseDir(options["directory"], options, stdout)
}

func main() {
	options := make(map[string]string)

	flag.Usage = showUsage
	flag.Parse()

	options["pattern"] = ""
	options["directory"] = "."
	options["caseSensitive"] = "false"
	if *caseSensitiveFlag {
		options["caseSensitive"] = "true"
	}
	options["filetype"] = *filetypeFlag

	if flag.NArg() == 1 {
		if *regexpFlag != "" { // fnd -e <regexp> <dir>
			options["pattern"] = *regexpFlag
			options["directory"] = flag.Arg(0)
		} else { // fnd <pattern>
			options["pattern"] = unixRegexp(flag.Arg(0))
		}
	}

	if flag.NArg() == 2 { // fnd <pattern> <dir>
		if *regexpFlag != "" {
			log.Fatal("Can't use both regexp and pattern")
		}
		options["pattern"] = unixRegexp(flag.Arg(0))
		options["directory"] = flag.Arg(1)
	}

	Find(options, os.Stdout)
}
