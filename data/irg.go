package data

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var (
	coreChars map[rune]CoreChar
)

type CoreChar struct {
	char       rune
	priority   rune
	sources    string
	sourceMask byte
}

const (
	PriorityA = 'A'
	PriorityB = 'B'
	PriorityC = 'C'
)

func loadIICore () {
	in, err:= os.Open(unihanDir + "Unihan_IRGSources.txt")
	if err != nil { fmt.Println("Error in input file open:", err); return }
	defer in.Close()
	text, err:= ioutil.ReadAll(in)
	if err != nil { fmt.Println("Error in file read:", err); return }
	lines:= strings.Split(string(text), "\n")
	totals:= map[string]int{}

	coreChars = map[rune]CoreChar{}
	for _, line:= range lines {
		if len(line) > 0 && line[0] != '#' {
			fields:= strings.Split(line, "\t")
			if fields[1] == "kIICore" {
				field0, err:= strconv.ParseInt(fields[0][2:], 16, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				coreChars[rune(field0)] = CoreChar {
					char:     rune(field0),
					priority: rune(fields[2][0]), //only ever 'A', 'B' or 'C'
					sources:  fields[2][1:], //will be one of: GHJKMPT
				}
			} else {
				field1:= fields[1]
				if _, exists:= totals[field1]; ! exists {
					totals[field1] = 1
				} else {
					totals[field1]++
				}
			}
		}
	}
	for k, v:= range totals {
		fmt.Printf("%22s: %6d\n", k, v)
	}
}

func WriteIICore () {
	if coreChars == nil {
		loadIICore()
	}
	out, err:= os.Create(outputDir + "iiCoreData.txt")
	if err != nil { fmt.Println("Error in out file open:", err); return }
	defer out.Close()
	out.Sync()
	w:= bufio.NewWriter(out)
	defer w.Flush()

	for p:= PriorityA; p <= PriorityC; p++ {
		w.WriteString(fmt.Sprintf("Priority %c\n", p))
		for char, record:= range coreChars {
			if record.priority == p {
				w.WriteString(fmt.Sprintf("%c: %s\n", char, record.sources))
			}
		}
		w.WriteString("==========\n")
	}
}

