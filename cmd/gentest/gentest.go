package main

import (
	"bufio"
	"os"
	"fmt"
	"io"
	"strings"
	"strconv"
	"unicode/utf8"
)

const prelude = `package charseg

var unicodeTestCases = []testCase{
`

func main() {
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	writeOrDie(out, "%s", prelude)
mainLoop:
	for in.Scan() {
		line := in.Text()
		if i := strings.IndexByte(line, '#'); i != -1 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pieceStrings := strings.Split(line, "÷")
		pieces := make([]string, 0, len(pieceStrings))
		for _, p := range pieceStrings {
			fields := strings.Fields(p)
			if len(fields) == 0 {
				continue
			}
			s := ""
			for _, n := range strings.Fields(p) {
				if n == "×" {
					continue
				}
				r, err := strconv.ParseInt(n, 16, 32)
				if err != nil {
					fmt.Fprintln(os.Stderr, "invalid test:", err)
					continue mainLoop
				}
				// Skip tests involving surrogate halves; charseg is designed to work only on UTF-8 text, and
				// those characters cannot be legally encoded in UTF-8.
				if !utf8.ValidRune(rune(r)) {
					continue mainLoop
				}
				s += string(r)
			}
			pieces = append(pieces, s)
		}
		writeOrDie(out, "\t{in: %q, out: %#v},\n", strings.Join(pieces, ""), pieces)
	}
	writeOrDie(out, "%s", "}\n")
	if err := out.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeOrDie(w io.Writer, s string, xs ...interface{}) {
	if _, err := fmt.Fprintf(w, s, xs...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}