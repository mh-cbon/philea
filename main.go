package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/docopt/docopt.go"
	"github.com/mh-cbon/verbose"
	"github.com/mattn/go-zglob"
)

var logger = verbose.Auto()

func main() {
	usage := `Philea - Apply commands on globbed files

Usage:
  philea [options] <cmds>...
  philea [-q|--quiet] [-e <pattern> | --exclude=<pattern>] <cmds>...
  philea [-q|--quiet] [-p <pattern> | --pattern=<pattern>] <cmds>...
  philea -q | --quiet
  philea -h | --help
  philea -v | --version

Options:
  -h --help             Show this screen.
  -v --version          Show version.
  -q --quiet            Less verbose.
  -e --exclude pattern  Exclude files from being processed [default: *vendor/*].
  -p --pattern pattern  Which kind of files to process [default: **/*.go].

Notes:
  cmd can contain %s, it will be replaced by the current file.
  philea will process all files and all commands and return an exit code=1 if any fails.

Examples:
  philea "cat %s" "grep t %s"
    It will process all go files, except those in vendors, and apply
    cat, then grep an each file.

  `

	arguments, err := docopt.Parse(usage, nil, true, "Philea 0.0.1", false)
	logger.Println(arguments)
	exitWithError(err)

	wd, err := os.Getwd()
	logger.Println("wd=" + wd)
	exitWithError(err)

	cmds := getCmds(arguments)
	exclude := getExclude(arguments)
	pattern := getPattern(arguments)
	quiet := isQuiet(arguments)

	if len(cmds) == 0 {
		exitWithError(errors.New("There is no commands to execute"))
	}

	excludeRe, err := getExlcudeRe(exclude, false)
	exitWithError(err)
	logger.Println("excludeRe=", excludeRe)
	paths, err := zglob.Glob(wd + "/" + pattern)
	logger.Println("paths=", paths)

	if len(paths) == 0 {
		exitWithError(errors.New("No matching files found in this directory"))
	}

	filteredPaths := filterPaths(paths, excludeRe)
	logger.Println("filteredPaths=", filteredPaths)

	if len(paths) == 0 {
		exitWithError(errors.New("Pattern has excluded all files!"))
	}

	errs := make([]error, 0)
	outs := make([]string, 0)
	var wg sync.WaitGroup
	for _, p := range filteredPaths {
		for _, c := range cmds {
			wg.Add(1)
			go func(path string, cmd string) {
				out, err := executeACommand(path, cmd)
				outs = append(outs, printOut(cmd, string(out), path, wd))
				if err != nil {
					errs = append(errs, err) // racy ?
				}
				wg.Done()
			}(p, c)
		}
	}
	wg.Wait()

	if quiet == false {
		for _, out := range outs {
			fmt.Print(out)
		}
		if len(errs) > 0 {
			fmt.Println("-------------")
			fmt.Println("There was errors: ", len(errs))
			for _, err := range errs {
				fmt.Println(err)
			}
		}
	}
	if len(errs) > 0 {
		os.Exit(1)
	}
}

func exitWithError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printOut(cmd string, out string, path string, wd string) string {
	ret := ""
	path = strings.Replace(path, wd+string(os.PathSeparator), "", -1)
	cmd = strings.Replace(cmd, "%s", path, -1)
	for _, line := range strings.Split(out, "\n") {
		ret = ret + cmd + ": " + line + "\n"
	}
	if ret == "" {
		ret = cmd + ": empty output"
	}
	return ret
}

func executeACommand(path string, cmd string) ([]byte, error) {
	cmd = strings.Replace(cmd, "%s", path, -1)
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return make([]byte, 0), errors.New("Failed to parse cmd: " + cmd)
	}
	bin, err := exec.LookPath(parts[0])
	if err != nil {
		return make([]byte, 0), err
	}
	oCmd := exec.Command(bin, parts[1:]...)
	return oCmd.CombinedOutput()
}

func filterPaths(paths []string, exclude *regexp.Regexp) []string {
	filtered := make([]string, 0)
	for _, v := range paths {
		if exclude.MatchString(v) == false {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func getExlcudeRe(exclude string, sensitive bool) (*regexp.Regexp, error) {
	exclude = strings.Replace(exclude, "*", ".+", -1)
	flags := ""
	if sensitive == false {
		flags = "(?i)"
	}
	return regexp.Compile(flags + exclude)
}

func getCmds(arguments map[string]interface{}) []string {
	cmds := make([]string, 0)
	if c, ok := arguments["<cmds>"].([]string); ok {
		cmds = append(cmds, c...)
	}
	return cmds
}

func getExclude(arguments map[string]interface{}) string {
	exclude := "*vendor/*"
	if val, ok := arguments["--exclude"].(string); ok {
		exclude = val
	} else {
		if v, ok := arguments["-e"].(string); ok {
			exclude = v
		}
	}
	return exclude
}

func getPattern(arguments map[string]interface{}) string {
	pattern := "**/*.go"
	if val, ok := arguments["--pattern"].(string); ok {
		pattern = val
	} else {
		if v, ok := arguments["-p"].(string); ok {
			pattern = v
		}
	}
	return pattern
}

func isQuiet(arguments map[string]interface{}) bool {
	quiet := false
	if isQuiet, ok := arguments["--quiet"].(bool); ok {
		quiet = isQuiet
	} else {
		if isQ, ok := arguments["-q"].(bool); ok {
			quiet = isQ
		}
	}
	return quiet
}
