package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Env struct {
	Vars     map[string]string
	VarTypes map[string]string
	D0       Operand
}

type ExprParser struct {
	ts  *TokStream
	D0  Operand
	Env *Env
}

func NewExprParser(f io.Reader) *ExprParser {
	ep := &ExprParser{
		ts: tokenize(f),
	}
	return ep
}

func (ep *ExprParser) Run(env *Env) Operand {
	ep.Env = env
	ep.compoundCondition()
	return ep.Env.D0
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

// D0 <- Vars[ident]
func (ep *ExprParser) field() {
	opr := ep.ts.MatchTok("IDENT")
	varName := opr.Val
	varType := ep.Env.VarTypes[varName]

	typ := "STR"
	if varType == "f" || varType == "d" {
		typ = "NUM"
	}
	ep.Env.D0 = Operand{typ, ep.Env.Vars[varName]}
}

// D0 <- num | str | field
// D0 <- (expr1)
func (ep *ExprParser) atom() {
	tok := ep.ts.PeekTok()
	if tok == nil {
		expected("field, number or string")
	}
	if tok.Typ == "IDENT" {
		ep.field()
	} else if tok.Typ == "NUM" {
		ep.Env.D0 = ep.ts.MatchTok("NUM")
	} else if tok.Typ == "STR" {
		ep.Env.D0 = ep.ts.MatchTok("STR")
	} else if tok.Typ == "LPAREN" {
		ep.ts.MatchTok("LPAREN")
		ep.expr()
		ep.ts.MatchTok("RPAREN")
	} else {
		expected("field, number or string")
	}
}

// D0 <- atom
// D0 <- atom * atom ...
// D0 <- atom / atom ...
func (ep *ExprParser) exprTerm() {
	ep.atom()

	tok := ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "MULT" {
			ep.ts.NextTok()
			ep.mult()
		} else if tok.Typ == "DIV" {
			ep.ts.NextTok()
			ep.div()
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
}

// D0 <- D0 * atom
func (ep *ExprParser) mult() {
	leftD0 := ep.Env.D0
	ep.atom()

	if ep.Env.D0.Typ != "NUM" {
		abort("can't multiply by non-number")
	}

	if leftD0.Typ == "STR" {
		ep.Env.D0.Val = strings.Repeat(leftD0.Val, int(toNum(ep.Env.D0.Val)))
		return
	}
	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) * toNum(ep.Env.D0.Val))
}

// D0 <- D0 / atom
func (ep *ExprParser) div() {
	leftD0 := ep.Env.D0
	ep.atom()

	if ep.Env.D0.Typ != "NUM" {
		abort("can't divide by non-number")
	}
	if leftD0.Typ == "STR" {
		abort("can't divide string")
	}
	if toNum(ep.Env.D0.Val) == 0.0 {
		abort("can't divide by zero")
	}
	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) / toNum(ep.Env.D0.Val))
}
func (ep *ExprParser) add() {
	leftD0 := ep.Env.D0
	ep.exprTerm()

	if leftD0.Typ == "STR" || ep.Env.D0.Typ == "STR" {
		ep.Env.D0.Val = leftD0.Val + ep.Env.D0.Val
		return
	}
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) + toNum(ep.Env.D0.Val))
}
func (ep *ExprParser) minus() {
	leftD0 := ep.Env.D0
	ep.exprTerm()

	if leftD0.Typ == "STR" || ep.Env.D0.Typ == "STR" {
		expected("number")
		return
	}
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) - toNum(ep.Env.D0.Val))
}

// D0 <- exprTerm
// D0 <- exprTerm + exprTerm ...
// D0 <- exprTerm - exprTerm ...
func (ep *ExprParser) expr() {
	// Check if unary + or -
	unaryMinus := false
	tok := ep.ts.PeekTok()
	if tok.Typ == "PLUS" {
		ep.ts.NextTok()
	} else if tok.Typ == "MINUS" {
		ep.ts.NextTok()
		unaryMinus = true
	}

	ep.exprTerm()

	// If unary minus, negate the D0 value
	if unaryMinus {
		if ep.Env.D0.Typ != "NUM" {
			expected("number after unary minus (-)")
		}
		ep.Env.D0.Val = toStr(0.0 - toNum(ep.Env.D0.Val))
	}

	tok = ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "PLUS" {
			ep.ts.NextTok()
			ep.add()
		} else if tok.Typ == "MINUS" {
			ep.ts.NextTok()
			ep.minus()
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
}

// D0 <- expr <cmpOp> expr
// Ex.
// title =~ "text within title"
// amt >= 100.0
// debit >= credit - amt + 100.0
// amt / 5 <= credit * 2
func (ep *ExprParser) comparison() {
	ep.expr()
	leftD0 := ep.Env.D0

	tok := ep.ts.PeekTok()
	if tok == nil {
		abort("expression needs a condition")
	}
	if !inSlc(tok.Typ, []string{"GT", "GTE", "LT", "LTE", "EQ", "REG_EQ", "NE"}) {
		abort("expression needs a condition")
	}
	cmpOp := tok.Typ
	ep.ts.NextTok()

	ep.expr()
	rightD0 := ep.Env.D0

	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = "0"
	if doCmp(leftD0, cmpOp, rightD0) {
		ep.Env.D0.Val = "1"
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
func (ep *ExprParser) condition() {
	fLParen := false
	tok := ep.ts.PeekTok()
	if tok.Typ == "LPAREN" {
		ep.ts.MatchTok("LPAREN")
		fLParen = true

		ep.compoundCondition()

		if fLParen {
			ep.ts.MatchTok("RPAREN")
		}
		return
	}

	ep.comparison()

	tok = ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ep.ts.NextTok()
			ep.conditionAnd()
		} else if tok.Typ == "OR" {
			ep.ts.NextTok()
			ep.conditionOr()
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
}
func (ep *ExprParser) conditionAnd() {
	leftD0 := ep.Env.D0
	ep.comparison()

	if toBool(leftD0.Val) && toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
}
func (ep *ExprParser) conditionOr() {
	leftD0 := ep.Env.D0
	ep.comparison()

	if toBool(leftD0.Val) || toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
}

func (ep *ExprParser) compoundCondition() {
	ep.condition()

	tok := ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ep.ts.NextTok()
			ep.compoundConditionAnd()
		} else if tok.Typ == "OR" {
			ep.ts.NextTok()
			ep.compoundConditionOr()
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
}
func (ep *ExprParser) compoundConditionAnd() {
	leftD0 := ep.Env.D0
	ep.condition()

	if toBool(leftD0.Val) && toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
}
func (ep *ExprParser) compoundConditionOr() {
	leftD0 := ep.Env.D0
	ep.condition()

	if toBool(leftD0.Val) || toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
}

func main() {
	env := Env{}
	env.Vars = map[string]string{
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

	env.VarTypes = map[string]string{
		"debit":  "f",
		"credit": "f",
		"amt":    "f",
		"price":  "f",
		"weight": "f",
		"age":    "d",
	}

	ep := NewExprParser(os.Stdin)
	ep.Run(&env)

	fmt.Println(env.D0.Val)
}
