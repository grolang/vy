package data

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sort"
	"unicode/utf8"
)

//================================================================================
var (
	ss Syllables
)

type (
	Syllables struct {
		data map[string]Syllable
		seq  []string
	}
	Syllable struct {
		syl       string
		structure SylStruct
		chars     []rune
		others    []rune
	}
)

func (ss *Syllables) AddSyllable (s string, cs ...rune) {
	if v, exists:= ss.data[s]; exists {
		ss.data[s] = Syllable{
			syl:       ss.data[s].syl,
			structure: ss.data[s].structure,
			chars:     append(v.chars, cs...),
			others:    ss.data[s].others,
		}
	} else {
		ss.seq = append(ss.seq, s)
		ss.data[s] = Syllable{
			syl:       s,
			structure: SplitSyllable(s),
			chars:     cs,
			others:    []rune{},
		}
	}
}

func (ss *Syllables) AddOtherCharsToSyllable (s string, cs ...rune) {
	if v, exists:= ss.data[s]; exists {
		ss.data[s] = Syllable{
			syl:       ss.data[s].syl,
			structure: ss.data[s].structure,
			chars:     ss.data[s].chars,
			others:    append(v.others, cs...),
		}
	} else {
		ss.seq = append(ss.seq, s)
		ss.data[s] = Syllable{
			syl:       s,
			structure: SplitSyllable(s),
			chars:     []rune{},
			others:    cs,
		}
	}
}

func (s SylStruct) String () string {
	return fmt.Sprintf("%s,%s,%d", s.initial, s.final, s.tone)
}

type sylSorter []string
func (a sylSorter) Len() int           { return len(a) }
func (a sylSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sylSorter) Less(i, j int) bool {
	initI, initJ:= findInitial(ss.data[a[i]].structure.initial), findInitial(ss.data[a[j]].structure.initial)
	if initI != initJ { return initI > initJ }
	finalI, finalJ:= findFinal(ss.data[a[i]].structure.final), findFinal(ss.data[a[j]].structure.final)
	if finalI != finalJ {return finalI > finalJ }
	return ss.data[a[i]].structure.tone > ss.data[a[j]].structure.tone
}

func loadUnihanReadings() {
	in, err:= os.Open(unihanDir + "Unihan_Readings.txt")
	if err != nil { fmt.Println("Error in input file open:", err); return }
	defer in.Close()
	text, err:= ioutil.ReadAll(in)
	if err != nil { fmt.Println("Error in file read:", err); return }
	lines:= strings.Split(string(text), "\n")

	ss = Syllables {
		seq: []string{},
		data: map[string]Syllable{},
	}
	for _, line:= range lines {
		if len(line) > 0 && line[0] != '#' {
			fields:= strings.Split(line, "\t")
			if fields[1] == "kMandarin" {
				field0, err:= strconv.ParseInt(fields[0][2:], 16, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				fields2Plus:= strings.Split(fields[2], " ")
				ss.AddSyllable(fields2Plus[0], rune(field0))
				if len(fields2Plus) > 1 {
					for _, f:= range fields2Plus[1:] {
						ss.AddOtherCharsToSyllable(f, rune(field0))
					}
				}
			}
		}
	}
	sort.Sort(sylSorter(ss.seq))
}

//================================================================================
var (
	chars Chars
)

type (
	Chars struct {
		data map[rune]Char
	}
	Char struct {
		char  rune
		freq  int
		simpl bool
	}
)

func loadChars () {
	chars = Chars {
		data: map[rune]Char{},
	}
	loadDictionaryLikeData()
	loadIRGSources()
}

func loadDictionaryLikeData () {
	in, err := os.Open(unihanDir + "Unihan_DictionaryLikeData.txt")
	if err != nil { fmt.Println("Error in input file open:", err); return }
	defer in.Close()
	text, err := ioutil.ReadAll(in)
	if err != nil { fmt.Println("Error in file read:", err); return }
	lines:= strings.Split(string(text), "\n")

	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			fields:= strings.Split(line, "\t")
			if fields[1] == "kFrequency" {
				field0, err := strconv.ParseInt(fields[0][2:], 16, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				field2, err := strconv.ParseInt(fields[2:][0], 10, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				chars.data[rune(field0)] = Char{
					char: rune(field0),
					freq: int(field2),
				}
			}
		}
	}
}

func loadIRGSources () {
	in, err := os.Open(unihanDir + "Unihan_IRGSources.txt")
	if err != nil { fmt.Println("Error in input file open:", err); return }
	defer in.Close()
	text, err := ioutil.ReadAll(in)
	if err != nil { fmt.Println("Error in file read:", err); return }
	lines:= strings.Split(string(text), "\n")

	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			fields:= strings.Split(line, "\t")
			if fields[1] == "kIRG_GSource" {
				field0, err := strconv.ParseInt(fields[0][2:], 16, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				if len(fields) > 2 && len(fields[2]) > 3 && fields[2][:3] == "G0-" {
					charData:= chars.data[rune(field0)]
					chars.data[rune(field0)] = Char{
						char: rune(field0),
						freq: charData.freq,
						simpl: true,
					}
				}
			}
		}
	}
}

