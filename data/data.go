package data

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
	unihanDir = "E:/General/Unicode/v 9.0.0/core/Unihan/"
	UcdDir    = "E:/General/Unicode/v 9.0.0/core/UCD/"
	inputDir  = "E:/Personal/Golang/src/github.com/grolang/vy/data/"
	outputDir = "E:/Personal/Golang/out/"
)

//================================================================================
var(
	decompMap  map[string]DecompData
	decomps    []string
)

type DecompData struct{
	shape   string
	comps   []string
	comment string
	owners  []string
}
type decompSorter []string
func (a decompSorter) Len() int           { return len(a) }
func (a decompSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a decompSorter) Less(i, j int) bool {
	ri, _:= utf8.DecodeRuneInString(a[i])
	rj, _:= utf8.DecodeRuneInString(a[j])
	var ni, nj int64
	if ri < 0x80 {
		ni, _ = strconv.ParseInt(a[i], 10, 64)
	}
	if rj < 0x80 {
		nj, _ = strconv.ParseInt(a[j], 10, 64)
	}
	if ri < 0x80 && rj < 0x80 { return ni < nj }
	if ri < 0x80 { return false }
	if rj < 0x80 { return true }
	return ri < rj
}

func buildDecompData() {
	fmt.Println("running buildDecompData")
	decompMap = map[string]DecompData{}
	decomps = []string{}

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
			decomps = append(decomps, char)
			decompMap[char] = DecompData{shape:shape, comps:comps, comment:comment, owners:[]string{}}
		}
	}

	if errCount > 1 {
		fmt.Printf("Errors in splitting lines: %d\n", errCount)
	}
	addReplacementChars()
	sort.Sort(decompSorter(decomps))
	addOwnersToData()
	fmt.Printf("total chars: %d\n", len(decomps))
	fmt.Printf("num duplicate chars: %d\n", len(decomps)-len(decompMap))
	checkAllIntermHasOwner()
	fmt.Println()
}

func addReplacementChars() {
	text, err := ioutil.ReadFile(inputDir + "replaceChars.txt")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	lines:= strings.Split(string(text), "\n")
	errCount:= 0
	for i, line:= range lines {
		fields:= strings.Split(string(line), ":")
		switch {
		default:
			fmt.Printf("Error in splitting replacement line %d: %s\n", i, fields)
			errCount++
		case len(fields) == 1 && fields[0] == "":
			errCount++
		case len(fields) == 2:
			s:= strings.SplitN(fields[1], "(", 2)
			t:= strings.SplitN(string(s[1]), ")", 2)
			char, shape, comment:= fields[0], s[0], t[1]
			comps:= strings.Split(string(t[0]), ",")
			if len(comps) == 1 && comps[0] == "" {
				comps = []string{}
			}
			_, ok:= decompMap[char]
			if ! ok {
				fmt.Printf("No existing character for replacement char: %s\n", char)
			}
			decompMap[char] = DecompData{shape:shape, comps:comps, comment:comment, owners:[]string{}}
		}
	}
	if errCount > 1 {
		fmt.Printf("Errors in splitting replacement lines: %d\n", errCount)
	}
}

func addOwnersToData() {
	for char, detail:= range decompMap {
		decompMap[char] = DecompData{
			shape:detail.shape,
			comps:detail.comps,
			comment:detail.comment,
			owners:[]string{},
		}
	}
	for char, detail:= range decompMap {
		for _, comp:= range detail.comps {
			_, exists:= decompMap[comp]
			switch {
			case len(comp) != 0:
				o:= append(decompMap[comp].owners, char)
				decompMap[comp] = DecompData{
					shape:decompMap[comp].shape,
					comps:decompMap[comp].comps,
					comment:decompMap[comp].comment,
					owners:o,
				}
			default:
				fmt.Printf("Didn't exist in decomp map: %s\n", comp)
			case ! exists:
			}
		}
	}
	for _, detail:= range decompMap {
		sort.Sort(decompSorter(detail.owners))
	}
}

func checkAllIntermHasOwner() {
	tally:= 0
	for _, decomp:= range decomps {
		decompData:= decompMap[decomp]
		_, sz:= utf8.DecodeRuneInString(decomp)
		if sz < len(decomp) && len(decompData.owners) == 0 {
			delete(decompMap, decomp)
			oldDecomps:= decomps
			decomps = []string{}
			for _, d:= range oldDecomps {
				if d != decomp {
					decomps = append(decomps, d)
				}
			}
			fmt.Printf("Intermediate decomp %s deleted because it doesn't have any owners\n", decomp)
			tally++
		}
	}
	fmt.Printf("number of ownerless intermediates: %d\n", tally)
	if tally > 0 {
		addOwnersToData()
		checkAllIntermHasOwner()
	}
}

//================================================================================
func LookupChar(c rune) (string, []string) {
	if decompMap == nil {
		buildDecompData()
	}
	decomp:= decompMap[string(c)]
	return decomp.shape, decomp.comps
}

//================================================================================
func MergeCharsIntoAnother(t string, fros ...string) {
	if decompMap == nil {
		buildDecompData()
	}
	for _, fro:= range fros {
		for _, owner:= range decompMap[fro].owners {
			for n, comp:= range decompMap[owner].comps {
				if comp == fro {
					decompMap[owner].comps[n] = t
				}
			}
		}
		delete(decompMap, fro)
	}
	oldDecomps:= decomps
	decomps = []string{}
	outer: for _, decomp:= range oldDecomps {
		for _, fro:= range fros {
			if decomp == fro {
				continue outer
			}
		}
		decomps = append(decomps, decomp)
	}
}

//================================================================================
func WriteNewCharData() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "newCharData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out) //don't write 0xfeff
	defer w.Flush()

	for _, decomp:= range decomps {
		decompData:= decompMap[decomp]
		w.WriteString(fmt.Sprintf("%s:%s(%s)%s\n",
			decomp,
			decompData.shape,
			strings.Join(decompData.comps, ","),
			decompData.comment,
		))
	}
}

