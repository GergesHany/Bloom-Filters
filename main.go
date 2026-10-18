package main

import (
	"flag"
	"fmt"
	"github.com/GergesHany/Bloom-Filters/bloomfilter"
)

func init() {
}

func main() {
	// 1. Define flags (returns a pointer to the value)
	expectedUrlsCount := flag.Int("n", 1, "Number of elements")
	FalsePositiveRate := flag.Float64("FP", 0.01, "False positive percentage")
	wordsFilePath := flag.String("words", "words.txt", "Path to the words file")
	wordsTestFilePath := flag.String("wordsTest", "words_test.txt", "Path to the words test file")

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
	fmt.Println("Bloom Filter count of elements:", bf.Count())
	fmt.Println("-----------------------------------------------------------")

	bf.Add("Gerges")

	if bf.Contains("Gerges") {
		fmt.Println("Gerges exists")
	}else {
		fmt.Println("Gerges not exists")
	}
	fmt.Println("Bloom Filter count of elements:", bf.Count())
}
