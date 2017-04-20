// +build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	vy "github.com/grolang/vy/data"
)

func main() {
	allReps()
	//merges()
	//gen110000plus()
}

func allReps() {
	vy.WriteNewCharData()

	vy.WriteRadicals()
	vy.SequenceCharData()
	vy.WriteCharsWithOwners()
	vy.WriteIntermCharsWithOwners()
	vy.WriteCharDataExpanded()
	vy.WriteCharDataDecompTags()
	vy.WriteCharDecompTagSummary()
	vy.WriteShapeData()
	vy.WriteShapeDataSummary()
	vy.WriteLeaderData()
	vy.WriteLeaderDataSummary()
	vy.WriteFlattenDataSummary()
	vy.WriteFormData()
	vy.WriteBottomUpView()
	vy.WritePossibleBases()
	vy.WriteLongs()
	vy.WriteShorts()
	vy.AnalBabelstoneData()

	vy.WriteVariants()
	vy.WriteIICore()
	vy.WriteSyllables()
}

func merges() {
	vy.MergeCharsIntoAnother("", "", )
	vy.MergeCharsIntoAnother("", "", )
	vy.MergeCharsIntoAnother("", "", )

	vy.WriteNewCharData()
}

func gen110000plus() {
	out, err := os.Create("scratch.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	w.WriteString(string(0xfeff))
	defer w.Flush()

	for i:= rune(0xf0000); i <= 0xf511a; i++ {
		w.WriteString(fmt.Sprintf("%c\n", i))
	}
}