//================================================================================
var possibleBaseKeys = []string{
	"modified", "overlapping",
	"repeatingHeng", "repeatingNa", "repeatingShu", "repeatingPie", "repeatingOther",
	"touching", "disjoint", "secondLevel",
}

var possibleBases = map[string][]string{
	"modified": []string{"⺄", "乁", "厂", "㇘", "59210", "59668", "99980", "63688"},
	"overlapping": []string{"七", "𠤎", "九", "十", "乂", "乄", "又", "力", "廴", "𠀁", "乜",
		"10013", "60954"},
	"repeatingHeng": []string{"99932", "二", "𠄟", "𠄠", "三", "99876", "亖", "65808"},
	"repeatingNa": []string{"⺀", "99951", "𠁼", "57433", "37546", "99797", "灬"},
	"repeatingShu": []string{"10001", "37372", "川", "57929"},
	"repeatingPie": []string{"37125", "彡"},
	"repeatingOther": []string{"巜", "巛", "99836", "𠃏", "99798", "37843", "69992", "37397", "26128", "𠃐"},
	"touching": []string{"匕", "⺁", "⺆", "⺊", "⻖", "㔾", "丁", "丂", "丄", "丅",
		"丆", "丩", "乃", "了", "亠", "亻", "人", "几", "刀", "勹", "匚", "卜", "卩",
		"厶", "阝", "𠁡", "𠁢", "𠂉", "𠃉", "𠄐", "𠚣", "𠤬", "𢎘",
		"37650", "37143", "37184", "37712", "49005", "62449", "99747", "37681", "37431", "37820", "49306",
		"53055", "55643", "61646", "62101", "99894", "99774", "99783", "99856", "56565", "51840", "99731"},
	"disjoint": []string{"八", "儿", "冫", "刁", "刂", "讠", "𠀂", "𠄍", "𪜊",
		"37473", "37698", "38288", "辶", "48378", "48997", "99907", "10018", "59243", "59409", "60665",
		"64037", "99702", "37476", "32292", "37255", "62211", "99773", "99824", "99847"},
	"secondLevel": []string{"㐅", "纟", "⺈", "⺋", "⺐", "⺙", "⺢", "⺦", "⺭", "⺼", "⺿", "⻎", "⻠",
		"㐄", "㔫", "㔿", "万", "丈", "上", "下", "丌", "不", "丑", "丬", "丯", "丷",
		"丸", "丹", "为", "久", "乆", "乇", "么", "义", "乞", "也", "习", "乡", "亍",
		"亏", "亐", "云", "亓", "井", "以", "兀", "入", "六", "冃", "冄", "内", "円",
		"冖", "凡", "凢", "凵", "凸", "刃", "刄", "勿", "匀", "匚", "匸", "千", "卄",
		"卅", "卌", "卍", "卝", "卬", "卯", "及", "囗", "土", "士", "夂", "大", "夫",
		"女", "子", "孒", "孓", "小", "州", "工", "巾", "干", "幺", "广", "彐", "彑",
		"彳", "心", "忄", "扌", "支", "攴", "文", "斗", "斤", "方", "月", "木", "止",
		"氏", "氵", "父", "犭", "癶", "禸", "见", "贝", "车", "门", "马", "龴", "𠀃",
		"𠀄", "𠀅", "𠀊", "𠁽", "𠂇", "𠂈", "𠂊", "𠂋", "𠂍", "𠂎", "𠂏", "𠂐", "𠂓",
		"𠂖", "𠃒", "𠃓", "𠃘", "𠄏", "𠄑", "𠄡", "𠆢", "𠆥", "𠑶", "𠔀", "𠔂", "𠔇",
		"𠔽", "𠕄", "𠘧", "𠘨", "𠙴", "𠚣", "𠠲", "𠣌", "𠥓", "𠥻", "𠥼", "𠦂", "𠧒",
		"𠫓", "𠫔", "𠬝", "𡕒", "𡤼", "𡭔", "𡭖", "𡯁", "𡯃", "𡿧", "𢌬", "𢍺", "𢖩",
		"𣂑", "𣎳", "𣦶", "𤓯", "𤣥", "𥘅", "𦉪", "𦉫", "𪜀", "𪜁", "𪟽", "𫝀", "𫝄",
		"𫞕", "𫡏", "𫶧", "𬂛",
	},
}

func WritePossibleBases() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "possibleBases.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	coreMap:= map[string]string{
		"stroke": "c", "katakana": "ck", "zhuyin": "cz",
	}
	coreInfo:= map[string][]string{}
	for name, core:= range coreMap {
		currDecomps:= []string{}
		for _, decomp:= range decomps {
			if decompMap[decomp].shape == core {
				currDecomps = append(currDecomps, decomp)
			}
		}
		coreInfo[name] = currDecomps
	}

	w.WriteString(strings.Repeat("=", 40) + "\n")
	w.WriteString(">>> coreComponents:\n")
	for name, _:= range coreMap {
		w.WriteString(fmt.Sprintf("%s: ", name))
		for _, info:= range coreInfo[name] {
			w.WriteString(info)
		}
		w.WriteString("\n")
	}

	for _, key:= range possibleBaseKeys {
		w.WriteString(strings.Repeat("=", 40) + "\n")
		w.WriteString(fmt.Sprintf(">>> %s:\n", key))
		for _, base:= range possibleBases[key] {
			w.WriteString(base + ":" + makeDecompTag(base))
			c, _:= utf8.DecodeRuneInString(base)
			if rangeName(c) == "Intermediate" {
				w.WriteString(makeOwnerTag(base))
			}
			w.WriteString("\n")
		}
	}
	w.WriteString(strings.Repeat("=", 40) + "\n")
}

