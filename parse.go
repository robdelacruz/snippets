package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Node map[string]string

type Token struct {
	Typ string
	Lit string
}

type Operand struct {
	Typ string
	Val string
}

type Stack struct {
	items []Operand
}

func NewStack() *Stack {
	return &Stack{
		items: []Operand{},
	}
}
func (st *Stack) Push(v Operand) {
	st.items = append(st.items, v)
}
func (st *Stack) Pop() Operand {
	if len(st.items) == 0 {
		panic("Pop() on empty stack")
	}
	nlen := len(st.items)
	v := st.items[nlen-1]
	st.items = st.items[0 : nlen-1]
	return v
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
	"+":  "PLUS",
	"-":  "MINUS",
	"*":  "MULT",
	"/":  "DIV",
}

var _keywords = map[string]string{
	"and": "AND",
	"or":  "OR",
}

func inSlc(s string, slc []string) bool {
	for _, n := range slc {
		if s == n {
			return true
		}
	}
	return false
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

func NextTokenFunc(f io.Reader) func() *Token {
	r := bufio.NewReader(f)

	return func() *Token {
		skipWhitespace(r)

		ch := readRune(r)
		if ch == 0 {
			return nil
		}

		// Test two letter ops such as ">=", "<>", etc.
		if ch == '>' || ch == '<' || ch == '=' {
			peekCh := peekRune(r)
			if peekCh != 0 {
				readRune(r) // advance read pos
				opTest := string([]rune{ch, peekCh})
				if _ops[opTest] != "" {
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
}

func error(s string) {
	fmt.Printf("error(%s)\n", s)
}

func abort(s string) {
	error(s)
	os.Exit(1)
}

func expected(s string) {
	abort(fmt.Sprintf("%s expected", s))
}

func tokenize(f io.Reader, nextTok func() *Token) []*Token {
	toks := []*Token{}
	for {
		tok := nextTok()
		if tok == nil {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

var toks []*Token
var iCurTok int
var iPeekTok int

func init() {
	nextTok := NextTokenFunc(os.Stdin)
	toks = tokenize(os.Stdin, nextTok)
	iCurTok = -1
	iPeekTok = 0

	gstack = NewStack()

	node = Node{
		"id":     "123",
		"title":  "Note Title",
		"tags":   "tag1,tag2,tag3",
		"debit":  "2.00",
		"credit": "3.00",
		"amt":    "5.00",
		"body":   "This is the body content.",
	}

	fieldTypes = map[string]string{
		"debit":  "f",
		"credit": "f",
		"amt":    "f",
		"price":  "f",
		"weight": "f",
		"age":    "d",
	}
}

func isNumField(fieldName string) bool {
	fieldType := fieldTypes[fieldName]
	if fieldType == "f" || fieldType == "d" {
		return true
	}
	return false
}

// Get num from either literal token num or node field
func tokenNum(tok Token) float64 {
	if tok.Typ != "IDENT" && tok.Typ != "NUM" {
		expected("number literal or number field")
	}

	if tok.Typ == "IDENT" {
		fieldName := tok.Lit
		if !isNumField(fieldName) {
			abort(fmt.Sprintf("Field '%s' is not numeric type", fieldName))
		}
		return toNum(node[fieldName])
	}
	return toNum(tok.Lit)
}

// Get string from either literal token str or node field
func tokenStr(tok Token) string {
	if tok.Typ != "IDENT" && tok.Typ != "STR" {
		expected("string literal or field string")
	}

	if tok.Typ == "IDENT" {
		fieldName := tok.Lit
		return node[fieldName]
	}
	return tok.Lit
}

func tok(i int) *Token {
	if i < 0 || i >= len(toks) {
		return nil
	}
	return toks[i]
}

func peekTok() *Token {
	return tok(iPeekTok)
}

func nextTok() *Token {
	tok := peekTok()

	iCurTok++
	iPeekTok++
	if iCurTok > len(toks) {
		iCurTok = len(toks)
	}
	if iPeekTok > len(toks)+1 {
		iPeekTok = len(toks) + 1
	}

	return tok
}

func matchTok(tokName string) Operand {
	tok := peekTok()
	if tok == nil || tok.Typ != tokName {
		expected(tokName)
	}
	nextTok() // advance read pos

	typ := "STR"
	if tok.Typ == "NUM" {
		typ = "NUM"
	}
	return Operand{typ, tok.Lit}
}

func toNum(s string) float64 {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return n
}
func toStr(n float64) string {
	return fmt.Sprintf("%f", n)
}

// D0 <- node[ident]
func field() {
	opr := matchTok("IDENT")
	fieldName := opr.Val
	fieldType := fieldTypes[fieldName]

	typ := "STR"
	if fieldType == "f" || fieldType == "d" {
		typ = "NUM"
	}
	D0 = Operand{typ, node[fieldName]}
}

// D0 <- num | str | field
func atom() {
	tok := peekTok()
	if tok == nil {
		expected("field, number or string")
	}
	if tok.Typ == "IDENT" {
		field()
	} else if tok.Typ == "NUM" {
		D0 = matchTok("NUM")
	} else if tok.Typ == "STR" {
		D0 = matchTok("STR")
	} else {
		expected("field, number or string")
	}
}

// D0 <- atom
// D0 <- atom * atom ...
// D0 <- atom / atom ...
func expr0() {
	atom()

	tok := peekTok()
	for tok != nil {
		if tok.Typ == "MULT" {
			nextTok()
			mult()
		} else if tok.Typ == "DIV" {
			nextTok()
			div()
		} else {
			break
		}
		tok = peekTok()
	}
}

// D0 <- expr0
// D0 <- expr0 + expr0 ...
// D0 <- expr0 - expr0 ...
func expr1() {
	expr0()

	tok := peekTok()
	for tok != nil {
		if tok.Typ == "PLUS" {
			nextTok()
			add()
		} else if tok.Typ == "MINUS" {
			nextTok()
			minus()
		} else {
			break
		}
		tok = peekTok()
	}
}

// D0 <- D0 * field{n}
// D0 <- D0 * num
func mult() {
	tok := peekTok()
	if tok == nil {
		return
	}
	factor := tokenNum(*tok)
	nextTok()

	if D0.Typ == "STR" {
		D0.Val = strings.Repeat(D0.Val, int(factor))
		return
	}
	D0.Typ = "NUM"
	D0.Val = toStr(toNum(D0.Val) * factor)
}

// D0 <- D0 / field{n}
// D0 <- D0 / num
func div() {
	tok := peekTok()
	if tok == nil {
		return
	}
	divisor := tokenNum(*tok)
	nextTok()

	if divisor == 0.0 {
		abort("can't divide by zero")
	}
	if D0.Typ == "STR" {
		abort("can't do division on string")
	}
	D0.Typ = "NUM"
	D0.Val = toStr(toNum(D0.Val) / divisor)
}

func add() {
	leftD0 := D0
	expr0()

	if leftD0.Typ == "STR" || D0.Typ == "STR" {
		D0.Val = leftD0.Val + D0.Val
		return
	}
	D0.Val = toStr(toNum(leftD0.Val) + toNum(D0.Val))
}

func minus() {
	leftD0 := D0
	expr0()

	if leftD0.Typ == "STR" || D0.Typ == "STR" {
		expected("number")
		return
	}
	D0.Val = toStr(toNum(leftD0.Val) - toNum(D0.Val))
}

var D0 Operand
var gstack *Stack
var node Node
var fieldTypes map[string]string

func main() {
	expr1()

	fmt.Println(D0.Val)
}
