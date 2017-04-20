package data_test

import(
	"github.com/grolang/vy/data"
	"testing"
)


func checkDecomp(t *testing.T, r rune, shape string, comps ...string) {
	sh, cs:= data.LookupChar(r)
	errFn:= func() { t.Errorf(
		"assert failed for %c.\n" +
		"......found: %s %s\n" +
		"...expected: %s %s\n", r, sh, cs, shape, comps) }
	if sh != shape || len(cs) != len(comps) {
		errFn()
		return
	}
	for n, c:= range cs {
		if c != comps[n] {
			errFn()
			return
		}
	}
}

func TestDecompSuccess(t *testing.T){
	checkDecomp(t, '拉', "a", "扌", "立")
	checkDecomp(t, 'A', "")
}

func TestDecompFails(t *testing.T){
	checkDecomp(t, '告', "a", "木", "目")
}

func TestBabelSuccess(t *testing.T){
	checkBabel(t, "⿰氶{27}", true, "a", 2)
	checkBabel(t, "⿱氶一", true, "d", 2)
	checkBabel(t, "両", false, "", 0)
	checkBabel(t, "⿲丨丨⿱⿸𠂉丶𫩏", true, "a", 3)
	checkBabel(t, "⿸⿲丨丨⿱𫩏𠂉丶", true, "stl", 2)
}

func checkBabel(t *testing.T, seq string, compl bool, shape string, length int) {
	cp, sh, le:= data.ParseIdeoDescripSeq(seq)
	errFn:= func() { t.Errorf(
		"assert failed for %s.\n" +
		"......found: %t %s %d\n" +
		"...expected: %t %s %d\n", seq, cp, sh, le, compl, shape, length) }
	if cp != compl || sh != shape || le != length {
		errFn()
		return
	}
}