//================================================================================
func SequenceCharData() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "sequencedData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	var m =  map[string]int{}
	var mm = map[string]map[string]int{}
	for _, decomp:= range decomps {
		r, _:= utf8.DecodeRuneInString(decomp)
		rn:= rangeName(r)
		if rn == "" { fmt.Print("$") }
		if m[rn] == 0 {
			m[rn] = 1
		} else {
			m[rn]++
		}
		comps:= decompMap[decomp].comps
		for _, comp:= range comps {
			c, _:= utf8.DecodeRuneInString(comp)
			rc:= rangeName(c)
			if rc == "" { fmt.Printf(">>> %c (U+%x)\n", c, c) }
			if mm[rn] == nil {
				mm[rn] = map[string]int{rc:1}
			} else if mm[rn][rc] == 0 {
				mm[rn][rc] = 1
			} else {
				mm[rn][rc]++
			}
		}
	}
	for _, k:= range unihanSeq {
		v:= m[k]
		w.WriteString(fmt.Sprintf("%s: %d\n", k, v))
	}
	w.WriteString(fmt.Sprintln())
	for _, k:= range unihanSeq {
		for _, kk:= range unihanSeq {
			v:= mm[k][kk]
			if v != 0 {
				w.WriteString(fmt.Sprintf("%s, %s: %d\n", k, kk, v))
			}
		}
	}
	w.WriteString(fmt.Sprintln())
}

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
	/*
	{ 0x3400,  0x4db5}: "ExtA",          //6582
	{ 0x4e00,  0x9fd5}: "Ideograph",     //20924
	{ 0xfa0e,  0xfa29}: "Compatibility", //12 (fa0e,fa0f,fa11,fa13,fa14,fa1f,fa21,fa23,fa24,fa27,fa28,fa29)
	{0x20000, 0x2a6d6}: "ExtB",          //42711
	{0x2a700, 0x2b734}: "ExtC",          //4149
	{0x2b740, 0x2b81d}: "ExtD",          //222
	{0x2b820, 0x2cea1}: "ExtE",          //5762
	{0x2ceb0, 0x2ebe0}: "ExtF",          //7473 //WAIT: proposed
	{0x2f800, 0x2fa1d}: "CompatSupp",    //TODO: add them?
	*/
	{0x100003, 0x1028f4}: "IVD",
}

var unihanSeq = []string{
	"Stroke",

	"Unihan",
	/*
	"Ideograph",
	"Compatibility",
	"ExtA",
	"ExtB",
	"ExtC",
	"ExtD",
	"ExtE",
	*/

	"Radical",
	"Katakana",
	"Zhuyin",
	"Intermediate",

	"IVD",
}

//================================================================================
func WriteCharSamples() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "charSamples.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	const charsPerLine = 50
	for _, seq:= range unihanSeq {
		if seq == "Intermediate" { continue }
		w.WriteString(seq + "\n")
		w.WriteString(strings.Repeat("-", 100) + "\n")
		var beg, end rune
		for k, v:= range unihanRanges {
			if v == seq { beg = k[0]; end = k[1] }
		}
		lineNo:= 0
		for n:= beg; n <= end; n++ {
			if seq == "Compatibility" {
				compats:= map[rune]interface{}{
					0xfa0e:nil, 0xfa0f:nil, 0xfa11:nil, 0xfa13:nil, 0xfa14:nil, 0xfa1f:nil,
					0xfa21:nil, 0xfa23:nil, 0xfa24:nil, 0xfa27:nil, 0xfa28:nil, 0xfa29:nil,
				}
				_, ok:= compats[n]
				if ! ok { continue }
			}
			w.WriteString(string(n))
			lineNo++
			if lineNo > charsPerLine {
				w.WriteString("\n")
				lineNo = 0
			}
		}
		w.WriteString("\n")
		w.WriteString(strings.Repeat("=", 100) + "\n")
	}
}

//================================================================================
func WriteCharDataDecompTags() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "charDataDecompTags.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, decomp:= range decomps {
		w.WriteString(decomp + ":" + makeDecompTag(decomp))
		c, _:= utf8.DecodeRuneInString(decomp)
		if rangeName(c) == "Intermediate" {
			w.WriteString(makeOwnerTag(decomp))
		}
		w.WriteString("\n")
	}
}

var (
	tagMap map[string][]string
	tags   []string
)

type tagSorter []string
func (a tagSorter) Len() int           { return len(a) }
func (a tagSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tagSorter) Less(i, j int) bool { return len(tagMap[tags[i]]) < len(tagMap[tags[j]]) }

func WriteCharDecompTagSummary() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "charDecompTagSummary.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	tagMap = map[string][]string{}
	tags = []string{}
	for _, decomp:= range decomps {
		tag:= makeDecompTag(decomp)
		//tag:= makeClippedDecompTag(decomp)
		if _, ok:= tagMap[tag]; ! ok {
			tagMap[tag] = []string{}
			tags = append(tags, tag)
		}
		tagMap[tag] = append(tagMap[tag], decomp)
	}

	sort.Sort(tagSorter(tags))
	for _, tag:= range tags {
		w.WriteString(tag + ":" + strings.Join(tagMap[tag], ",") + "\n")
		if len(tagMap[tag]) > 1 {
			w.WriteString(fmt.Sprintf("\tdata.MergeCharsIntoAnother("))
			w.WriteString("\"" + strings.Join(tagMap[tag], "\", \"") + "\")\n")
			for _, char:= range tagMap[tag] {
				w.WriteString(fmt.Sprintf("\t//%s: %s; %s\n", char, makeDecompTag(char), makeOwnerTag(char)))
			}
		}

	}
}

func makeDecompTag(decomp string) string {
	decompData:= decompMap[decomp]
	var compStrs []string = []string{}
	for _, comp:= range decompData.comps {
		r, _:= utf8.DecodeRuneInString(comp)
		if r < 0x80 {
			compStrs = append(compStrs, makeDecompTag(comp))
		} else {
			compStrs = append(compStrs, comp)
		}
	}
	return fmt.Sprintf("%s(%s)", decompData.shape, strings.Join(compStrs, ","))
}

