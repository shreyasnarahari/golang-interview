// Command mdcheck compile-checks Go code blocks embedded in Markdown files.
//
// It extracts every ```go fenced code block, keeps only blocks that are
// complete programs (contain "package main" and "func main("), writes each
// to a temporary directory, and compiles it with `go build`. Fragments
// (type declarations, expressions, partial functions) are skipped — they are
// illustrative by nature. Complete programs, however, must compile.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	var failures []string
	checked, skipped := 0, 0

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		blocks := extractGoBlocks(path)
		for i, block := range blocks {
			if !isCompleteProgram(block) {
				skipped++
				continue
			}
			checked++
			if err := compile(block); err != nil {
				rel, _ := filepath.Rel(root, path)
				failures = append(failures, fmt.Sprintf("%s (go block #%d): %v", rel, i+1, err))
			}
		}
		return nil
	})
	if err != nil {
		fatal(err)
	}

	fmt.Printf("mdcheck: %d complete programs compiled, %d fragments skipped\n", checked, skipped)
	if len(failures) > 0 {
		fmt.Println("FAILURES:")
		for _, f := range failures {
			fmt.Println("  -", f)
		}
		os.Exit(1)
	}
}

// extractGoBlocks returns the contents of all ```go fenced code blocks in a
// Markdown file.
func extractGoBlocks(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var blocks []string
	lines := strings.Split(string(data), "\n")
	inBlock := false
	var cur strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock && (trimmed == "```go" || trimmed == "```golang") {
			inBlock = true
			cur.Reset()
			continue
		}
		if inBlock && strings.HasPrefix(trimmed, "```") {
			inBlock = false
			blocks = append(blocks, cur.String())
			continue
		}
		if inBlock {
			cur.WriteString(line)
			cur.WriteString("\n")
		}
	}
	return blocks
}

// isCompleteProgram reports whether a code block is a self-contained,
// compilable Go program rather than an illustrative fragment.
func isCompleteProgram(block string) bool {
	return strings.Contains(block, "package main") && strings.Contains(block, "func main(")
}

func compile(src string) error {
	dir, err := os.MkdirTemp("", "mdcheck-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	mainGo := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainGo, []byte(src), 0o644); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "bin"), mainGo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		snippet := strings.TrimSpace(string(out))
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		return fmt.Errorf("%s", snippet)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mdcheck:", err)
	os.Exit(2)
}
