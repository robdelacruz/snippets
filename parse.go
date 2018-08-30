package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Node map[string]string

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

func inSlc(s string, slc []string) bool {
	for _, n := range slc {
		if s == n {
			return true
		}
	}
	return false
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

func init() {
	ts = Tokenize(os.Stdin)

	node = Node{
		"id":     "123",
		"title":  "Note Title",
		"date":   "2018-08-30",
		"cat":    "grocery",
		"tags":   "tag1,tag2,tag3",
		"debit":  "2.00",
		"credit": "3.00",
		"amt":    "99.00",
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

func toBool(s string) bool {
	if s == "0" {
		return false
	}
	return true
}

// D0 <- node[ident]
func field() {
	opr := ts.MatchTok("IDENT")
	fieldName := opr.Val
	fieldType := fieldTypes[fieldName]

	typ := "STR"
	if fieldType == "f" || fieldType == "d" {
		typ = "NUM"
	}
	D0 = Operand{typ, node[fieldName]}
}

// D0 <- num | str | field
// D0 <- (expr1)
func atom() {
	tok := ts.PeekTok()
	if tok == nil {
		expected("field, number or string")
	}
	if tok.Typ == "IDENT" {
		field()
	} else if tok.Typ == "NUM" {
		D0 = ts.MatchTok("NUM")
	} else if tok.Typ == "STR" {
		D0 = ts.MatchTok("STR")
	} else if tok.Typ == "LPAREN" {
		ts.MatchTok("LPAREN")
		expr()
		ts.MatchTok("RPAREN")
	} else {
		expected("field, number or string")
	}
}

// D0 <- atom
// D0 <- atom * atom ...
// D0 <- atom / atom ...
func exprTerm() {
	atom()

	tok := ts.PeekTok()
	for tok != nil {
		if tok.Typ == "MULT" {
			ts.NextTok()
			mult()
		} else if tok.Typ == "DIV" {
			ts.NextTok()
			div()
		} else {
			break
		}
		tok = ts.PeekTok()
	}
}

// D0 <- D0 * atom
func mult() {
	leftD0 := D0
	atom()

	if D0.Typ != "NUM" {
		abort("can't multiply by non-number")
	}

	if leftD0.Typ == "STR" {
		D0.Val = strings.Repeat(leftD0.Val, int(toNum(D0.Val)))
		return
	}
	D0.Typ = "NUM"
	D0.Val = toStr(toNum(leftD0.Val) * toNum(D0.Val))
}

// D0 <- D0 / atom
func div() {
	leftD0 := D0
	atom()

	if D0.Typ != "NUM" {
		abort("can't divide by non-number")
	}
	if leftD0.Typ == "STR" {
		abort("can't divide string")
	}
	if toNum(D0.Val) == 0.0 {
		abort("can't divide by zero")
	}
	D0.Typ = "NUM"
	D0.Val = toStr(toNum(leftD0.Val) / toNum(D0.Val))
}
func add() {
	leftD0 := D0
	exprTerm()

	if leftD0.Typ == "STR" || D0.Typ == "STR" {
		D0.Val = leftD0.Val + D0.Val
		return
	}
	D0.Val = toStr(toNum(leftD0.Val) + toNum(D0.Val))
}
func minus() {
	leftD0 := D0
	exprTerm()

	if leftD0.Typ == "STR" || D0.Typ == "STR" {
		expected("number")
		return
	}
	D0.Val = toStr(toNum(leftD0.Val) - toNum(D0.Val))
}

// D0 <- exprTerm
// D0 <- exprTerm + exprTerm ...
// D0 <- exprTerm - exprTerm ...
func expr() {
	// Check if unary + or -
	unaryMinus := false
	tok := ts.PeekTok()
	if tok.Typ == "PLUS" {
		ts.NextTok()
	} else if tok.Typ == "MINUS" {
		ts.NextTok()
		unaryMinus = true
	}

	exprTerm()

	// If unary minus, negate the D0 value
	if unaryMinus {
		if D0.Typ != "NUM" {
			expected("number after unary minus (-)")
		}
		D0.Val = toStr(0.0 - toNum(D0.Val))
	}

	tok = ts.PeekTok()
	for tok != nil {
		if tok.Typ == "PLUS" {
			ts.NextTok()
			add()
		} else if tok.Typ == "MINUS" {
			ts.NextTok()
			minus()
		} else {
			break
		}
		tok = ts.PeekTok()
	}
}

// D0 <- expr <cmpOp> expr
// Ex.
// title =~ "text within title"
// amt >= 100.0
// debit >= credit - amt + 100.0
// amt / 5 <= credit * 2
func comparison() {
	expr()
	leftD0 := D0

	tok := ts.PeekTok()
	if tok == nil {
		abort("expression needs a condition")
	}
	if !inSlc(tok.Typ, []string{"GT", "GTE", "LT", "LTE", "EQ", "REG_EQ", "NE"}) {
		abort("expression needs a condition")
	}
	cmpOp := tok.Typ
	ts.NextTok()

	expr()
	rightD0 := D0

	D0.Typ = "NUM"
	D0.Val = "0"
	if doCmp(leftD0, cmpOp, rightD0) {
		D0.Val = "1"
	}
}

func doCmp(l Operand, op string, r Operand) bool {
	if l.Typ != r.Typ {
		abort("operand mismatch")
	}

	if l.Typ == "STR" {
		// string comparison
		switch op {
		case "GT":
			return l.Val > r.Val
		case "GTE":
			return l.Val >= r.Val
		case "LT":
			return l.Val < r.Val
		case "LTE":
			return l.Val <= r.Val
		case "EQ":
			return l.Val == r.Val
		case "NE":
			return l.Val != r.Val
		case "REG_EQ":
			matched, _ := regexp.MatchString("(?i)"+r.Val, l.Val)
			return matched
		default:
			abort("unknown comparison operator")
		}
	} else {
		// num comparison
		lNum := toNum(l.Val)
		rNum := toNum(r.Val)

		switch op {
		case "GT":
			return lNum > rNum
		case "GTE":
			return lNum >= rNum
		case "LT":
			return lNum < rNum
		case "LTE":
			return lNum <= rNum
		case "EQ":
			return lNum == rNum
		case "NE":
			return lNum != rNum
		case "REG_EQ":
			matched, _ := regexp.MatchString("(?i)"+r.Val, l.Val)
			return matched
		default:
			abort("unknown comparison operator")
		}
	}

	abort("unknown error")
	return false
}

// D0 <- comparison [and|or comparison]...
func condition() {
	fLParen := false
	tok := ts.PeekTok()
	if tok.Typ == "LPAREN" {
		ts.MatchTok("LPAREN")
		fLParen = true
	}

	comparison()

	tok = ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ts.NextTok()
			conditionAnd()
		} else if tok.Typ == "OR" {
			ts.NextTok()
			conditionOr()
		} else {
			break
		}
		tok = ts.PeekTok()
	}

	if fLParen {
		ts.MatchTok("RPAREN")
	}
}
func conditionAnd() {
	leftD0 := D0
	comparison()

	if toBool(leftD0.Val) && toBool(D0.Val) {
		D0.Val = "1"
	} else {
		D0.Val = "0"
	}
}
func conditionOr() {
	leftD0 := D0
	comparison()

	if toBool(leftD0.Val) || toBool(D0.Val) {
		D0.Val = "1"
	} else {
		D0.Val = "0"
	}
}

func compoundCondition() {
	condition()

	tok := ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ts.NextTok()
			compoundConditionAnd()
		} else if tok.Typ == "OR" {
			ts.NextTok()
			compoundConditionOr()
		} else {
			break
		}
		tok = ts.PeekTok()
	}
}
func compoundConditionAnd() {
	leftD0 := D0
	condition()

	if toBool(leftD0.Val) && toBool(D0.Val) {
		D0.Val = "1"
	} else {
		D0.Val = "0"
	}
}
func compoundConditionOr() {
	leftD0 := D0
	condition()

	if toBool(leftD0.Val) || toBool(D0.Val) {
		D0.Val = "1"
	} else {
		D0.Val = "0"
	}
}

var D0 Operand
var ts *TokStream
var gstack *Stack
var node Node
var fieldTypes map[string]string

func main() {
	compoundCondition()

	fmt.Println(D0.Val)
}