func makeClippedDecompTag(decomp string) string {
	decompData:= decompMap[decomp]
	var compStrs []string = []string{}
	for _, comp:= range decompData.comps {
		r, _:= utf8.DecodeRuneInString(comp)
		if r < 0x80 {
			compStrs = append(compStrs, makeClippedDecompTag(comp))
		} else {
			compStrs = append(compStrs, comp)
		}
	}
	return fmt.Sprintf("%s(%s)", clipShape(decompData.shape), strings.Join(compStrs, ","))
}

//================================================================================
func WriteCharDataExpanded() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "charDataExpanded.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, decomp:= range decomps {
		w.WriteString(makeDecompShape(decomp, true, 99) + "\n")
	}
}

func makeDecompShape(decomp string, base bool, expands int) string {
	decompData:= decompMap[decomp]
	if ( len(decompData.comps) == 0 && ! base ) || expands == 0 {
		return decomp
	}
	var compStrs []string = []string{}
	for _, comp:= range decompData.comps {
		compStrs = append(compStrs, makeDecompShape(comp, false, expands-1))
	}
	return fmt.Sprintf("%s:%s(%s)", decomp, decompData.shape, strings.Join(compStrs, ","))
}

//================================================================================
type stringSorter []string
func (a stringSorter) Len() int           { return len(a) }
func (a stringSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stringSorter) Less(i, j int) bool { return a[i] < a[j] }

var (
	longs   []string
	longMap map[string]string
)

func buildLongs() {
	if babelMap == nil {
		buildBabelData()
	}

	longs = []string{}
	longMap = map[string]string{}

	for _, decomp:= range decomps {
		decompData:= decompMap[decomp]
		if len(decompData.comps) > 2 {
			key:= fmt.Sprintf("%s:%s", decompData.shape, len(decompData.comps))
			if _, ok:= longMap[key]; ! ok {
				longMap[key] = ""
				longs = append(longs, key)
			}
			longMap[key] = longMap[key] + fmt.Sprintf("%s:%s(%s)...%s(%s)\n",
				decomp,
				decompData.shape,
				strings.Join(decompData.comps, ","),
				babelMap[decomp].shape,
				strings.Join(babelMap[decomp].comps, ","),
			)
		}
	}
	sort.Sort(stringSorter(longs))
}

func WriteLongs() {
	if longMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildLongs()
	}

	out, err := os.Create(outputDir + "longs.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, long:= range longs {
		w.WriteString(longMap[long])
	}
}

//================================================================================
var (
	shorts   []string
	shortMap map[string]string
)

func decompsEqual(as, bs []string) bool {
	if len(as) != len(bs) { return false }
	for i, a:= range as {
		if a != bs[i] { return false }
	}
	return true
}

func isBabelString(s string) bool {
	if s == "" { return true }
	r, _:= utf8.DecodeRuneInString(s)
	switch r {
	case '⿰', '⿱', '⿴', '⿵', '⿶', '⿷', '⿸', '⿹', '⿺', '⿻', '⿲', '⿳', '↔', '↷', '{', '？':
		return true
	default:
		return false
	}
}

func buildShorts() {
	if babelMap == nil {
		buildBabelData()
	}

	shorts = []string{}
	shortMap = map[string]string{}

	outer: for _, decomp:= range decomps {
		decompData:= decompMap[decomp]
		if len(decompData.comps) == 2 {
			key:= fmt.Sprintf("%s:%s", decompData.shape, len(decompData.comps))
			if _, ok:= shortMap[key]; ! ok {
				shortMap[key] = ""
				shorts = append(shorts, key)
			}
			switch babelMap[decomp].shape {
				case "", "single", "alt": continue
			}
			for _, s:= range babelMap[decomp].comps {
				if isBabelString(s) { continue outer }
			}
			if decompsEqual(babelMap[decomp].comps, decompMap[decomp].comps) { continue }
			shortMap[key] = shortMap[key] + fmt.Sprintf("%s:%s(%s)...%s(%s)\n",
				decomp,
				decompData.shape,
				strings.Join(decompData.comps, ","),
				babelMap[decomp].shape,
				strings.Join(babelMap[decomp].comps, ","),
			)
		}
	}
	sort.Sort(stringSorter(shorts))
}

func WriteShorts() {
	if shortMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildShorts()
	}

	out, err := os.Create(outputDir + "shorts.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, short:= range shorts {
		w.WriteString(shortMap[short])
	}
}

//================================================================================
func WriteRadicals() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "radicals.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	a, b:= 0x2e80, 0x2ef3 //Radicals
	radicals:= []string{}
	for r:= a; r <= b; r++ {
		if r == 0x2e9a { continue }
		radicals = append(radicals, string(r))
	}

	output:= []string{"", "", "", "", ""}
	for _, radical:= range radicals {
		decompData:= decompMap[radical]
		lenOwners:= len(decompData.owners)
		if lenOwners > 12 { lenOwners = 12 }
		n:= 4
		switch {
		case lenOwners == 0 && decompData.shape == "me":
			n = 0
		case lenOwners == 0 && len(decompData.shape) > 0 && decompData.shape[0] == 'm':
			n = 1
		case lenOwners == 0:
			n = 2
		case lenOwners > 0 && len(decompData.shape) > 0 && decompData.shape[0] == 'm':
			n = 3
		}
		output[n] = output[n] + makeDecompShape(radical, true, 99) + ";owners:(" + strings.Join(decompData.owners[:lenOwners], ",") + ")\n"
	}

	headers:= [5]string{"No owners, me...", "No owners, other m...", "No owners, not m...", "Have owners, various m...", "Have owners, not m..."}
	w.WriteString(strings.Repeat("=", 60) + "\n")
	for i, o:= range output {
		w.WriteString(headers[i] + "\n" + strings.Repeat("-", 40) + "\n" + o + strings.Repeat("=", 60) + "\n")
	}
}

