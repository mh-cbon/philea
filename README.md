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
  philea [-e <pattern> | --exclude=<pattern>] <cmds>...
  philea [-p <pattern> | --pattern=<pattern>] <cmds>...
  philea -h | --help
  philea -v | --version
Options:

  -h --help             Show this screen.
  -v --version          Show version.
  -e --exclude pattern  Exclude files from being processed [default: *vendors/*].
  -p --pattern pattern  Which kind of files to process [default: **.go].
```

# Examples

```sh
philea "cat %s" "grep t %s"
```
