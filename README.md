# md-inject

Keep markdown files up to date by automatically injecting content into them.

## Usage

<!-- START md-inject:default -->
```console
$ md-inject --help
Inject text from stdin into markdown files and keep it up to date.

Usage:
  md-inject [OPTIONS] FILE

Examples:
  $ cat foo.txt | md-inject README.md
  $ ./foo --help 2>&1 | md-inject --template='{{ printf "```plaintext\n%s```" .stdin }}' readme.md
  $ ls -1 | md-inject --fail-on-diff readme.md

Options:
  -fail-on-diff
    	set to true to get exit code 2 if the file would be changed
  -id string
    	identifier for the tags to inject content between. (default "default")
  -print-only
    	print the final output to stdout (this does not write anything to the file)
  -template string
    	Go template to apply to the stdin before injecting (default "{ .stdin }")
```
<!-- END md-inject:default -->

The code block above is generated with:

```console
$ md-inject --help 2>&1 | md-inject --template=$'```console\n$ md-inject --help\n{{ .stdin }}```' README.md
README.md successfully updated!
```

## Install

Assuming you have Go installed:

```text
go install github.com/esprimo/md-inject@latest
```