//================================================================================
func WriteCharsWithOwners() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "charsWithOwners.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, decomp:= range decomps {
		decompData:= decompMap[decomp]
		w.WriteString(fmt.Sprintf("%s:%s(%s);owners:(%s)\n",
			decomp,
			decompData.shape,
			strings.Join(decompData.comps, ","),
			strings.Join(decompData.owners, ","),
		))
	}
}

//================================================================================
var(
	intermDecomps []string
)

type intermDecompSorter []string
func (a intermDecompSorter) Len() int           { return len(a) }
func (a intermDecompSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a intermDecompSorter) Less(i, j int) bool {
	oi, oj:= decompMap[a[i]].owners, decompMap[a[j]].owners
	if len(oi) == len(oj) {
		return oi[0] < oj[0]
	} else {
		return  len(oi) < len(oj)
	}
}

func buildIntermDecomps() {
	intermDecomps = []string{}
	for _, char:= range decomps {
		ri, _:= utf8.DecodeRuneInString(char)
		if ri < 0x80 {
			intermDecomps = append(intermDecomps, char)
		}
	}
	sort.Sort(intermDecompSorter(intermDecomps))
}

func WriteIntermCharsWithOwners() {
	if intermDecomps == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildIntermDecomps()
	}
	out, err := os.Create(outputDir + "intermCharsWithOwners.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, decomp:= range intermDecomps {
		decompData:= decompMap[decomp]
		w.WriteString(fmt.Sprintf("%s:%s(%s);owners:(%s)\n",
			decomp,
			decompData.shape,
			strings.Join(decompData.comps, ","),
			strings.Join(decompData.owners, ","),
		))
	}
}

//================================================================================
var(
	shapeMap     map[string]ShapeData
	shapes       []string)

type ShapeData struct{
	chars    []string
}

type shapeSorter []string
func (a shapeSorter) Len() int           { return len(a) }
func (a shapeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a shapeSorter) Less(i, j int) bool { return len(shapeMap[shapes[i]].chars) < len(shapeMap[shapes[j]].chars) }

func buildShapeData() {
	shapeMap = map[string]ShapeData{}
	shapes   = []string{}

	for char, detail:= range decompMap {
		shape:= detail.shape
		if _, exists:= shapeMap[shape]; ! exists {
			shapes = append(shapes, shape)
			shapeMap[shape] = ShapeData{chars:[]string{}}
		}
		o:= append(shapeMap[shape].chars, char)
		shapeMap[shape] = ShapeData{
			chars:o,
		}
	}
	sort.Sort(shapeSorter(shapes))
	fmt.Printf("No. of transformed shape types in data: %d\n", len(shapes))
}

func WriteShapeData() {
	if shapeMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildShapeData()
	}
	out, err := os.Create(outputDir + "shapeData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, shape:= range shapes {
		shapeData:= shapeMap[shape]
		w.WriteString(fmt.Sprintf("%s:\n", shape))
		for _, char:= range shapeData.chars {
			w.WriteString(fmt.Sprintf("\t%s: %s\n", char, decompMap[char].comps))
		}
	}
}

func WriteShapeDataSummary() {
	if shapeMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildShapeData()
	}
	out, err := os.Create(outputDir + "shapeDataSummary.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	shapeDescripMap:= map[string]bool{}
	for _, shapeDescrip:= range shapeDescrips {
		shapeDescripMap[shapeDescrip] = true
	}
	for shape, _:= range shapeMap {
		if ! shapeDescripMap[shape] {
			fmt.Printf("No summary code for shape %s.\n", shape)
		}
	}

	numDuds:= 0
	for _, shape:= range shapeDescrips {
		if shape == "/" {
			w.WriteString("\n")
			numDuds++
			continue
		}
		shapeData, ok:= shapeMap[shape]
		if ! ok {
			fmt.Printf("No shapes match summary code %s.\n", shape)
			continue
		}
		countInterm:= 0
		for _, char:= range shapeData.chars {
			r, _:= utf8.DecodeRuneInString(char)
			if r < 0x80 { countInterm++ }
		}
		lenChars:= len(shapeData.chars)
		//if countInterm != 0 {
		ratio:= 100 * countInterm / lenChars
		//}
		w.WriteString(fmt.Sprintf("%s(%d,%d,%d%%), ", shape, lenChars, countInterm, ratio))
	}
	fmt.Printf("No. of valid shape descriptions: %d\n", len(shapeDescrips) - numDuds)

	w.WriteString(strings.Repeat("=", 40) + "\n")
	for _, shape:= range shapes {
		shapeData:= shapeMap[shape]
		countInterm:= 0
		for _, char:= range shapeData.chars {
			r, _:= utf8.DecodeRuneInString(char)
			if r < 0x80 { countInterm++ }
		}
		w.WriteString(fmt.Sprintf("%s (%d, %d): ", shape, len(shapeData.chars), countInterm))
		charCount:= 0
		for _, char:= range shapeData.chars {
			if charCount > 50 { w.WriteString(fmt.Sprintf("\n...")); charCount = 0 }
			c, _:= utf8.DecodeRuneInString(char)
			if rangeName(c) == "Intermediate" {
				w.WriteString(fmt.Sprintf("%s:%s,", char, makeOwnerTag(char)))
			} else {
				w.WriteString(fmt.Sprintf("%s,", char))
			}
			charCount++
		}
		w.WriteString(fmt.Sprintf("\n"))
	}
}

//c 0: component
//lock 2: components locked together
//m.* 1: modified in some way, e.g. me=equivalent, msp=special, mo=outline, ml=left radical version
//a >=2: flows across
//d >=2: flows downwards
//s.* 2: first component surrounds second, e.g. s=surrounds fully, str=surrounds around the top-right
//w.* 2: second constituent contained within first in some way, e.g. w=within at the center, wbl=within at bottom left
//ba|d 2: second between first moving across or downwards
//r.* 1: repeats and/or reflects in some way, e.g. refh=reflect horizontally, rot=rotate 180 degrees, rrefr= repeat with a reflection rightwards, ra=repeat across, r3d=repeat 3 times downwards, r3tr=repeat in a triangle, rst=repeat surrounding around the top
//The s, a, d, and r codes may be followed by /t touch, /m mold, /s snap, or /o overlap

