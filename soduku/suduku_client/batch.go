package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MaxnSter/network_practice/soduku"
)

func main() {
	f, err := os.OpenFile("./network_practice/soduku/test1000", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	runLocal(f)
}

func runLocal(rd io.Reader) {
	scan := bufio.NewScanner(rd)
	count := 0
	start := time.Now()

	for scan.Scan() {
		soduku.Solve(scan.Text())
		count++
	}

	elapsed := time.Now().Sub(start)
	fmt.Printf("%f seconds, %d total sudoku, %f us per sudoku\n", elapsed.Seconds(),
		count,
		float64(elapsed.Nanoseconds())/float64(count))
}
