# Bloom-Filters
 Bloom Filters: Theory, Engineering Trade‑offs, and Implementation in Go 

## Commands 
```markdown
- go run generator/main.go -out data/words.txt 1000000
- go run generator/main.go -out data/test_words.txt 1000000
- go run main.go -n 200000 -FP 0.01 -words data/words.txt -wordsTest data/test_words.txt -runs 10
```