var shapeDescrips = []string{
	"c", "ck", "cz", "built", "lock", "/",
	"m", "me", "msp", "/",
	"ms", "mt", "ml", "mr", "mb", "mc", "mtl", "mo", "mo2", "mo4", "mo5", "/",
	"a", "a/t", "a/m", "a/s", "/",
	"d", "d/t", "d/m", "d/s", "d/o", "/",
	"s", "s/t", "st", "st/t", "sl", "sl/m", "sl/s", "sb", "sr", "/",
	"stl", "stl/s", "str", "str/o", "sbl", "sbl/m", "sbl/o", "sbr", "/",
	"w", "w/t", "w/o", "wc", "wt", "wt/t", "wb", "wl", "wr", "/",
	"wtl", "wtr", "wbl", "wbr", "/",
	"ba", "ba/o", "bd", "bd/t", "bd/m", "/",
	"refh", "refv", "rot", "/",
	"ra", "ra/m", "rd", "rd/t", "rd/m", "rd/s", "/",
	"rs", "rst", "rsb", "rstl", "rsbr", "/",
	"r3tr", "r3gw", "r3a", "r3d", "r3d/m", "r3st", "r3stl", "r3str", "/",
	"r4sq", "r4a", "r4d", "r4dia", "r4ref/t", "r5x", "r5p", "/",
	"rrefl", "rrefl/m", "rrefr", "rrefr/t", "rrefr/m", "rrefr/o", "rrefd", "rrefd/m", "/",
	"rrotr", "rrotu", "rrotd", "/",
}

/*
a/m:𠨷𦬌𡔝戊𫶧𡗾𠃰
a/t:火𠤬𧰧非⻜𩄩爪⻖𤋅段𠭊𩇦𩇧片⺆𫬯𢎐爲𤕪𡶓俱
ba/o:𢂸(儿,廿)
ra/m: 册𤍍𡦹𠛹𠥡𦁚
r4a:卌灬𠕲
r3a:巛𦟭𡥦川𤩙𠄷䂂𧢛鼠糹州
rrefl:𩇨北卵臦戼𠚒𤕰𢏽𠁁𠂼𩇦亞𡆵𣇅卯𨳈𠁥
	𠀌𢀯𦥑卝𩇧𡘼𣥠𦧄𢍴𫬯𨛜非
rrefl/m:𠃢臼
rrefr:𡴧𣶒𨳇丱𪔂𠁰𡶳𬞂𦣩𦥮𬵼
	𠒅𫸪門八癶丣𨺅𠑹𠒂𦥓鬥
rrefr/t:人𠄷
rrefr/m:𦷣𤔘
rrefr/o:乂
*/

//================================================================================
/*
我:a(123,你)
123:a(他,比)
//我:a(a(他,比),你)
咱:a(他,456)
456:a(比,你)
//我:a(他,a(比,你))
*/
var(
	flattenMap map[string]FlattenData
	flattens   []string
)

type FlattenData struct{
	chars    []string
}

type flattenSorter []string
func (a flattenSorter) Len() int           { return len(a) }
func (a flattenSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a flattenSorter) Less(i, j int) bool { return len(flattenMap[flattens[i]].chars) < len(flattenMap[flattens[j]].chars) }

func buildFlattenData() {
	flattenMap = map[string]FlattenData{}
	flattens   = []string{}

	for char, detail:= range decompMap {
		flatten:= clipShape(detail.shape)
		toplevel:= ""
		for _, comp:= range detail.comps {
			toplevel += comp
			flatten += flattenComp(comp, detail.shape, 0)
		}
		if _, exists:= flattenMap[flatten]; ! exists {
			flattens = append(flattens, flatten)
			flattenMap[flatten] = FlattenData{chars:[]string{}}
		}

		r, _:= utf8.DecodeRuneInString(char)
		if r < 0x80 { char += makeOwnerTag(char) }
		o:= append(flattenMap[flatten].chars, char + "(" + toplevel + ")")
		flattenMap[flatten] = FlattenData{
			chars:o,
		}
	}
	sort.Sort(flattenSorter(flattens))
}

func clipShape(sh string) string {
	return strings.Split(sh, "/")[0]
}

func flattenComp(comp, shape string, recurse int) string {
	//if recurse > 30 { fmt.Printf(">>> CYCLIC: %s\n", comp); return "" }
	result:= ""
	subDecomp:= decompMap[comp]
	if clipShape(subDecomp.shape) == clipShape(shape) {
		for _, subComp:= range subDecomp.comps {
			result += flattenComp(subComp, shape, recurse + 1)
		}
	} else {
		result += comp
	}
	return result
}

func WriteFlattenDataSummary() {
	if flattenMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildFlattenData()
	}
	out, err := os.Create(outputDir + "flattenDataSummary.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, flatten:= range flattens {
		flattenData:= flattenMap[flatten]
		if len(flattenData.chars) <= 1 { continue }
		w.WriteString(fmt.Sprintf("%s (%d): ", flatten, len(flattenData.chars)))
		charCount:= 0
		for _, char:= range flattenData.chars {
			if charCount > 50 { w.WriteString(fmt.Sprintf("\n...")); charCount = 0 }
			w.WriteString(fmt.Sprintf("%s,", char))
			charCount++
		}
		w.WriteString(fmt.Sprintf("\n"))
	}
}

//================================================================================
var(
	leaderMap map[string]LeaderData
	leaders   []string
)

type LeaderData struct{
	chars    []string
}

type leaderSorter []string
func (a leaderSorter) Len() int           { return len(a) }
func (a leaderSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a leaderSorter) Less(i, j int) bool { return len(leaderMap[leaders[i]].chars) < len(leaderMap[leaders[j]].chars) }

