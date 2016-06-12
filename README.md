# Philea

Apply commands on globbed files

# Install

```sh
mkdir -p $GOPATH/github.com/mh-cbon
cd $GOPATH/github.com/mh-cbon
cd philea
git clone https://github.com/mh-cbon/philea.git
glide install
go install
```

# Usage

```sh
Philea - Apply commands on globbed files

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
```
