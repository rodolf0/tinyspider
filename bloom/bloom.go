package bloom

import (
	"math"
	"math/rand"
)

type Bloom struct {
	bitarray []uint8
	size     uint64
	keys     [][16]byte
}

func (b *Bloom) set(bit uint64) {
	b.bitarray[bit/8] |= 0x80 >> (bit % 8)
}

func (b *Bloom) get(bit uint64) bool {
	return b.bitarray[bit/8]&(0x80>>(bit%8)) != 0
}

/*
http://en.wikipedia.org/wiki/Bloom_filter
n is an estimate of the number of elements being filtered
p is the false positive probability
*/
func newbloom(n uint64, p float64) (b Bloom) {
	m := math.Ceil(-float64(n) * math.Log(p) / math.Ln2 / math.Ln2)
	k := math.Ceil(m * math.Ln2 / float64(n))
	// fmt.Printf("MiB %v keys %v\n", m/8/1024/1024, k)
	arrsize := (uint64(m) + 8) / 8
	b.bitarray = make([]uint8, arrsize)
	b.size = uint64(m)
	b.keys = make([][16]byte, int(k))
	for i := 0; i < len(b.keys); i++ {
		for j := 0; j < len(b.keys[i]); j++ {
			b.keys[i][j] = byte(rand.Int31n(256))
		}
	}
	return
}

func New(size uint64) (b Bloom) {
	return newbloom(size, 0.005)
}

func (b *Bloom) Has(value string) bool {
	for i := 0; i < len(b.keys); i++ {
		hash := SipHash24([]byte(value), b.keys[i])
		if !b.get(hash % b.size) {
			return false
		}
	}
	return true
}

func (b *Bloom) Add(value string) {
	for i := 0; i < len(b.keys); i++ {
		hash := SipHash24([]byte(value), b.keys[i])
		b.set(hash % b.size)
	}
}