func buildLeaderData() {
	leaderMap = map[string]LeaderData{}
	leaders   = []string{}

	for char, detail:= range decompMap {
		var leader string
		if len(detail.comps) > 0 {
			leader = detail.comps[0]
		} else {
			leader = ""
		}

		if _, exists:= leaderMap[leader]; ! exists {
			leaders = append(leaders, leader)
			leaderMap[leader] = LeaderData{chars:[]string{}}
		}
		o:= append(leaderMap[leader].chars, char)
		leaderMap[leader] = LeaderData{
			chars:o,
		}
	}
	sort.Sort(leaderSorter(leaders))
}

func WriteLeaderData() {
	if leaderMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildLeaderData()
	}
	out, err := os.Create(outputDir + "leaderData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, leader:= range leaders {
		leaderData:= leaderMap[leader]
		w.WriteString(fmt.Sprintf("%s:\n", leader))
		for _, char:= range leaderData.chars {
			w.WriteString(fmt.Sprintf("\t%s: %s %s\n", char, decompMap[char].shape, decompMap[char].comps))
		}
	}
}

func WriteLeaderDataSummary() {
	if leaderMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildLeaderData()
	}
	out, err := os.Create(outputDir + "leaderDataSummary.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, leader:= range leaders {
		leaderData:= leaderMap[leader]
		w.WriteString(fmt.Sprintf("%s (%d): ", leader, len(leaderData.chars)))
		charCount:= 0
		for _, char:= range leaderData.chars {
			if charCount > 50 { w.WriteString(fmt.Sprintf("\n...")); charCount = 0 }
			w.WriteString(fmt.Sprintf("%s,", char))
			charCount++
		}
		w.WriteString(fmt.Sprintf("\n"))
	}
}

//================================================================================
var(
	formMap map[string]FormData
	forms   []string
)

type FormData struct{
	chars    []string
}

type formSorter []string
func (a formSorter) Len() int           { return len(a) }
func (a formSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a formSorter) Less(i, j int) bool {
	return len(formMap[forms[i]].chars) < len(formMap[forms[j]].chars)
}

func buildFormData() {
	surroundMap:= map[string]string {
		"w":"s", "wl":"sr", "wt":"sb", "wb":"st", "wr":"sl", "wtl":"sbr", "wtr":"sbl", "wbl":"str", "wbr":"stl",
	}
	clipShape:= func (sh string) string {
		return strings.Split(sh, "/")[0]
	}

	formMap = map[string]FormData{}
	forms   = []string{}

	for char, detail:= range decompMap {
		switch detail.shape {
		case "c", "ck", "cz": continue
		case "": fmt.Printf("Illegal shape for char: %s\n", char)
		}
		if detail.shape[0] == 'w' && len(detail.comps) == 2 {
			detail = DecompData{
				shape:   surroundMap[clipShape(detail.shape)],
				comps:   []string{detail.comps[1], detail.comps[0]},
				comment: detail.comment,
				owners:  detail.owners,
			}
		}
		form:= fmt.Sprintf("%s", append([]string{clipShape(detail.shape)}, detail.comps...))
		if _, exists:= formMap[form]; ! exists {
			forms = append(forms, form)
			formMap[form] = FormData{chars:[]string{}}
		}
		o:= append(formMap[form].chars, char)
		formMap[form] = FormData{
			chars:o,
		}
	}
	sort.Sort(formSorter(forms))
}

func WriteFormData() {
	if formMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildFormData()
	}
	out, err := os.Create(outputDir + "formData.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for _, form:= range forms {
		formData:= formMap[form]
		sort.Sort(decompSorter(formData.chars))
		if len(formData.chars) > 1 {
			w.WriteString(fmt.Sprintf("\tdata.MergeCharsIntoAnother("))
			w.WriteString("\"" + strings.Join(formData.chars, "\", \"") + "\")\n")
			for _, char:= range formData.chars {
				w.WriteString(fmt.Sprintf("\t//%s: %s; %s\n", char, makeDecompTag(char), makeOwnerTag(char)))
			}
		}
	}
}

