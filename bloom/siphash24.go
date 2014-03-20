package bloom

// https://131002.net/siphash/siphash24.c

func _u8to64(in []byte) uint64 {
	return uint64(in[0]) |
		uint64(in[1])<<8 |
		uint64(in[2])<<16 |
		uint64(in[3])<<24 |
		uint64(in[4])<<32 |
		uint64(in[5])<<40 |
		uint64(in[6])<<48 |
		uint64(in[7])<<56
}

func _u64to8(in uint64) (out [8]byte) {
	out[0] = uint8(in)
	out[1] = uint8(in >> 8)
	out[2] = uint8(in >> 16)
	out[3] = uint8(in >> 24)
	out[4] = uint8(in >> 32)
	out[5] = uint8(in >> 40)
	out[6] = uint8(in >> 48)
	out[7] = uint8(in >> 56)
	return
}

func _rotl(x uint64, b uint) uint64 {
	return (x << b) | (x >> (64 - b))
}

func _sipround(v0, v1, v2, v3 *uint64) {
	*v0 += *v1
	*v1 = _rotl(*v1, 13)
	*v1 ^= *v0
	*v0 = _rotl(*v0, 32)
	*v2 += *v3
	*v3 = _rotl(*v3, 16)
	*v3 ^= *v2
	*v0 += *v3
	*v3 = _rotl(*v3, 21)
	*v3 ^= *v0
	*v2 += *v1
	*v1 = _rotl(*v1, 17)
	*v1 ^= *v2
	*v2 = _rotl(*v2, 32)
}

func _SipHash24(in []byte, key [16]byte) [8]byte {
	return _u64to8(SipHash24(in, key))
}

// siphash-2-4
func SipHash24(in []byte, key [16]byte) uint64 {
	k0 := _u8to64(key[:8])
	k1 := _u8to64(key[8:])
	var (
		v0 uint64 = 0x736f6d6570736575 ^ k0
		v1 uint64 = 0x646f72616e646f6d ^ k1
		v2 uint64 = 0x6c7967656e657261 ^ k0
		v3 uint64 = 0x7465646279746573 ^ k1
		b  uint64 = uint64(len(in)) << 56
	)

	for len(in) >= 8 {
		m := _u8to64(in[:8])
		v3 ^= m
		_sipround(&v0, &v1, &v2, &v3)
		_sipround(&v0, &v1, &v2, &v3)
		v0 ^= m
		in = in[8:]
	}

	switch len(in) {
	case 7:
		b |= uint64(in[6]) << 48
		fallthrough
	case 6:
		b |= uint64(in[5]) << 40
		fallthrough
	case 5:
		b |= uint64(in[4]) << 32
		fallthrough
	case 4:
		b |= uint64(in[3]) << 24
		fallthrough
	case 3:
		b |= uint64(in[2]) << 16
		fallthrough
	case 2:
		b |= uint64(in[1]) << 8
		fallthrough
	case 1:
		b |= uint64(in[0])
	}

	v3 ^= b
	_sipround(&v0, &v1, &v2, &v3)
	_sipround(&v0, &v1, &v2, &v3)
	v0 ^= b
	v2 ^= 0xff
	_sipround(&v0, &v1, &v2, &v3)
	_sipround(&v0, &v1, &v2, &v3)
	_sipround(&v0, &v1, &v2, &v3)
	_sipround(&v0, &v1, &v2, &v3)
	return v0 ^ v1 ^ v2 ^ v3
}
