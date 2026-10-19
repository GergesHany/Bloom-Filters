package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	db "github.com/GergesHany/Bloom-Filters/DB"
	"github.com/GergesHany/Bloom-Filters/bloomfilter"

	"database/sql"

	_ "github.com/lib/pq"
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

func importDatasetToBloomFilter(bf *BloomFilter.BloomFilter, words []string) error {
	for _, word := range words {
		bf.Add(word)
	}
	return nil
}

func importDatasetToPostgres(datasetPath string, db *sql.DB, words []string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO words (word) VALUES ($1) ON CONFLICT DO NOTHING")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, word := range words {
		_, err := stmt.Exec(word)
		if err != nil {
			return fmt.Errorf("failed to insert word: %v", err)
		}
	}

	return tx.Commit()
}

func benchmarkLookups(bf *BloomFilter.BloomFilter, testWords []string, runs int) (int, time.Duration) {
	hits := 0
	totalLookups := 0
	start := time.Now()
	for i := 0; i < runs; i++ {
		for _, word := range testWords {
			if bf.Contains(word) {
				hits++
			}
			totalLookups++
		}
	}
	duration := time.Since(start)
	return hits, duration
}

func benchmarkPostgresLookups(db *sql.DB, testWords []string, runs int) (int, time.Duration, error) {
	hits := 0
	totalLookups := 0
	start := time.Now()
	for i := 0; i < runs; i++ {
		for _, word := range testWords {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM words WHERE word = $1)", word).Scan(&exists)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to query word: %v", err)
			}
			if exists {
				hits++
			}
			totalLookups++
		}
	}
	duration := time.Since(start)
	return hits, duration, nil
}

func main() {
	expectedElements := flag.Int("n", 1000, "Number of elements")
	falsePositiveRate := flag.Float64("FP", 0.01, "False positive percentage")
	wordsFilePath := flag.String("words", "data/words.txt", "Path to the words file")
	wordsTestFilePath := flag.String("wordsTest", "data/test_words.txt", "Path to the words test file")
	runs := flag.Int("runs", 100, "Number of lookup benchmark repetitions over the test set")

	flag.Parse()

	bf := BloomFilter.New(uint64(*expectedElements), *falsePositiveRate)

	// Read the dataset once and load it into both stores
	words, err := readWords(*wordsFilePath)
	if err != nil {
		fmt.Printf("Failed to read words file: %v\n", err)
		return
	}

	if err := importDatasetToBloomFilter(bf, words); err != nil {
		fmt.Printf("Failed to import dataset to Bloom Filter: %v\n", err)
		return
	}

	// Connect to PostgreSQL
	database, err := db.InitDB()
	if err != nil {
		fmt.Printf("Failed to connect to PostgreSQL: %v\n", err)
		return
	}
	defer database.Close()

	// Import dataset to PostgreSQL
	if err := importDatasetToPostgres(*wordsFilePath, database, words); err != nil {
		fmt.Printf("Failed to import dataset to PostgreSQL: %v\n", err)
		return
	}

	// ----------------------------------------------------------------------------------------------------------------

	// Benchmark lookups in Bloom Filter
	testWords, err := readWords(*wordsTestFilePath)
	if err != nil {
		fmt.Printf("Failed to read test words file: %v\n", err)
		return
	}

	testWords = append(testWords, words...)

	fmt.Println("Starting benchmark...")
	fmt.Printf("Total words in dataset: %d\n", len(words))
	fmt.Printf("Total words in test set: %d\n", len(testWords))
	fmt.Printf("Number of runs for benchmarking: %d\n", *runs)


	bfHits, bfDuration := benchmarkLookups(bf, testWords, *runs)
	fmt.Printf("Bloom Filter: %d hits out of %d lookups in %v\n", bfHits, len(testWords)*(*runs), bfDuration)

	// Benchmark lookups in PostgreSQL
	pgHits, pgDuration, err := benchmarkPostgresLookups(database, testWords, *runs)
	if err != nil {
		fmt.Printf("Failed to benchmark PostgreSQL lookups: %v\n", err)
		return
	}
	fmt.Printf("PostgreSQL: %d hits out of %d lookups in %v\n", pgHits, len(testWords)*(*runs), pgDuration)
}
