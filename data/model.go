// +build ignore

package main

import(
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	inputDir  = "E:/Personal/Golang/src/github.com/grolang/vy/data/"
	outputDir = "E:/Personal/Golang/out/"
)

func main() {
	WriteModelData()
}

//================================================================================
type Decomp interface {
	Char()     string
	Comps()    []string
	String()   string
	Owners()   []Decomp
	OwnersTag() string
	addOwner(Decomp)
	flushOwners()
}

type Core struct {
	char    string
	comment string
	owners  []Decomp
}

func (d Core) Char()   string { return d.char }
func (d Core) Comps()  []string { return []string{} }
func (d Core) String() string { return "c()" + d.comment }
func (d Core) Owners() []Decomp { return d.owners }
func (d *Core) addOwner(o Decomp) { d.owners = append(d.owners, o) }
func (d *Core) flushOwners() { d.owners = []Decomp{} }

func (d Core) OwnersTag() string {
	owners:= []string{}
	for _, owner:= range d.owners {
		c, _:= utf8.DecodeRuneInString(owner.Char())
		if rangeName(c) != "Intermediate" {
			owners = append(owners, owner.Char())
		}
	}
	if len(owners) == 0 {
		for _, owner:= range d.owners {
			owners = append(owners, owner.OwnersTag())
		}
	}
	return "[" + strings.Join(owners, "") + "]"
}

type RepeatAcross struct {
	Core
	comp    string
}
func (d RepeatAcross) Comps() []string { return []string{ d.comp } }
func (d RepeatAcross) String() string { return "ra(" + d.comp + ")" + d.comment }

type BetweenAcross struct {
	Core
	comps   [2]string
}
func (d BetweenAcross) Comps() []string { return []string{ d.comps[0], d.comps[1] } }
func (d BetweenAcross) String() string { return "ba(" + d.comps[0] +"," + d.comps[1] + ")" + d.comment }

type Across struct {
	Core
	comps   []string
}
func (d Across) Comps() []string { return d.comps }
func (d Across) String() string { return "a(" + strings.Join(d.comps, ",") + ")" + d.comment }
/*func (d Across) Expand() string {
	expansion:= ""
	for c:= range d.comps {
		switch theDecomps.data[c] {
		case Across:
		default:
		}
	}
}*/

type Other struct {
	Core
	shape   string
	comps   []string
}
func (d Other) Comps() []string { return d.comps }
func (d Other) String() string { return d.shape + "(" + strings.Join(d.comps, ",") + ")" + d.comment }

//================================================================================
type Decomps struct {
	data map[string]Decomp
	seq  []string
}
func NewDecomps() Decomps {
	return Decomps{data: map[string]Decomp{}, seq:[]string{}}
}

var theDecomps Decomps

//================================================================================
func (ds *Decomps) AddDecomp(char string, shape string, comps []string, comment string) {
	switch shape {
	case "c":
		if len(comps) != 0 {
			fmt.Printf(">>> Wrong number of comps for Core char %s\n", char)
			return
		}
		ds.data[char] = &Core{char:char, comment:comment, owners:[]Decomp{}}
	case "ra":
		if len(comps) != 1 {
			fmt.Printf(">>> Wrong number of comps for RepeatAcross char %s\n", char)
			return
		}
		ds.data[char] = &RepeatAcross{Core:Core{char:char, comment:comment, owners:[]Decomp{}}, comp:comps[0]}
	case "ba":
		if len(comps) != 2 {
			fmt.Printf(">>> Wrong number of comps for BetweenAcross char %s\n", char)
			return
		}
		ds.data[char] = &BetweenAcross{Core:Core{char:char, comment:comment, owners:[]Decomp{}}, comps:[2]string{comps[0], comps[1]}}
	case "a":
		if len(comps) < 2 {
			fmt.Printf(">>> Wrong number of comps for Across char %s\n", char)
			return
		}
		ds.data[char] = &Across{Core:Core{char:char, comment:comment, owners:[]Decomp{}}, comps:comps}
	default:
		ds.data[char] = &Other{Core:Core{char:char, comment:comment, owners:[]Decomp{}}, shape:shape, comps:comps}
	}
	ds.seq = append(ds.seq, char)
}

