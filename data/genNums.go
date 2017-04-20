// +build ignore

package main

import(
	"bufio"
	"fmt"
	"os"
)

func main () {
	outputDir:= "E:/Personal/Golang/src/github.com/grolang/vy/data/"
	out, err := os.Create(outputDir + "newNums.txt")
	if err != nil { fmt.Println("Error in opening output file:", err); return }
	defer out.Close()
	out.Sync()
	w := bufio.NewWriter(out)
	defer w.Flush()

	for n:= 0x9fbc; n <= 0x9fd5; n++ {
		w.WriteString(fmt.Sprintf("%c:empty(,,,)\n", n))
	}
}

