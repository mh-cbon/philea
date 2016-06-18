# Philea

Apply commands on globbed files

# Install

You can grab a pre-built binary file in the [releases page](https://github.com/mh-cbon/philea/releases)

```sh
mkdir -p $GOPATH/github.com/mh-cbon
cd $GOPATH/github.com/mh-cbon
git clone https://github.com/mh-cbon/philea.git
cd philea
glide install
go install
```

# Usage

```sh
Philea - Apply commands on globbed files

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
  -C --change-dir         Change current working directory.
  -S --series             Execute process in series instead of parallel.
  -d --dry                Show commands only, do not run anything.
  -s --short              Shorter display, wihtout formatting.

Notes:
  cmd can contain
    %s, it will be replaced by the current file path.
    %d, it will be replaced by the path of the directory of the current file.
    %dname, it will be replaced by the directory name of the current file.
    %f, it will be replaced by the name of the current file.
    %fname, it will be replaced by the name of the current file minus its extension.
  philea will process all files and all commands and return an exit code=1 if any fails.

Examples:
  philea "cat %s" "grep t %s"
    It will process all go files, except those in vendors, and apply
    cat, then grep an each file.
  philea "cat %s" "grep t %s" "ls -al %d" "echo '%dname'"
```

# Example

```sh
$ philea --pattern main.go "echo %fname" "echo %f" "echo %dname" "echo %d" "echo %s"
echo ./: ./
echo main.go: main.go
echo main: main
echo ./main.go: ./main.go
echo .: .
```
