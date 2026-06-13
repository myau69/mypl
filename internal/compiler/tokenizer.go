package compiler

import (
	"fmt"
	"strings"
	"unicode"
)

type Token struct {
	Text string
	Line int
	Col  int
}

func Tokenize(src string) ([]Token, error) {
	var out []Token
	lines := strings.Split(src, "\n")
	for ln, raw := range lines {
		lineNo := ln + 1
		runes := []rune(raw)
		for col := 0; col < len(runes); {
			r := runes[col]
			if unicode.IsSpace(r) {
				col++
				continue
			}
			if r == '\\' {
				break
			}
			if r == '/' && col+1 < len(runes) && runes[col+1] == '/' {
				break
			}
			start := col
			if r == '"' {
				col++
				for col < len(runes) && runes[col] != '"' {
					if runes[col] == '\\' && col+1 < len(runes) {
						col += 2
						continue
					}
					col++
				}
				if col >= len(runes) || runes[col] != '"' {
					return nil, fmt.Errorf("unterminated string at line %d", lineNo)
				}
				col++
				out = append(out, Token{Text: string(runes[start:col]), Line: lineNo, Col: start + 1})
				continue
			}
			for col < len(runes) {
				r2 := runes[col]
				if unicode.IsSpace(r2) {
					break
				}
				if r2 == '\\' {
					break
				}
				if r2 == '/' && col+1 < len(runes) && runes[col+1] == '/' {
					break
				}
				col++
			}
			out = append(out, Token{Text: string(runes[start:col]), Line: lineNo, Col: start + 1})
		}
	}
	return out, nil
}

func UnquoteForthString(tok string) (string, error) {
	if len(tok) < 2 || tok[0] != '"' || tok[len(tok)-1] != '"' {
		return "", fmt.Errorf("expected quoted string, got%q", tok)
	}
	
	s := tok[1 : len(tok)-1]

	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s, nil
}
