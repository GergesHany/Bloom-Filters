package BloomFilter

import (
	// "sync"
	"fmt"
	"hash/fnv"
	"math"
)

type BloomFilter struct {
	bits           []byte // Bits array to store the filter
	count          uint64 // Number of elements added to the filter
	size           uint64 // Size of the bits array
	hashFuncsCount uint64

	//	mu sync.Mutex // Mutex for thread-safe operations
}

func validateParameters(expectedElements uint64, falsePositiveRate float64) error {
	if expectedElements == 0 {
		return fmt.Errorf("expected elements must be greater than 0")
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return fmt.Errorf("false positive rate must be between 0 and 1 (exclusive)")
	}
	return nil
}

func New(expectedElements uint64, falsePositiveRate float64) *BloomFilter {
	if err := validateParameters(expectedElements, falsePositiveRate); err != nil {
		panic(err)
	}

	// Calculate the size of the bit array (m) and the number of hash functions (k)
	m := uint64(-1 * float64(expectedElements) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2))

	// round m up to the nearest multiple of 8 for byte alignment
	if m%8 != 0 {
		m += 8 - (m % 8)
	}

	// Calculate the number of hash functions (k)
	k := uint64(math.Ceil((float64(m) / float64(expectedElements)) * math.Ln2))

	k = max(k, 1) // Ensure at least one hash function

	return &BloomFilter{
		bits:           make([]byte, m/8),
		size:           m,
		hashFuncsCount: k,
	}
}


// Getters for the BloomFilter struct fields

func (bf *BloomFilter) Size() uint64 {
	return bf.size
}

func (bf *BloomFilter) HashFuncsCount() uint64 {
	return bf.hashFuncsCount
}

func (bf *BloomFilter) Count() uint64 {
	return bf.count
}

func (bf *BloomFilter) Add(element string) {
	// get the hash positions for the element
	positions := bf.getHashPositions(element)
	for _, pos := range positions {
		// set the bit at the position to 1
		bf.setBit(pos)
	}
	bf.count++
}

func (bf *BloomFilter) getHashPositions(element string) []uint64 {
	positions := make([]uint64, bf.hashFuncsCount)
	data := []byte(element)

	h1 := hash1(data)
	h2 := hash2(data) | 1 // odd, so the sequence visits distinct positions

	for i := uint64(0); i < bf.hashFuncsCount; i++ {
		positions[i] = (h1 + i*h2) % bf.size
	}
	return positions
}

func hash1(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}
	
func hash2(data []byte) uint64 {
	h := fnv.New64()
	h.Write(data)
	return h.Sum64()
}

func (bf *BloomFilter) setBit(position uint64) {
	byteIndex := position / 8
	bitIndex := position % 8

	if byteIndex >= uint64(len(bf.bits)) {
		panic(fmt.Sprintf("bit position %d out of range (filter size %d)", position, bf.size))
	}

	bf.bits[byteIndex] |= (1 << bitIndex)
} 

func (bf *BloomFilter) Check(element string) bool {
	positions := bf.getHashPositions(element)
	for _, pos := range positions {
		if !bf.isBitSet(pos) {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) isBitSet(position uint64) bool {
	byteIndex := position / 8
	bitIndex := position % 8

	if byteIndex >= uint64(len(bf.bits)) {
		return false
	}

	return (bf.bits[byteIndex] & (1 << bitIndex)) != 0
}

func (bf *BloomFilter) Contains(element string) bool {
	positions := bf.getHashPositions(element)
	for _, pos := range positions {
		if !bf.isBitSet(pos) {
			return false
		}
	}
	return true
}