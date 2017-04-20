package data

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sort"
)

//================================================================================
var (
	variants      map[rune]Variant
	variantGroups [][]VariantData
)

type VariantData struct {
	char    rune
	variant rune
	typ     string
	face    bool
	tag     string
}

type Variant struct {
	list   []VariantData
	marked bool
}

type variantSorter [][]VariantData
func (a variantSorter) Len() int           { return len(a) }
func (a variantSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a variantSorter) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

//================================================================================
func addVariant (char, variant rune, typ, tag string) {
	frontVariant:= VariantData{
		char:    char,
		variant: variant,
		typ:     typ,
		face:    true,
		tag:     tag,
	}
	backVariant:= VariantData{
		char:    variant,
		variant: char,
		typ:     typ,
		tag:     tag,
	}
	if v, exists:= variants[char]; exists {
		variants[char] = Variant{
			list: append(v.list, frontVariant),
		}
	} else {
		variants[char] = Variant{
			list: []VariantData{frontVariant},
		}
	}
	if v, exists:= variants[variant]; exists {
		variants[variant] = Variant{
			list: append(v.list, backVariant),
		}
	} else {
		variants[variant] = Variant{
			list: []VariantData{backVariant},
		}
	}
}

func sweepVariant (list []VariantData) []VariantData {
	grp:= []VariantData{}
	for _, vd:= range list {
		grp = append(grp, vd)
		v:= variants[vd.variant]
		if ! v.marked {
			variants[vd.variant] = Variant{
				list:   v.list,
				marked: true,
			}
			grp = append(grp, sweepVariant(v.list)...)
		}
	}
	return grp
}

func loadVariants () {
	in, err:= os.Open(unihanDir + "Unihan_Variants.txt")
	if err != nil { fmt.Println("Error in input file open:", err); return }
	defer in.Close()
	text, err:= ioutil.ReadAll(in)
	if err != nil { fmt.Println("Error in file read:", err); return }
	lines:= strings.Split(string(text), "\n")

	variants = map[rune]Variant{}
	for _, line:= range lines {
		if len(line) > 0 && line[0] != '#' {
			fields:= strings.Split(line, "\t")
			field0, err:= strconv.ParseInt(fields[0][2:], 16, 32)
			if err != nil { fmt.Println("Error in parseInt:", err); return }
			fields2Plus:= strings.Split(fields[2], " ")
			for _, varData:= range fields2Plus {
				if varData == "" {continue}
			vars:= strings.Split(varData, "<")
				variant, err:= strconv.ParseInt(vars[0][2:], 16, 32)
				if err != nil { fmt.Println("Error in parseInt:", err); return }
				tag:= ""
				if len(vars) > 1 {tag = vars[1]}
				addVariant(rune(field0), rune(variant), fields[1], tag)
			}
		}
	}

	variantGroups = [][]VariantData{}
	for k, v:= range variants {
		if ! v.marked {
			variants[k] = Variant{
				list:   v.list,
				marked: true,
			}
			grp:= sweepVariant(v.list)
			variants[k] = Variant{
				list:   v.list,
				marked: true,
			}
			variantGroups = append(variantGroups, grp)
		}
	}
	sort.Sort(variantSorter(variantGroups))
}

//================================================================================
func WriteVariants () {
	if variants == nil {
		loadVariants()
	}
	out, err:= os.Create(outputDir + "variantsList.txt")
	if err != nil { fmt.Println("Error in out file open:", err); return }
	defer out.Close()
	out.Sync()
	w:= bufio.NewWriter(out)
	defer w.Flush()

	for _, vGrp:= range variantGroups {
		if len(vGrp) > 0 {
			set:= map[rune]bool{}
			for _, v:= range vGrp {
				set[v.char]= true
			}
			str:= ""
			for k, _:= range set {
				str += fmt.Sprintf("%c(%x)", k, k)
			}
			w.WriteString(fmt.Sprintf("%s: %d... ", str, len(vGrp)/2))
			for _, v:= range vGrp {
				if v.face {
					w.WriteString(fmt.Sprintf("%c/%c: %t, %s, %s; ", v.char, v.variant, v.face, v.typ, v.tag))
				}
			}
			w.WriteString("\n")
		}
	}
}

//================================================================================

