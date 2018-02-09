package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/essentialbooks/books/pkg/common"
	"github.com/kjk/u"
)

const (
	showStartLine = "// :show start"
	showEndLine   = "// :show end"
)

func isShowStart(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == showStartLine
}

func isShowEnd(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == showEndLine
}

func countStartChars(s string, c byte) int {
	for i := range s {
		if s[i] != c {
			return i
		}
	}
	return len(s)
}

// remove longest common space/tab prefix on non-empty lines
func shiftLines(lines []string) {
	maxTabPrefix := 1024
	maxSpacePrefix := 1024
	// first determine how much we can remove
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		n := countStartChars(line, ' ')
		if n > 0 {
			if n < maxSpacePrefix {
				maxSpacePrefix = n
			}
			continue
		}
		n = countStartChars(line, '\t')
		if n > 0 {
			if n < maxTabPrefix {
				maxTabPrefix = n
			}
			continue
		}
		// if doesn't start with space or tab, early abort
		return
	}
	if maxSpacePrefix == 1024 && maxTabPrefix == 1024 {
		return
	}

	toRemove := maxSpacePrefix
	if maxTabPrefix != 1024 {
		toRemove = maxTabPrefix
	}
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		lines[i] = line[toRemove:]
	}
}

func extractCodeSnippets(path string) ([]string, error) {
	lines, err := common.ReadFileAsLines(path)
	if err != nil {
		return nil, err
	}
	var res []string
	inShow := false
	for _, line := range lines {
		if isShowStart(line) {
			if inShow {
				return nil, fmt.Errorf("file '%s': consequitive '%s' lines", path, showStartLine)
			}
			inShow = true
			continue
		}
		if isShowEnd(line) {
			if !inShow {
				return nil, fmt.Errorf("file '%s': '%s' without start line", path, showEndLine)
			}
			inShow = false
			continue
		}
		if inShow {
			res = append(res, line)
		}
	}
	shiftLines(res)
	return res, nil
}

func getLangFromFileExt(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".go":
		return "go"
	}
	fmt.Printf("Couldn't deduce language from file name '%s'\n", fileName)
	// TODO: more languages
	return ""
}

func trimEmptyLines(lines []string) []string {
	for len(lines) > 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// TODO: implement 'output' option
func extractCodeSnippetsAsMarkdownLines(baseDir string, line string) ([]string, error) {
	// line is:
	// @file ${fileName} [output]
	addOutput := false
	u.PanicIf(!strings.HasPrefix(line, "@file"))
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid @file line: '%s'", line)
	}
	if parts[0] != "@file" {
		return nil, fmt.Errorf("invalid @file line: '%s'", line)
	}
	fileName := parts[1]
	path := filepath.Join(baseDir, fileName)
	if !fileExists(path) {
		return nil, fmt.Errorf("no file '%s' in line '%s'", path, line)
	}
	rest := parts[2:]
	for len(rest) > 0 {
		s := rest[0]
		rest = rest[1:]
		switch s {
		case "output":
			addOutput = true
		default:
			return nil, fmt.Errorf("unknown option '%s' in '%s'", s, line)
		}
	}
	lines, err := extractCodeSnippets(path)
	if err != nil {
		return nil, err
	}
	lang := getLangFromFileExt(path)
	res := []string{"```" + lang}
	res = append(res, lines...)
	res = append(res, "```")

	if addOutput {
		out, err := getOutput(path)
		if err != nil {
			fmt.Printf("getOutput('%s'): error '%s', output: '%s'\n", path, err, out)
			maybePanicIfErr(err)
		} else {
			res = append(res, "")
			res = append(res, "**Output**:")
			res = append(res, "")
			res = append(res, "```text")
			lines := strings.Split(out, "\n")
			lines = trimEmptyLines(lines)
			res = append(res, lines...)
			res = append(res, "```")
		}
	}
	return res, nil
}

// runs `go run ${path}` and returns captured output`
func getGoOutput(path string) (string, error) {
	cmd := exec.Command("go", "run", path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// it executes a code file and captures the output
func getOutput(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".go" {
		return getGoOutput(path)
	}
	return "", fmt.Errorf("getOutpu(%s): files with extension '%s' are not supported", path, ext)
}