//================================================================================
type freqSorter []rune
func (a freqSorter) Len() int           { return len(a) }
func (a freqSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a freqSorter) Less(i, j int) bool {
	ai:= chars.data[a[i]]
	if ai.freq == 0 { ai.freq = 6 }
	if ! ai.simpl { ai.freq += 10 }
	aj:= chars.data[a[j]]
	if aj.freq == 0 { aj.freq = 6 }
	if ! aj.simpl { aj.freq += 10 }
	return ai.freq < aj.freq
}

func WriteSyllables () {
	if chars.data == nil {
		if ss.data == nil {
			loadUnihanReadings()
		}
		loadChars()
	}
	out, err:= os.Create(outputDir + "syllableList.txt")
	if err != nil { fmt.Println("Error in out file open:", err); return }
	defer out.Close()
	out.Sync()
	w:= bufio.NewWriter(out)
	defer w.Flush()

	for _, syl:= range ss.seq {
		sort.Sort(freqSorter(ss.data[syl].chars))
		s := ""
		for _, c := range ss.data[syl].chars {
			s = s + string(c)
		}
		s += "/"
		for _, c := range ss.data[syl].others {
			s = s + string(c)
		}
		s += " [" + fmt.Sprint(ss.data[syl].structure) + "]"
		w.WriteString(fmt.Sprintf("%s: %s\n", syl, s))
	}

}

//================================================================================
var vowelData map[rune]VowelData

type (
	SylStruct struct {
		initial string
		final   string
		tone    int
	}
	VowelData struct{
		base rune
		tone int
	}
)

func SplitSyllable (syls string) SylStruct {
	if vowelData == nil {
		initVowelData()
	}
	splitSyls:= strings.Split(syls, " ")
	syl:= splitSyls[0]
	ri, m:= utf8.DecodeRuneInString(syl)
	initial:= string(ri)
	tone:= 0
	rest:= ""
	switch initial {
	case "c", "s", "z":
		rj, n:= utf8.DecodeRuneInString(syl[m:])
		if string(rj) == "h" {
			initial += "h"
			m += n
		}
		rest = syl[m:]
	case "w", "y", "b", "p", "m", "f", "d", "t", "g", "k", "h", "j", "q", "x", "n", "l", "r":
		if initial == "n" && len(syl) == 1 {
			return SylStruct{"n", "-", 5}
		}
		rest = syl[m:]
	case "ḿ":
		return SylStruct{"m", "-", vowelData[ri].tone}
	case "ń", "ň", "ǹ":
		return SylStruct{"n", "-", vowelData[ri].tone}
	default: //vowel
		initial = "-"
		rest = syl
	}

	final:= ""
	tone = 5
	for _, l:= range rest {
		if data, isTonedLetter:= vowelData[l]; isTonedLetter && data.tone != 5 && l != 'n' {
			final += string(data.base)
			tone = data.tone
		} else {
			final += string(l)
		}
	}
	return SylStruct{string(initial), string(final), tone}
}

func initVowelData () {
	vowelData = map[rune]VowelData{}
	vowelBases:= map[rune]string{
		'a': "aāáǎà",
		'e': "eēéěè",
		'i': "iīíǐì",
		'o': "oōóǒò",
		'u': "uūúǔù",
		'v': "üǘǚǜ", //ü with 1st tone not in data
		'-': "ḿnńňǹ",
	}
	for b, ss:= range vowelBases {
		for _, s:= range ss {
			vowelData[s] = VowelData{base: b, tone: 0}
		}
	}

	vowelTones:= map[int]string{
		1: "āēīōū",
		2: "áéíḿńóúǘ",
		3: "ǎěǐňǒǔǚ",
		4: "àèìǹòùǜ",
		5: "aeinouü",
	}
	for t, ss:= range vowelTones {
		for _, s:= range ss {
			vowelData[s] = VowelData{base: vowelData[s].base, tone: t}
		}
	}
}

func findInitial (r string) int {
	initials:= []string{
		"-", "w", "y", "b", "p", "m", "f", "d", "t", "g", "k", "h", "j", "q", "x", "zh", "ch", "sh", "z", "c", "s", "n", "l", "r",
	}
	for x, s := range initials {
		if r == s { return x }
	}
	return len(initials) + 1
}

func findFinal (r string) int {
	finals:= []string{
		"an", "ang", "en", "eng", "ong", "in", "ing", "ian", "iang", "iong", "un", "ung", "uan", "uang", "vn", "van",
		"a", "e", "ai", "ei", "ao", "ou", "r", "i", "ai", "iao", "ie", "iu", "u", "ua", "o", "uo", "uai", "ui", "v", "ve", "-",
	}
	for x, s := range finals {
		if r == s { return x }
	}
	return len(finals) + 1
}

//================================================================================

