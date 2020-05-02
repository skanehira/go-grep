package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
)

var (
	ErrRequireArgs = errors.New("require args")
)

var (
	isRecursion = flag.Bool("R", false, "再帰")
	hasLine     = flag.Bool("l", false, "行番号")
)

func onExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func getFiles(dir string) []string {
	files := []string{}

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		onExit(err)
	}

	for _, f := range entries {
		if f.Name() == ".git" {
			continue
		}
		if f.IsDir() {
			dir := filepath.Join(dir, f.Name())
			files = append(files, getFiles(dir)...)
		} else {
			file := filepath.Join(dir, f.Name())
			files = append(files, file)
		}
	}

	return files
}

func parseArgs() (string, []string) {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		onExit(ErrRequireArgs)
	}

	word := args[0]
	files := []string{}

	if *isRecursion {
		files = getFiles(".")
		return word, files
	}

	if len(args) == 1 {
		entries, err := ioutil.ReadDir(".")
		if err != nil {
			onExit(err)
		}
		for _, f := range entries {
			if !f.IsDir() {
				files = append(files, f.Name())
			}
		}

		return word, files
	}

	if len(args) > 1 {
		files = args[1:]
	}

	return word, files
}

func main() {
	word, files := parseArgs()
	output, err := grep(word, files)
	if err != nil {
		onExit(err)
	}

	for _, out := range output {
		fmt.Println(out)
	}
}

func grep(word string, files []string) ([]string, error) {
	includeWordFiles := []string{}

	for _, f := range files {
		if fi, err := os.Stat(f); err != nil {
			return nil, err
		} else if fi.IsDir() {
			continue
		}

		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Index(line, word) != -1 {
				var out string
				fileName := aurora.Cyan(f).String()
				if *hasLine {
					out = fmt.Sprintf("%s:%s:%s", fileName, aurora.Magenta(lineNum).String(), line)
				} else {
					out = fmt.Sprintf("%s:%s", fileName, line)
				}
				includeWordFiles = append(includeWordFiles, out)
			}
			lineNum++
		}
	}

	return includeWordFiles, nil
}
