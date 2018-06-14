package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"github.com/MaxnSter/network_practice/topk"
)

func main() {
	fmt.Println("pid:", os.Getpid())
	wordCount := map[string]int{}

	fr, err := os.Open("input.txt")
	if err != nil {
		panic(err)
	}
	scan := bufio.NewScanner(fr)
	scan.Split(bufio.ScanWords)
	for scan.Scan() {
		wordCount[scan.Text()]++
	}

	var wcs topk.WordCounters
	for word, count := range wordCount {
		wcs = append(wcs, &topk.WordCounter{word, count})
	}

	sort.Sort(wcs)
	f, err := os.Create("output")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, wc := range wcs {
		f.WriteString(fmt.Sprintf("%s:%d\n", wc.Word, wc.Count))
	}
}