//================================================================================
func (ds *Decomps) CalcOwners() {
	for _, detail:= range ds.data {
		detail.flushOwners()
	}
	for char, detail:= range ds.data {
		for _, comp:= range detail.Comps() {
			_, exists:= ds.data[comp]
			switch {
			case ! exists:
				fmt.Printf("Didn't exist in decomp map: %s\n", comp)
			case len(comp) != 0:
				ds.data[comp].addOwner(ds.data[char])
			default:
				fmt.Printf("Length comp is zero: %s\n", comp)
			}
		}
	}
	for _, detail:= range ds.data {
		sort.Sort(ownerSorter(detail.Owners()))
	}
}

//================================================================================
type ownerSorter []Decomp
func (a ownerSorter) Len() int           { return len(a) }
func (a ownerSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ownerSorter) Less(i, j int) bool { return stringLessThan(a[i].Char(), a[j].Char()) }

func stringLessThan(si, sj string) bool {
	ri, _:= utf8.DecodeRuneInString(si)
	rj, _:= utf8.DecodeRuneInString(sj)
	var ni, nj int64
	if ri < 0x80 {
		ni, _ = strconv.ParseInt(si, 10, 64)
	}
	if rj < 0x80 {
		nj, _ = strconv.ParseInt(sj, 10, 64)
	}
	if ri < 0x80 && rj < 0x80 { return ni < nj }
	if ri < 0x80 { return false }
	if rj < 0x80 { return true }
	return ri < rj
}

//================================================================================
func buildDecompData() {
	theDecomps = NewDecomps()

	text, err := ioutil.ReadFile(inputDir + "charData.txt")
	if err != nil {
		fmt.Printf("Error in reading file: %s\n", err)
		return
	}

	lines:= strings.Split(string(text), "\n")
	errCount:= 0
	for i, line:= range lines {
		fields:= strings.Split(string(line), ":")
		switch {
		default:
			fmt.Printf("Error in splitting line %d: %s\n", i, fields)
			errCount++
		case len(fields) == 1 && fields[0] == "":
			errCount++
		case len(fields) == 2:
			s:= strings.SplitN(fields[1], "(", 2)
			t:= strings.SplitN(string(s[1]), ")", 2)
			char, shape, comment:= fields[0], s[0], strings.TrimRight(t[1], "\r")
			comps:= strings.Split(string(t[0]), ",")
			if len(comps) == 1 && comps[0] == "" {
				comps = []string{}
			}
			theDecomps.AddDecomp(char, shape, comps, comment)
		}
	}

	if errCount > 1 {
		fmt.Printf("Errors in splitting lines: %d\n", errCount)
	}
	theDecomps.CalcOwners()
}

//================================================================================
func WriteModelData() {
	if theDecomps.data == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "modelData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out) //don't write 0xfeff
	w.WriteString(string(0xfeff))
	defer w.Flush()

	fmt.Printf("size of seq is: %d.\n", len(theDecomps.seq)) //##############
	for _, decomp:= range theDecomps.seq {
		data:= theDecomps.data[decomp]
		w.WriteString(fmt.Sprintf("%s:%s //%s\n", decomp, data.String(), data.OwnersTag()))
	}
}

//================================================================================
func rangeName (r rune) string {
	for k, v:= range unihanRanges {
		if r >= k[0] && r <= k[1] {
			return v
		}
	}
	return ""
}

var unihanRanges = map[[2]rune]string{
	{   0x20,    0x7f}: "Intermediate",  // (begin with digits 0 to 9)
	{ 0x2e80,  0x2ef3}: "Radical",       //115 (2e9a blank)
	{ 0x2f00,  0x2fd5}: "Kangxi",        //TODO: add them
	{ 0x30a0,  0x30ff}: "Katakana",      //60 (various letters only from range)
	{ 0x3105,  0x3129}: "Zhuyin",        //37 //TODO: add 312a to 312d; add 31a0 to 31ba
	{ 0x31c0,  0x31e3}: "Stroke",        //36
	{ 0x3400, 0x2fa1d}: "Unihan",
	{0x100003, 0x1028f4}: "IVD",
}

//================================================================================
