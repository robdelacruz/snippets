package main

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

// Usage:
// ts := Tokenize(os.Stdin)
// tok := ts.peekTok()
// ts.NextTok()
// tok = ts.tok()

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
	"+":  "PLUS",
	"-":  "MINUS",
	"*":  "MULT",
	"/":  "DIV",
}

var _keywords = map[string]string{
	"and": "AND",
	"or":  "OR",
}

type Token struct {
	Typ string
	Lit string
}

type Operand struct {
	Typ string
	Val string
}

type TokStream struct {
	toks     []*Token
	iPeekTok int
}

func Tokenize(f io.Reader) *TokStream {
	r := bufio.NewReader(f)
	toks := []*Token{}
	for {
		tok := readTok(r)
		if tok == nil {
			break
		}
		toks = append(toks, tok)
	}

	return &TokStream{
		toks:     toks,
		iPeekTok: 0,
	}
}

func (ts *TokStream) tok(i int) *Token {
	if i < 0 || i >= len(ts.toks) {
		return nil
	}
	return ts.toks[i]
}

func (ts *TokStream) PeekTok() *Token {
	return ts.tok(ts.iPeekTok)
}

func (ts *TokStream) NextTok() *Token {
	tok := ts.PeekTok()

	ts.iPeekTok++
	if ts.iPeekTok > len(ts.toks)+1 {
		ts.iPeekTok = len(ts.toks) + 1
	}

	return tok
}

func (ts *TokStream) MatchTok(tokName string) Operand {
	tok := ts.PeekTok()
	if tok == nil || tok.Typ != tokName {
		expected(tokName)
	}
	ts.NextTok() // advance read pos

	typ := "STR"
	if tok.Typ == "NUM" {
		typ = "NUM"
	}
	return Operand{typ, tok.Lit}
}

func peekRune(r *bufio.Reader) rune {
	bs := []byte{}
	for {
		peekBs, _ := r.Peek(1)
		if len(peekBs) == 0 {
			return 0
		}
		bs = append(bs, peekBs...)

		if ch, _ := utf8.DecodeRune(bs); ch != utf8.RuneError {
			return ch
		}
	}
	return 0
}

func readRune(r *bufio.Reader) rune {
	ch, _, err := r.ReadRune()
	if err == io.EOF {
		return 0
	}
	return ch
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

func readTok(r *bufio.Reader) *Token {
	skipWhitespace(r)

	ch := readRune(r)
	if ch == 0 {
		return nil
	}

	// Test two letter ops such as ">=", "<>", etc.
	if ch == '>' || ch == '<' || ch == '=' {
		peekCh := peekRune(r)
		if peekCh != 0 {
			opTest := string([]rune{ch, peekCh})
			if _ops[opTest] != "" {
				readRune(r) // advance read pos
				return &Token{_ops[opTest], opTest}
			}
		}
	}

	// Test single letter ops such as ">", "="
	sch := string(ch)
	if _ops[sch] != "" {
		return &Token{_ops[sch], sch}
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

			ch := readRune(r) // advance read pos
			chs = append(chs, ch)
		}
		return &Token{"NUM", string(chs)}
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

			ch = readRune(r) // advance read pos
			chs = append(chs, ch)
		}

		sident := string(chs)
		if _keywords[sident] != "" {
			return &Token{_keywords[sident], sident}
		}
		return &Token{"IDENT", sident}
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
				readRune(r)
				break
			}

			ch := readRune(r) // advance read pos
			chs = append(chs, ch)
		}
		return &Token{"STR", string(chs)}
	}

	return &Token{"UNDEF", string(ch)}
}
