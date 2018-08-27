package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

func main() {
	nextTok := NextTokenFunc(os.Stdin)

	for {
		tok := nextTok()
		if tok.name == "EOF" {
			break
		}
		fmt.Println(tok)
	}
}

type Token struct {
	name string
	lit  string
}

// Sample:
// (date >= "2018-08-01" and date < "2018-09-01") or (cat = 'grocery' or
// cat = 'household') and (amt >= 100.0 or (cat = "dineout" and amt > 75.0))
//
// body =~ "todo"
// cat <> ""

// Token list:
// >, >=, <, <=, =, =~, <>
// string literal, num literal
//

var _ops = map[string]string{
	">":  "GT",
	">=": "GTE",
	"<":  "LT",
	"<=": "LTE",
	"=":  "EQ",
	"=~": "REG_EQ",
	"<>": "NE",
	"(":  "LPAREN",
	")":  "RPAREN",
}

var _keywords = map[string]string{
	"and": "AND",
	"or":  "OR",
}

func peekRune(r *bufio.Reader) rune {
	// Incrementally peek 1 more byte ahead until we get a valid rune.
	nPeek := 0
	for {
		peekBs, _ := r.Peek(nPeek)
		if len(peekBs) < nPeek {
			return 0
		}
		bs := []byte{}
		bs = append(bs, peekBs...)
		if ch, _ := utf8.DecodeRune(bs); ch != utf8.RuneError {
			return ch
		}
		nPeek++
	}
	return 0
}

func skipWhitespace(r *bufio.Reader) {
	for {
		peekCh := peekRune(r)
		// Stop on non whitespace char
		if peekCh == 0 || !unicode.IsSpace(peekCh) {
			break
		}
		r.ReadRune() // advance read pos
	}
}

func NextTokenFunc(f io.Reader) func() Token {
	r := bufio.NewReader(f)

	return func() Token {
		skipWhitespace(r)

		ch, _, err := r.ReadRune()
		if err == io.EOF {
			return Token{"EOF", ""}
		}

		// Test two letter ops such as ">=", "<>", etc.
		if ch == '>' || ch == '<' || ch == '=' {
			peekCh := peekRune(r)
			if peekCh != 0 {
				r.ReadRune() // advance read pos
				opTest := string([]rune{ch, peekCh})
				if _ops[opTest] != "" {
					return Token{_ops[opTest], opTest}
				}
			}
		}

		// Test single letter ops such as ">", "="
		sch := string(ch)
		if _ops[sch] != "" {
			return Token{_ops[sch], sch}
		}

		// Test numbers: 123.45, 5 (both floats and ints supported)
		if unicode.IsDigit(ch) {
			chs := []rune{ch}

			fDecimalPt := false
			for {
				peekCh := peekRune(r)
				if peekCh == 0 {
					break
				}
				// Stop when non-digit or non-decimal point encountered
				if !unicode.IsDigit(peekCh) && peekCh != '.' {
					break
				}
				// Stop if second decimal point encountered  Ex. 123.45.67
				if peekCh == '.' && fDecimalPt {
					break
				}

				if peekCh == '.' {
					fDecimalPt = true
				}

				ch, _, err := r.ReadRune() // advance read pos
				if err == io.EOF {
					break
				}
				chs = append(chs, ch)
			}
			return Token{"NUM", string(chs)}
		}

		// Test identifiers: var1, foo, Foo, camelCasedVar, Field1
		// Or keywords: and, or
		if unicode.IsLetter(ch) {
			chs := []rune{ch}

			for {
				peekCh := peekRune(r)
				if peekCh == 0 {
					break
				}
				// Stop when non-letter, non-digit, '_' encountered.
				if !unicode.IsLetter(peekCh) && !unicode.IsDigit(peekCh) &&
					peekCh != '_' {
					break
				}

				ch, _, err := r.ReadRune() // advance read pos
				if err == io.EOF {
					break
				}
				chs = append(chs, ch)
			}

			sident := string(chs)
			if _keywords[sident] != "" {
				return Token{_keywords[sident], sident}
			}
			return Token{"IDENT", sident}
		}

		if ch == '"' || ch == '\'' {
			quoteChar := ch
			chs := []rune{}

			for {
				peekCh := peekRune(r)
				if peekCh == 0 {
					break
				}
				// Stop when closing quote char encountered.
				if peekCh == quoteChar {
					r.ReadRune()
					break
				}

				ch, _, err := r.ReadRune() // advance read pos
				if err == io.EOF {
					break
				}
				chs = append(chs, ch)
			}
			return Token{"STR_LIT", string(chs)}
		}

		return Token{"UNDEF", string(ch)}
	}
}
