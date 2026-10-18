package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/GergesHany/Bloom-Filters/bloomfilter"
)

func readWords(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}
	return words, scanner.Err()
}

func main() {
	// 1. Define flags (returns a pointer to the value)
	expectedUrlsCount := flag.Int("n", 1, "Number of elements")
	FalsePositiveRate := flag.Float64("FP", 0.01, "False positive percentage")
	wordsFilePath := flag.String("words", "data/words.txt", "Path to the words file")
	wordsTestFilePath := flag.String("wordsTest", "data/test_words.txt", "Path to the words test file")

	// 2. Parse the command-line flags
	flag.Parse()

	// 3. Print the values of the flags
	fmt.Println("Number of elements:", *expectedUrlsCount)
	fmt.Println("False positive percentage:", *FalsePositiveRate)
	fmt.Println("Path to the words file:", *wordsFilePath)
	fmt.Println("Path to the words test file:", *wordsTestFilePath)

	bf := BloomFilter.New(uint64(*expectedUrlsCount), *FalsePositiveRate)

	fmt.Println("Bloom Filter Size:", bf.Size())
	fmt.Println("Number of Hash Functions:", bf.HashFuncsCount())

	words, err := readWords(*wordsFilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read words file:", err)
		os.Exit(1)
	}

	for _, word := range words {
		bf.Add(word)
	}

	fmt.Println("Bloom Filter count of elements:", bf.Count())
	fmt.Println("-----------------------------------------------------------")

	// Test with random words taken from words.txt (already added, so they
	// must all be reported as present — bloom filters have no false negatives).
	sampleSize := 20
	if sampleSize > len(words) {
		sampleSize = len(words)
	}
	fmt.Printf("Checking %d random words from %s:\n", sampleSize, *wordsFilePath)
	found, missing := 0, 0
	for _, idx := range rand.Perm(len(words))[:sampleSize] {
		word := words[idx]
		if bf.Contains(word) {
			found++
			fmt.Printf("%q -> exists\n", word)
		} else {
			missing++
			fmt.Printf("%q -> NOT exists (false negative!)\n", word)
		}
	}
	fmt.Println("Found:", found, "| Missing:", missing)
	fmt.Println("-----------------------------------------------------------")

	testWords, err := readWords(*wordsTestFilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read test words file:", err)
		os.Exit(1)
	}

	// test_words.txt is disjoint from words.txt, so any word reported
	// as present is a false positive and false negatives are impossible.
	reportedPresent, reportedAbsent := 0, 0
	for _, word := range testWords {
		if bf.Contains(word) {
			reportedPresent++
		} else {
			reportedAbsent++
		}
	}

	fmt.Println("-----------------------------------------------------------")
	stats := bf.Stats()
	fmt.Println("Test words (not in the dataset):", len(testWords))
	fmt.Println("False positives (reported present):", reportedPresent)
	fmt.Println("Correct rejections (reported absent):", reportedAbsent)
	fmt.Printf("Measured false positive rate: %.6f (%d/%d)\n",
		float64(reportedPresent)/float64(len(testWords)), reportedPresent, len(testWords))
	fmt.Printf("Theoretical false positive rate: %.6f (from n, FP params)\n", stats.ExpectedFPRate)
	fmt.Printf("Density-based estimate: %.6f (bits-set density ^ hash count)\n", stats.ActualFPRate)
}
