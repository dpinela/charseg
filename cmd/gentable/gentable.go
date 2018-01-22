package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sort"
	"unicode"
)

func die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error building tables:", err)
		os.Exit(1)
	}
}

func or(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func writeOrDie(out *bufio.Writer, s string) {
	_, err := out.WriteString(s)
	die(err)
}

const prelude = `// generated by charseg/cmd/gentable; DO NOT EDIT

package charseg

type runeRange struct {
	Begin, End rune
}

type category int8

var ranges = []runeRange{
`

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	writeOrDie(out, prelude)
	categories := make(map[string]int8)
	var ranges []runeRange
	for {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		die(err)
		if p := strings.IndexByte(line, '#'); p != -1 {
			line = line[:p]
		}
		if strings.IndexFunc(line, func(r rune) bool { return !unicode.IsSpace(r) }) == -1 {
			continue
		}
		scPos := strings.IndexByte(line, ';')
		if scPos == -1 {
			die(fmt.Errorf("record without a semicolon: %s", line))
		}
		r, err := parseRange(strings.TrimSpace(line[:scPos]))
		die(err)
		category := strings.TrimSpace(line[scPos+1:])
		n, ok := categories[category]
		if !ok {
			n = int8(len(categories) + 1)
			categories[category] = n
		}
		r.Category = n
		ranges = append(ranges, r)
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Begin < ranges[j].Begin })
	for _, r := range ranges {
		_, err := fmt.Fprintf(out, "\t{%#x, %#x},\n", r.Begin, r.End)
		die(err)
	}
	writeOrDie(out, "}\n\nvar categories = []category{")
	for _, r := range ranges {
		_, err := fmt.Fprintf(out, "%d,", r.Category)
		die(err)
	}
	writeOrDie(out, "}\n\nconst (\n\tcatNone category = 0\n")
	for k, v := range categories {
		_, err := fmt.Fprintf(out, "\tcat%s = %d\n", k, v)
		die(err)
	}
	writeOrDie(out, ")\n")
	die(out.Flush())
}

type runeRange struct {
	Begin, End rune
	Category int8
}

// Numeric base used to represent runes in the data files.
const runeBase = 16

const rangeSep = ".."

func parseRange(desc string) (runeRange, error) {
	p := strings.Index(desc, rangeSep)
	if p == -1 {
		n, err := strconv.ParseInt(desc, runeBase, 32)
		return runeRange{Begin: rune(n), End: rune(n) + 1}, err
	}
	m, err := strconv.ParseInt(desc[:p], runeBase, 32)
	n, err2 := strconv.ParseInt(desc[p+len(rangeSep):], runeBase, 32)
	return runeRange{Begin: rune(m), End: rune(n) + 1}, or(err, err2)
}
