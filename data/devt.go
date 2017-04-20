// +build ignore

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	inputDir  = "E:/Personal/Golang/src/github.com/grolang/vy/data/"
	outputDir = "E:/Personal/Golang/out/"
)

var mapChars = map[string]struct{}{}

func main () {

// >>>>>>>>>>>>>>>>>>>>>>>>>>>> FIX THESE <<<<<<<<<<<<<<<<<<<<<<<<<<<
	assert(ParseDecomp("ab:c()", 51)) ("ab:c(", 51)
	assert(ParseDecomp("ab:c(def,gh())", 51)) (`51:gh(
ab:c(def,51)`, 52)


	assert(ParseDecomp("ab:c(def,gh)", 51)) ("ab:c(def,gh)", 51)
	assert(ParseDecomp("ab:c(,gh)", 51)) ("ab:ERROR: item , is not alphanumeric", 51)
	assert(ParseDecomp("ab:c(def,)", 51)) ("ab:ERROR: comma isn't followed by item", 51)

	assert(ParseDecomp(`𬺚:sbl(d（立，电），m（青）)`, 51)) (`51:d(立,电)
52:m(青)
𬺚:sbl(51,52)`, 53)

	assert(ParseDecomp(`𫣂:a(亻,d(w(乃,又),皿))`, 51)) (`51:w(乃,又)
52:d(51,皿)
𫣂:a(亻,52)`, 53)

	assert(ParseDecomp(`𫣂:a(亻,d(w(乃,),皿))`, 51)) (`51:ERROR: comma isn't followed by item
52:d(51,皿)
𫣂:a(亻,52)`, 53)

	assert(ParseDecomp(`𫣂:a(亻,d(w(乃,又),))`, 51)) (`51:w(乃,又)
52:ERROR: comma isn't followed by item
𫣂:a(亻,52)`, 53)

	analExtEData()
}

//================================================================================
var assertCounter = 0

func assert (pIs string, nIs int) func (string, int) {
	return func (pShouldBe string, nShouldBe int) {
		assertCounter++
		if nIs != nShouldBe {
			fmt.Printf("assert #%d: n is %d but should be %d\n", assertCounter, nIs, nShouldBe)
		} else if pIs != pShouldBe {
			fmt.Printf("assert #%d: p is \"%s\" but should be \"%s\"\n", assertCounter, pIs, pShouldBe)
		}
	}
}

//================================================================================
func analExtEData () {
	text, err := ioutil.ReadFile(inputDir + "extEData.txt")
	if err != nil {
		fmt.Printf("Error in reading file: %s", err)
		return
	}

	out, err := os.Create(outputDir + "newExtEData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	defer w.Flush()

	mapChars = map[string]struct{}{}

	lines:= strings.Split(string(text), "\n")
	n:= 70100
	numLines:= 0
	for _, line:= range lines {
		if len(line) > 0 {
			r, _:= utf8.DecodeRuneInString(line)
			if r != '#' && r != 0xfeff {
				line, n = ParseDecomp(line, n)
				numLines++
			}
			w.WriteString(fmt.Sprintf("%s\n", line))
		}
	}
	fmt.Printf("Num lines: %d\n", numLines)

	fmt.Printf("Num map entries: %d\n", len(mapChars))
	w.WriteString(fmt.Sprintf("================================================================================\n"))
	count:= 0
	for k, _:= range mapChars {
		w.WriteString(fmt.Sprintf("%s", k))
		count++
		if count >= 50 {
			w.WriteString(fmt.Sprintf("\n"))
			count = 0
		}
	}
	w.WriteString(fmt.Sprintf("\n"))
}

//================================================================================
/*
ParseDecomp parses a possibly nested decomposition description for a Unihan character,
and returns a sequence of unnested descriptions using intermediate characters
designated by consecutive integers which start from the integer parameter.
The returned integer is the parameter incremented by the number of intermediate characters.
For example, the parameters
(`𫣂:a(亻,d(w(乃,又),皿))`, 51)
will return
(`51:w(乃,又)
52:d(51,皿)
𫣂:a(亻,52)`, 53)
*/
func ParseDecomp(d string, n int) (string, int) {
	parser:= newParser(d)
	return parser.parseStmt(n)
}

func isAlphanum (s string) bool {
	return s != "(" && s != ")" && s != ":" && s != "," && s != ""
}

func (p *Parser) parseStmt (n int) (string, int) {
	result:= ""
	head:= p.next()
	if ! isAlphanum(head) {
		return "ERROR: head is not alphanumeric", n
	}
	mapChars[head] = struct{}{}
	result += head
	colon:= p.next()
	if colon != ":" {
		return "ERROR: colon is not in second position", n
	}
	result += colon
	expr, extras, newN:= p.parseExpr(p.next(), n)
	result += expr
	return extras + result, newN
}

func (p *Parser) parseExpr (exprHead string, n int) (string, string, int) {
	result, extras, newN:= exprHead, "", n

	openParen:= p.next()
	if openParen != "(" {
		return "ERROR: open parenthesis doesn't follow an expression head", "", newN
	}
	result += openParen

	item:= p.next()
	if item == ")" {
		return result, extras, newN
	} else if ! isAlphanum(item) {
		return fmt.Sprintf("ERROR: item %s is not alphanumeric", item), "", newN
	}
	for {
		switch p.peek() {
		case "(":
			newItem, moreExtras:= "", ""
			newItem, moreExtras, newN = p.parseExpr(item, newN) //lookahead saw a nested expression
			extras += moreExtras
			item = fmt.Sprint(newN)
			extras += fmt.Sprint(newN) + ":" + newItem + "\n"
			newN++
			continue
		case ",":
			p.next()
			result += item + ","
			item = p.next()
			if item == ")" {
				return "ERROR: comma isn't followed by item", extras, newN
			} else {
				continue
			}
		case ")":
			p.next()
			result += item + ")"
			return result, extras, newN
		case "":
			return result, extras, newN
		default:
			return "ERROR: item isn't followed by open paren, comma, or close paren", extras, newN
		}
	}
}

type Parser struct {
	lex Lexer
	buf string
}

func newParser (d string) Parser {
	l:= Lexer {
		str: d,
		pos: 0,
	}
	return Parser {
		lex: l,
		buf: l.scanNext(),
	}
}

func (p *Parser) next() string {
	ret:= p.buf
	p.buf = p.lex.scanNext()
	return ret
}

func (p *Parser) peek() string {
	return p.buf
}

type Lexer struct {
	str string
	pos int
}

func (l *Lexer) scanNext() string {
	if len(l.str) <= l.pos {
		return ""
	}
	oldPos:= l.pos
	ru, le:= utf8.DecodeRuneInString(l.str[l.pos:])
	switch ru {
	case '(', ')', ':', ',':
		l.pos++
		return string(l.str[oldPos])
	case '（':
		l.pos += le
		return "("
	case '）':
		l.pos += le
		return ")"
	case '，':
		l.pos += le
		return ","
	case '：':
		l.pos += le
		return ":"
	default:
		n:= 0
		for _, r:= range l.str[l.pos:] {
			lenR:= utf8.RuneLen(r)
			if r == '(' || r == ')' || r == ':' || r == ',' || r == '（' || r == '）' || r == '：' || r == '，' {
				l.pos += n
				return l.str[oldPos:l.pos]
			}
			n += lenR
		}
		l.pos = len(l.str) //TODO: fix this
		return l.str[oldPos:]
	}
}

//================================================================================

