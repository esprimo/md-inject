package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type config struct {
	ID         string // Identifier for the tags to inject content between
	failOnDiff bool   // If true, exits with code 2 if file would be changed
	printOnly  bool   // If true, prints output to stdout instead of writing to file
	template   string // Go template to apply to stdin before injecting
	filename   string // Target file to modify
}

const (
	tagStartFormat        = "<!-- START md-inject:%s -->"
	tagEndFormat          = "<!-- END md-inject:%s -->"
	defaultOutputTemplate = "{ .stdin }"
	defaultTagID          = "default"
)

func main() {
	flag.Usage = usage
	cfg := parseArgs()

	// read content to be injected
	contentToInject, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// read current content of the file
	oldContent, err := fileContent(cfg.filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// apply a template to the content to be injected
	contentToInject, err = applyTemplate(cfg.template, contentToInject)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error applying template: %v\n", err)
		os.Exit(1)
	}

	// inject new content
	startTag := fmt.Sprintf(tagStartFormat, cfg.ID)
	endTag := fmt.Sprintf(tagEndFormat, cfg.ID)
	updatedContent, err := injectContent(oldContent, contentToInject, startTag, endTag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// nothing is changed, nothing to do
	if oldContent == updatedContent {
		fmt.Printf("No content change needed for %s, nothing to do!\n", cfg.filename)
		os.Exit(0)
	}

	// fail if --fail-on-diff is set
	if cfg.failOnDiff {
		fmt.Fprintf(os.Stderr, "%s would be changed. The file is out of date.\n", cfg.filename)
		os.Exit(2)
	}

	// print to stdout if --print-only is set
	if cfg.printOnly {
		fmt.Print(updatedContent)
		os.Exit(0)
	}

	// update the file
	// This does not change the permissions of the file, it's just a required argument
	// in case the file doesn't exist, but this program fails if it doesn't exist.
	if err := os.WriteFile(cfg.filename, []byte(updatedContent), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", cfg.filename, err)
		os.Exit(1)
	}
	fmt.Printf("%s successfully updated!\n", cfg.filename)
}

func usage() {
	fmt.Fprint(os.Stderr, `Inject text from stdin into markdown files and keep it up to date.

Usage:
  md-inject [OPTIONS] FILE

Examples:
  $ cat foo.txt | md-inject README.md
  $ ./foo --help 2>&1 | md-inject --template='{{ printf "`+"```plaintext\\n%s```"+`" .stdin }}' readme.md
  $ ls -1 | md-inject --fail-on-diff readme.md

Options:
`)
	flag.PrintDefaults()
}

func parseArgs() *config {
	cfg := &config{}

	flag.StringVar(&cfg.ID, "id", defaultTagID, "identifier for the tags to inject content between.")
	flag.BoolVar(&cfg.failOnDiff, "fail-on-diff", false, "set to true to get exit code 2 if the file would be changed")
	flag.BoolVar(&cfg.printOnly, "print-only", false, "print the final output to stdout (this does not write anything to the file)")
	flag.StringVar(&cfg.template, "template", defaultOutputTemplate, "Go template to apply to the stdin before injecting")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "Error: Please specify a target file to inject text into, for example 'README.md'.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	cfg.filename = args[0]

	return cfg
}

func readStdin() (string, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}

	return string(input), nil
}

func fileContent(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func applyTemplate(tmpl, content string) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = t.Execute(&output, map[string]interface{}{
		"stdin": content,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func injectContent(original, addition, startTag, endTag string) (string, error) {
	// look for opening/closing tags
	startTagPos := strings.Index(original, startTag)
	endTagPos := strings.Index(original, endTag)

	// both tags missing - append the content
	if startTagPos < 0 && endTagPos < 0 {
		return fmt.Sprintf("%s\n%s\n%s\n%s\n", original, startTag, addition, endTag), nil
	}

	if startTagPos < 0 {
		return "", fmt.Errorf("missing start tag %s while end tag is present", startTag)
	}
	if endTagPos < 0 {
		return "", fmt.Errorf("missing end tag %s while start tag is present", endTag)
	}
	if startTagPos > startTagPos {
		return "", fmt.Errorf("end tag is before the start tag")
	}

	// both tags are found where they should - inject content
	return fmt.Sprintf("%s\n%s\n%s", original[:startTagPos+len(startTag)], addition, original[endTagPos:]), nil
}