//================================================================================
func WriteBottomUpView() {
	if decompMap == nil {
		buildDecompData()
	}
	out, err := os.Create(outputDir + "bottomUpView.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	coreComps, coreCompMap:= []string{}, map[string]interface{}{}
	for _, decomp:= range decomps {
		if len(decompMap[decomp].comps) == 0 {
			coreCompMap[decomp] = nil
			coreComps = append(coreComps, decomp)
		}
	}
	prevComps:= coreComps

	w.WriteString(strings.Repeat("=", 40) + "\n")
	w.WriteString(">>> core:\n")
	for _, comp:= range coreComps {
		w.WriteString(fmt.Sprintf("%s:%s\n", comp, makeDecompTag(comp)))
	}

	allComps, allCompMap:= []string{}, map[string]interface{}{}
	for _, coreComp:= range coreComps {
		allComps = append(allComps, coreComp)
		allCompMap[coreComp] = nil
	}

	for i:= 1; i <= 3; i++ {
		theseComps, thisCompMap:= []string{}, map[string]interface{}{}
		for _, comp:= range prevComps {
			eachOwner: for _, compOwner:= range decompMap[comp].owners {
				if _, ok:= thisCompMap[compOwner]; ok { continue eachOwner } //if owner already added, ignore
				for _, ownerComp:= range decompMap[compOwner].comps {
					if _, ok:= allCompMap[ownerComp]; ! ok { continue eachOwner } //if all of owners comps aren't already added, ignore
				}
				theseComps = append(theseComps, compOwner)
				thisCompMap[compOwner] = nil
			}
		}
		sort.Sort(decompSorter(theseComps))

		w.WriteString(strings.Repeat("=", 40) + "\n")
		w.WriteString(fmt.Sprintf(">>> level %d:\n", i))
		for _, comp:= range theseComps {
			w.WriteString(fmt.Sprintf("%s:%s", comp, makeDecompTag(comp)))
			c, _:= utf8.DecodeRuneInString(comp)
			if rangeName(c) == "Intermediate" {
				w.WriteString(makeOwnerTag(comp))
			}
			w.WriteString("\n")
		}

		for _, levelOneComp:= range theseComps {
			allComps = append(allComps, levelOneComp)
			allCompMap[levelOneComp] = nil
		}
		prevComps = theseComps
	}

	w.WriteString(strings.Repeat("=", 40) + "\n")
}

//================================================================================
func makeOwnerTag(comp string) string {
	owners:= []string{}
	for _, owner:= range decompMap[comp].owners {
		c, _:= utf8.DecodeRuneInString(owner)
		if rangeName(c) != "Intermediate" {
			owners = append(owners, owner)
		}
	}
	if len(owners) == 0 {
		for _, owner:= range decompMap[comp].owners {
			owners = append(owners, makeOwnerTag(owner))
		}
	}
	return "[" + strings.Join(owners, "") + "]"
}

//================================================================================
func ParseIdeoDescripSeq(seq string) (shape string, comps []string) {
	descrips:= strings.Split(seq, "\t")
	if len(descrips) > 1 {
		for _, d:= range descrips {
			comps = append(comps, d[1:len(d)-1])
		}
		shape = "alt"
		return
	}
	seq = seq[1:len(seq)-1]
	char, sz:= utf8.DecodeRuneInString(seq)
	switch char {
	case '⿲', '⿳': //3-param ideo-descrip
		for i:= 0; i < 3; i++ {
			oldSz:= sz
			sz = parseIdeoDescrip(seq, sz)
			comps = append(comps, seq[oldSz:sz])
		}
		shape = babelShapes[char]
		return
	case '⿰', '⿱', '⿴', '⿵', '⿶', '⿷', '⿸', '⿹', '⿺', '⿻': //2-param ideo-descrip
		for i:= 0; i < 2; i++ {
			oldSz:= sz
			sz = parseIdeoDescrip(seq, sz)
			comps = append(comps, seq[oldSz:sz])
		}
		shape = babelShapes[char]
		return
	case '↔', '↷': //horiz mirror or 180° rotation op
		oldSz:= sz
		sz = parseIdeoDescrip(seq, sz)
		comps = append(comps, seq[oldSz:sz])
		shape = babelShapes[char]
		return
	case '？': //fullwidth question mark
		shape = "unknown"
		return
	case '{': //0..9 in curly brackets {}
		panic("Intermediate {xx} at top level in Babelstone data.")
	default: //CJK Unified Ideograph, or Stroke or Radical (where there is no Unified Ideograph)
		comps = append(comps, string(char))
		shape = "single"
		return
	}
}

var babelShapes = map[rune]string{
	'⿰':"a", '⿱':"d", '⿴':"s", '⿵':"st", '⿶':"sb", '⿷':"sl", '⿸':"stl", '⿹':"str", '⿺':"sbl", '⿻':"lock",
	'⿲':"a", '⿳':"d",
	'↔':"ref", '↷':"rot",
}

func parseIdeoDescrip(seq string, pos int) int {
	char, n:= utf8.DecodeRuneInString(seq[pos:])
	pos += n
	switch char {
	case '⿲', '⿳':
		for i:= 0; i < 3; i++ {
			pos = parseIdeoDescrip(seq, pos)
		}
	case '⿰', '⿱', '⿴', '⿵', '⿶', '⿷', '⿸', '⿹', '⿺', '⿻':
		for i:= 0; i < 2; i++ {
			pos = parseIdeoDescrip(seq, pos)
		}
	case '↔', '↷':
		pos = parseIdeoDescrip(seq, pos)
	case '{':
		for char:= rune(0); char != '}'; char, n = utf8.DecodeRuneInString(seq[pos:]) {
			pos += n
		}
		pos += n
	}
	return pos
}

type BabelData struct{
	shape   string
	comps   []string
}

var(
	babelMap map[string]BabelData
	babels   []string
)

func buildBabelData() {
	babelMap = map[string]BabelData{}
	babels = []string{}

	text, err := ioutil.ReadFile(inputDir + "Babelstone Unicode 8.0 IDS.txt")
	if err != nil {
		fmt.Printf("Error in reading file: %s", err)
		return
	}

	lines:= strings.Split(string(text), "\r\n")
	for _, line:= range lines {
		if len(line) > 0 {
			r, _:= utf8.DecodeRuneInString(line)
			if r != '#' && r != 0xfeff {
				fields:= strings.SplitN(string(line), "\t", 3)
				babels = append(babels, fields[1])
				babelShape, babelComps:= ParseIdeoDescripSeq(fields[2])
				babelMap[fields[1]] = BabelData{
					shape: babelShape,
					comps: babelComps,
				}
			}
		}
	}
}

func AnalBabelstoneData() {
	if babelMap == nil {
		if decompMap == nil {
			buildDecompData()
		}
		buildBabelData()
	}
	out, err := os.Create(outputDir + "analBabelstone.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	foundMap:= map[string]bool{}
	for _, d:= range decomps {
		decompData, babelData := decompMap[d], babelMap[d]
		switch babelData.shape {
		case "", "alt", "single": continue
		}
		if len(decompData.comps) != len(babelData.comps) { continue }
		var charA, charB string
		for i, comp:= range decompData.comps {
			r, _:= utf8.DecodeRuneInString(comp)
			if r < 0x80 {
				charA = comp
				charB = babelData.comps[i]
				break
			}
		}
		if charA == "" { continue }
		if foundMap[charA] { continue } else {
			foundMap[charA] = true
		}
		r, _:= utf8.DecodeRuneInString(charB)
		if babelShapes[r] != "" { continue }

		w.WriteString(fmt.Sprintf("\tdata.MergeCharsIntoAnother(\"%s\", \"%s\", ) //%s:%s(%s)...%s(%s)\n",
			charB,
			charA,
			d,
			decompData.shape,
			strings.Join(decompData.comps, ","),
			babelData.shape,
			strings.Join(babelData.comps, ","),
		))
	}
}

//================================================================================

