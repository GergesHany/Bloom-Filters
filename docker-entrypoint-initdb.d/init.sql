-- Create the words table
CREATE TABLE IF NOT EXISTS words (
    id SERIAL PRIMARY KEY,
    word TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create a hash index on the word column (ideal for exact '=' matches)
CREATE INDEX IF NOT EXISTS idx_words_word_hash ON words USING hash (word);
