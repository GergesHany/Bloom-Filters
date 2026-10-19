# Bloom-Filters

Bloom Filters: Theory, Engineering Trade-offs, and Implementation in Go.

This project implements a Bloom filter in Go and benchmarks its lookups
against equivalent lookups in PostgreSQL, using a shared generated dataset.

## Requirements

- Go 1.25+
- Docker (for the PostgreSQL comparison)

## Configuration

Database connection settings are read from environment variables (with
defaults matching `docker-compose.yml`):

| Variable      | Default            |
| ------------- | ------------------ |
| `DB_HOST`     | `localhost`        |
| `DB_PORT`     | `5433`             |
| `DB_USER`     | `myuser`           |
| `DB_PASSWORD` | `mysecretpassword` |
| `DB_NAME`     | `mydatabase`       |

Copy `.env.example` to `.env` and adjust as needed. `.env` is git-ignored.

## Commands

### 1. Start PostgreSQL

```bash
docker compose up -d
```

This starts PostgreSQL 17 on host port `5433` and runs
`docker-entrypoint-initdb.d/init.sql`, which creates the `words` table.

### 2. Generate datasets

The datasets are generated and git-ignored, so create them first:

```bash
go run generator/main.go -out data/words.txt 1000000
go run generator/main.go -out data/test_words.txt 1000000
```

### 3. Run the benchmark

```bash
go run main.go -n 200000 -FP 0.01 -words data/words.txt -wordsTest data/test_words.txt -runs 10
```

#### Flags

| Flag         | Default              | Description                                        |
| ------------ | -------------------- | -------------------------------------------------- |
| `-n`         | `1000`               | Expected number of elements                        |
| `-FP`        | `0.01`               | Target false positive rate                         |
| `-words`     | `data/words.txt`     | Dataset loaded into both the filter and PostgreSQL |
| `-wordsTest` | `data/test_words.txt`| Lookup test set (appended to the dataset)          |
| `-runs`      | `100`                | Benchmark repetitions over the test set            |

The dataset is read once and loaded into both the Bloom filter and the
`words` table. Lookups are then timed for each store and printed side by
side, letting you compare the Bloom filter's speed against the extra false
positives it reports relative to PostgreSQL.
