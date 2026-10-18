package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyz"
	minLen  = 4
	maxLen  = 10
)

func randomWord() string {
	n := minLen + rand.Intn(maxLen-minLen+1)
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	outputFile := flag.String("out", "data/words.txt", "output file path")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go run generator/main.go -out <file> <number-of-words>")
		os.Exit(1)
	}

	count, err := strconv.Atoi(flag.Arg(0))
	if err != nil || count <= 0 {
		fmt.Fprintln(os.Stderr, "number of words must be a positive integer")
		os.Exit(1)
	}

	f, err := os.Create(*outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create file:", err)
		os.Exit(1)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Keep words unique so the bloom filter test data has no duplicates.
	seen := make(map[string]struct{}, count)
	for len(seen) < count {
		word := randomWord()
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		fmt.Fprintln(w, word)
	}

	fmt.Printf("generated %d words in %s\n", count, *outputFile)
}
