// Command ankiexport converts the repo's question banks into an Anki-importable
// TSV deck (anki.tsv).
//
// It walks core-topics/<NN-*>/question-bank/*.md and parses cards in the
// mandatory format:
//
//	### Q: <question, one line>
//	**Frequency**: often | sometimes | rare-senior
//	**Why interviewers ask**: ...
//	**Answer**: ...
//	**Follow-ups**: ...
//
// Output columns: front, back, tags. Tags carry the module number, the
// difficulty file (easy/medium/hard/real), and the frequency value, so the
// deck can be filtered in Anki (e.g. "tag:m07 tag:often").
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var freqRe = regexp.MustCompile(`(?m)^\*\*Frequency\*\*:\s*([a-z-]+)`)

type card struct {
	front string
	back  string
	tags  []string
}

func main() {
	out := flag.String("out", "anki.tsv", "output file")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	var cards []card
	err = filepath.WalkDir(filepath.Join(root, "core-topics"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") ||
			!strings.Contains(path, "question-bank") {
			return nil
		}
		module, difficulty := classify(path)
		cards = append(cards, parseFile(path, module, difficulty)...)
		return nil
	})
	if err != nil {
		fatal(err)
	}

	var b strings.Builder
	b.WriteString("#separator:tab\n#html:true\n#notetype:Basic\n#columns:front,back,tags\n")
	for _, c := range cards {
		fmt.Fprintf(&b, "%s\t%s\t%s\n", field(c.front), field(c.back), strings.Join(c.tags, " "))
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("ankiexport: wrote %d cards to %s\n", len(cards), *out)
	if len(cards) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no cards found — question-bank format drifted?")
		os.Exit(1)
	}
}

// classify derives the module tag (m01..m15) and difficulty tag from the path.
func classify(path string) (module, difficulty string) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if len(p) > 3 && p[2] == '-' && p[0] >= '0' && p[0] <= '9' {
			module = "m" + p[:2]
		}
	}
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	if strings.HasPrefix(base, "real-interview") {
		base = "real"
	}
	return module, base
}

func parseFile(path, module, difficulty string) []card {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")

	var cards []card
	var front string
	var back []string
	flush := func() {
		if front == "" {
			return
		}
		body := strings.TrimSpace(strings.Join(back, "\n"))
		freq := "untagged"
		if m := freqRe.FindStringSubmatch(body); m != nil {
			freq = m[1]
		}
		cards = append(cards, card{
			front: front,
			back:  body,
			tags:  []string{module, difficulty, freq},
		})
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "### Q:") {
			flush()
			front = strings.TrimSpace(strings.TrimPrefix(line, "### Q:"))
			back = back[:0]
			continue
		}
		if front != "" {
			back = append(back, line)
		}
	}
	flush()
	return cards
}

// field makes a string safe for single-line TSV with html:true: tabs are
// flattened and newlines become <br>.
func field(s string) string {
	s = strings.ReplaceAll(s, "\t", "    ")
	return strings.ReplaceAll(s, "\n", "<br>")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ankiexport:", err)
	os.Exit(2)
}
