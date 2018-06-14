package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	wordsNum := flag.Int("w", 10000, "words number")
	flag.Parse()

	f, err := os.Create("input.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rand.Seed(time.Now().Unix())
	for i := 0; i < *wordsNum; i++ {
		line := fmt.Sprintf("%020d\n", rand.Intn(*wordsNum/10))
		f.WriteString(line)
	}
}
