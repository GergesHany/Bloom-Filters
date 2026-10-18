package BloomFilter

import (
	"sync"
	"sync/atomic"
	"fmt"
	"hash/fnv"
	"math"
)

type BloomFilter struct {
	bits           []byte // Bits array to store the filter
	count          uint64 // Number of elements added to the filter
	size           uint64 // Size of the bits array
	hashFuncsCount uint64

	mu sync.Mutex // Mutex for thread-safe operations
}

type Stats struct {
	BitsTotal uint64 // Total number of bits in the filter
	BitsSet uint64 // Number of bits set to 1 in the filter
	ItemsCount uint64 // Number of items added to the filter
	Density float64 // Density of the filter (ratio of bits set to total bits) 
	ExpectedFPRate float64 // Expected false positive rate based on the number of items added and the size of the filter
	ActualFPRate float64 // Actual false positive rate based on the number of items added and the size of the filter

	HashFunctionsCount uint64 // Number of hash functions used in the filter
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
	bf.mu.Lock()
	defer bf.mu.Unlock()

	// get the hash positions for the element
	positions := bf.getHashPositions(element)
	for _, pos := range positions {
		// set the bit at the position to 1
		bf.setBit(pos)
	}

	// increment the count of elements added to the filter
	// NOTE: Use atomic operation to safely increment the count in a concurrent environment
	atomic.AddUint64(&bf.count, 1) 
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
	bf.mu.Lock()
	defer bf.mu.Unlock()

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
	return bf.Check(element)
}

func (bf *BloomFilter) countBitsSet() uint64 {
	count := uint64(0)

	for pos := range int(bf.size) {
		if bf.isBitSet(uint64(pos)) {
			count++
		}
	}

	return count
}

func (bf *BloomFilter) Stats() Stats {
	bitsSet := bf.countBitsSet()
	var expectedFPRate float64

	density := float64(bitsSet) / float64(bf.size)

	if bf.count > 0 {
		expectedFPRate = math.Pow(1-math.Exp(-float64(bf.hashFuncsCount)*float64(bf.count)/float64(bf.size)), float64(bf.hashFuncsCount))
	}

	actualFPRate := math.Pow(density, float64(bf.hashFuncsCount))

	return Stats{
		BitsTotal:          bf.size,
		BitsSet:            bitsSet,
		Density:            density,
		ItemsCount:         bf.count,
		ExpectedFPRate:     expectedFPRate,
		ActualFPRate:       actualFPRate,
		HashFunctionsCount: bf.hashFuncsCount,
	}

}