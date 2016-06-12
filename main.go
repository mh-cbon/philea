package main

import (
  "os"
  "os/exec"
  "fmt"
  "errors"
  "regexp"
  "strings"
  "path/filepath"
  "sync"

  "github.com/docopt/docopt.go"
  "github.com/mh-cbon/verbose"
)

var logger = verbose.Auto()

func main () {
usage := `Philea - Apply commands on globbed files
Usage:
  philea [options] <cmds>...
  philea [-e <pattern> | --exclude=<pattern>] <cmds>...
  philea [-p <pattern> | --pattern=<pattern>] <cmds>...
  philea -h | --help
  philea -v | --version
Options:
  -h --help             Show this screen.
  -v --version          Show version.
  -e --exclude pattern  Exclude files from being processed.
  -p --pattern pattern  Which kind of files to process.
  `

	arguments, err := docopt.Parse(usage, nil, true, "Philea 0.0.1", false)
  logger.Println(arguments)
  exitWithError(err)

  path, err := os.Getwd()
	logger.Println("path=" + path)
  exitWithError(err)

  cmds := getCmds(arguments)
  exclude := getExclude(arguments)
  pattern := getPattern(arguments)

  if len(cmds)==0 {
    exitWithError(errors.New("There is no commands to execute"))
  }

  excludeRe, err := getExlcudeRe(exclude, false)
  exitWithError(err)
	logger.Println("excludeRe=", excludeRe)
  paths, err := filepath.Glob(path + "/" +pattern)
	logger.Println("paths=", paths)

  if len(paths)==0 {
    exitWithError(errors.New("No matching files found in this directory"))
  }

  filteredPaths := filterPaths(paths, excludeRe)
	logger.Println("filteredPaths=", filteredPaths)

  if len(paths)==0 {
    exitWithError(errors.New("Pattern has excluded all files!"))
  }

  errs := make([]error, 0)
  var wg sync.WaitGroup
  for _, p := range filteredPaths {
    for _, c := range cmds {
      wg.Add(1)
      go func (path string, cmd string) {
        out, err := executeACommand(path, cmd)
        if err == nil {
          printOut(string(out), path)
        } else {
          errs = append(errs, err) // racy ?
        }
        wg.Done()
      }(p, c)
    }
  }
  wg.Wait()

  if len(errs)>0 {
    fmt.Println("-------------")
    fmt.Println("There was errors: ", len(errs))
    for _, err := range errs {
      fmt.Println(err)
    }
    os.Exit(1)
  }
}

func exitWithError (err error) {
  if err!=nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func printOut (out string, path string) {
  for _, line := range strings.Split(out, "\n") {
    fmt.Println(path + ": " + line)
  }
}

func executeACommand(path string, cmd string) ([]byte, error) {
  cmd = strings.Replace(cmd, "%s", path, -1)
  parts := strings.Split(cmd, " ")
  if len(parts)==0 {
    return make([]byte, 0), errors.New("Failed to parse cmd: " + cmd)
  }
  bin := parts[0]
  args := parts[1:]
  oCmd := exec.Command(bin, args...)
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
  if sensitive==false {
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
	exclude := "*vendors/*"
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
	pattern := "**.go"
	if val, ok := arguments["--pattern"].(string); ok {
		pattern = val
	} else {
		if v, ok := arguments["-p"].(string); ok {
			pattern = v
		}
	}
	return pattern
}
