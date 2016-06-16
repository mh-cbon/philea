package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/docopt/docopt.go"
	"github.com/mattn/go-zglob"
	"github.com/mh-cbon/verbose"
)

var logger = verbose.Auto()
var VERSION = "0.0.0"

func main() {
	usage := `Philea - Apply commands on globbed files

Usage:
  philea [options] <cmds>...
  philea -q | --quiet
  philea -h | --help
  philea -v | --version

Options:
  -h --help               Show this screen.
  -v --version            Show version.
  -q --quiet              Less verbose.
  -e --exclude pattern    Exclude files from being processed [default: *vendor/*].
  -p --pattern pattern    Which kind of files to process [default: **/*.go].
  -C --change-dir dir     Change current working directory.
  -S --series             Execute process in series instead of parallel.
  -d --dry                Show commands only, do not run anything.

Notes:
  cmd can contain %s, it will be replaced by the current file.
  cmd can contain %d, it will be replaced by the directory path of the current file.
  cmd can contain %dname, it will be replaced by the directory name of the current file.
  philea will process all files and all commands and return an exit code=1 if any fails.

Examples:
  philea "cat %s" "grep t %s" "ls -al %d" "echo '%dname'"
    It will process all go files, except those in vendors, and apply
    cat, then grep an each file.

  `

	arguments, err := docopt.Parse(usage, nil, true, "Philea - "+VERSION, false)
	logger.Println(arguments)
	exitWithError(err)

	cmds := getCmds(arguments)
	exclude := getExclude(arguments)
	pattern := getPattern(arguments)
	quiet := isQuiet(arguments)
	serie := isSerie(arguments)
	dry := isDry(arguments)
	changeDir := getWd(arguments)

	if len(changeDir) > 0 {
		err := os.Chdir(changeDir)
		exitWithError(err)
	}

	wd, err := os.Getwd()
	logger.Println("wd=" + wd)
	exitWithError(err)

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
	if dry {
		for _, path := range filteredPaths {
			for _, cmd := range cmds {
				fmt.Println(forgeCmd(path, cmd, wd))
			}
		}
	} else {
		if serie == false {
			var wg sync.WaitGroup
			for _, p := range filteredPaths {
				for _, c := range cmds {
					wg.Add(1)
					go func(path string, cmd string) {
						out, err := executeACommand(path, cmd, wd)
						outs = append(outs, printOut(cmd, string(out), path, wd))
						if err != nil {
							errs = append(errs, err) // racy ?
						}
						wg.Done()
					}(p, c)
				}
			}
			wg.Wait()
		} else {
			for _, path := range filteredPaths {
				for _, cmd := range cmds {
					out, err := executeACommand(path, cmd, wd)
					outs = append(outs, printOut(cmd, string(out), path, wd))
					if err != nil {
						errs = append(errs, err)
					}
				}
			}
		}
	}

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
	cmd = forgeCmd(path, cmd, wd)
	for _, line := range strings.Split(out, "\n") {
		ret = ret + cmd + ": " + line + "\n"
	}
	if ret == "" {
		ret = cmd + ": empty output"
	}
	return ret
}

func executeACommand(path string, cmd string, wd string) ([]byte, error) {
	cmd = forgeCmd(path, cmd, wd)
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return make([]byte, 0), errors.New("Failed to parse cmd: " + cmd)
	}
	bin, err := exec.LookPath(parts[0])
	if err != nil {
		return make([]byte, 0), err
	}
	oCmd := exec.Command(bin, parts[1:]...)
	oCmd.Dir = wd
	oCmd.Env = os.Environ()
	return oCmd.CombinedOutput()
}

func forgeCmd(path string, cmd string, wd string) string {
	logger.Printf("forgeCmd in  cmd='%s' path='%s'", cmd, path)
	dir := filepath.Dir(path)
	path = strings.Replace(path, wd+string(os.PathSeparator), "./", -1)
	path = strings.Replace(path, wd, "./", -1)
	dir = strings.Replace(dir, wd+string(os.PathSeparator), "./", -1)
	dir = strings.Replace(dir, wd, "./", -1)
	if dir == "" {
		dir = "."
	}
	dname := filepath.Base(dir)
	cmd = strings.Replace(cmd, "%dname", dname, -1)
	cmd = strings.Replace(cmd, "%s", path, -1)
	cmd = strings.Replace(cmd, "%d", dir, -1)
	logger.Printf("forgeCmd out cmd='%s' path='%s'", cmd, path)
	return cmd
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
	exclude = strings.Replace(exclude, "**", ".+", -1)
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

func getWd(arguments map[string]interface{}) string {
	dir := ""
	if val, ok := arguments["--change-dir"].(string); ok {
		dir = val
	} else {
		if v, ok := arguments["-C"].(string); ok {
			dir = v
		}
	}
	return dir
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

func isDry(arguments map[string]interface{}) bool {
	dry := false
	if isDry, ok := arguments["--dry"].(bool); ok {
		dry = isDry
	} else {
		if isD, ok := arguments["-d"].(bool); ok {
			dry = isD
		}
	}
	return dry
}

func isSerie(arguments map[string]interface{}) bool {
	serie := false
	if inSerie, ok := arguments["--series"].(bool); ok {
		serie = inSerie
	} else {
		if isS, ok := arguments["-s"].(bool); ok {
			serie = isS
		}
	}
	return serie
}
